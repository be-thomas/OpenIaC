package collect

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"unicode/utf8"
)

// Archive holds the text files that matter to the corpus. It never contains
// binaries, state, plans or credentials files.
type Archive struct {
	SHA256    string
	Bytes     int64
	TextBytes int64
	Files     map[string][]byte // .tf and .tf.json, paths relative to repo root
	Licenses  map[string][]byte // root-level license candidates
	Notices   map[string][]byte // root-level notice candidates
	Truncated bool
}

var skippedSegments = map[string]bool{
	".git":             true,
	".github":          true,
	".terraform":       true,
	"node_modules":     true,
	"vendor":           true,
	"terragrunt-cache": true,
}

// ExtractArchive reads a GitHub tar.gz or zip archive and keeps only Terraform
// source and license files. The caller supplies the hard text cap so a
// decompression bomb cannot exhaust memory.
func ExtractArchive(data []byte, maxText, maxFile int64) (*Archive, error) {
	sum := sha256.Sum256(data)
	a := &Archive{
		SHA256:   hex.EncodeToString(sum[:]),
		Bytes:    int64(len(data)),
		Files:    make(map[string][]byte),
		Licenses: make(map[string][]byte),
		Notices:  make(map[string][]byte),
	}

	var raw map[string][]byte
	var err error
	switch {
	case len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b:
		raw, err = readTarGz(data, maxText, maxFile, a)
	case len(data) >= 4 && bytes.Equal(data[:4], []byte("PK\x03\x04")):
		raw, err = readZip(data, maxText, maxFile, a)
	default:
		return nil, fmt.Errorf("unsupported archive format")
	}
	if err != nil {
		return nil, err
	}
	raw = stripArchiveRoot(raw)
	for name, content := range raw {
		if !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
			continue
		}
		switch {
		case isTerraformFile(name):
			a.Files[name] = content
		case isRootLicense(name):
			a.Licenses[name] = content
		case isRootNotice(name):
			a.Notices[name] = content
		}
	}
	if a.Truncated {
		return nil, fmt.Errorf("archive text exceeds %d bytes", maxText)
	}
	if len(a.Files) == 0 {
		return nil, fmt.Errorf("archive contains no Terraform files")
	}
	return a, nil
}

func readTarGz(data []byte, maxText, maxFile int64, a *Archive) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	out := make(map[string][]byte)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			continue
		}
		if !interestingName(h.Name) {
			continue
		}
		content, err := readLimited(tr, maxFile)
		if err != nil {
			return nil, err
		}
		if content == nil {
			continue
		}
		if a.TextBytes+int64(len(content)) > maxText {
			a.Truncated = true
			return nil, nil
		}
		a.TextBytes += int64(len(content))
		out[cleanArchivePath(h.Name)] = content
	}
	return out, nil
}

func readZip(data []byte, maxText, maxFile int64, a *Archive) (map[string][]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	out := make(map[string][]byte)
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !f.Mode().IsRegular() {
			continue
		}
		if !interestingName(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := readLimited(rc, maxFile)
		rc.Close()
		if err != nil {
			return nil, err
		}
		if content == nil {
			continue
		}
		if a.TextBytes+int64(len(content)) > maxText {
			a.Truncated = true
			return nil, nil
		}
		a.TextBytes += int64(len(content))
		out[cleanArchivePath(f.Name)] = content
	}
	return out, nil
}

func readLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, nil
	}
	return data, nil
}

func cleanArchivePath(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Clean("/" + name)
	return strings.TrimPrefix(name, "/")
}

func interestingName(name string) bool {
	p := cleanArchivePath(name)
	if p == "." || strings.HasPrefix(p, "../") {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if skippedSegments[seg] || strings.HasPrefix(seg, ".") {
			return false
		}
	}
	return isTerraformFile(p) || isLicenseName(p) || isNoticeName(p)
}

func stripArchiveRoot(files map[string][]byte) map[string][]byte {
	if len(files) == 0 {
		return files
	}
	prefix := ""
	for name := range files {
		first := strings.SplitN(name, "/", 2)[0]
		if prefix == "" {
			prefix = first
		}
		if first != prefix {
			return files
		}
	}
	if prefix == "" {
		return files
	}
	prefix += "/"
	out := make(map[string][]byte, len(files))
	for name, content := range files {
		if !strings.HasPrefix(name, prefix) {
			return files
		}
		out[strings.TrimPrefix(name, prefix)] = content
	}
	return out
}

func isTerraformFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".tf") || strings.HasSuffix(lower, ".tf.json")
}

func isLicenseName(name string) bool {
	base := strings.ToLower(path.Base(name))
	switch base {
	case "license", "license.md", "license.txt", "license.rst", "copying", "copying.md", "copying.txt":
		return true
	}
	return false
}

func isNoticeName(name string) bool {
	base := strings.ToLower(path.Base(name))
	switch base {
	case "notice", "notice.md", "notice.txt":
		return true
	}
	return false
}

func isRootLicense(name string) bool { return !strings.Contains(name, "/") && isLicenseName(name) }
func isRootNotice(name string) bool  { return !strings.Contains(name, "/") && isNoticeName(name) }

// PrimaryLicense returns the preferred root license file, if any. The choice
// is deterministic: LICENSE first, then COPYING, then lexicographic order.
func (a *Archive) PrimaryLicense() (string, []byte, bool) {
	if len(a.Licenses) == 0 {
		return "", nil, false
	}
	names := make([]string, 0, len(a.Licenses))
	for name := range a.Licenses {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		rank := func(s string) int {
			low := strings.ToLower(s)
			switch {
			case low == "license":
				return 0
			case low == "license.md":
				return 1
			case low == "license.txt":
				return 2
			case low == "copying":
				return 3
			default:
				return 4
			}
		}
		ri, rj := rank(names[i]), rank(names[j])
		if ri != rj {
			return ri < rj
		}
		return names[i] < names[j]
	})
	name := names[0]
	return name, a.Licenses[name], true
}

// PrimaryNotice returns the preferred root notice file, if any.
func (a *Archive) PrimaryNotice() (string, []byte, bool) {
	if len(a.Notices) == 0 {
		return "", nil, false
	}
	names := make([]string, 0, len(a.Notices))
	for name := range a.Notices {
		names = append(names, name)
	}
	sort.Strings(names)
	return names[0], a.Notices[names[0]], true
}
