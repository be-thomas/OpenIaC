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
		case isLicenseName(name):
			a.Licenses[name] = content
		case isNoticeName(name):
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
	// Skipped entries still have to be decompressed, so bound the total
	// expanded size rather than trusting the compressed archive size.
	limit := maxText * 8
	if limit <= 0 {
		limit = maxText
	}
	tr := tar.NewReader(&cappedReader{r: gz, remaining: limit})
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
		content, tooLarge, err := readLimited(tr, maxFile)
		if err != nil {
			return nil, err
		}
		if tooLarge {
			return nil, fmt.Errorf("archive file %s exceeds %d bytes", h.Name, maxFile)
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
		if int64(f.UncompressedSize64) > maxFile {
			return nil, fmt.Errorf("archive file %s exceeds %d bytes", f.Name, maxFile)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, tooLarge, err := readLimited(rc, maxFile)
		rc.Close()
		if err != nil {
			return nil, err
		}
		if tooLarge {
			return nil, fmt.Errorf("archive file %s exceeds %d bytes", f.Name, maxFile)
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

// cappedReader bounds how many decompressed bytes a reader may consume.
type cappedReader struct {
	r         io.Reader
	remaining int64
}

func (c *cappedReader) Read(p []byte) (int, error) {
	if c.remaining <= 0 {
		return 0, fmt.Errorf("archive expands beyond the text budget")
	}
	if int64(len(p)) > c.remaining {
		p = p[:c.remaining]
	}
	n, err := c.r.Read(p)
	c.remaining -= int64(n)
	return n, err
}

func readLimited(r io.Reader, max int64) ([]byte, bool, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > max {
		return nil, true, nil
	}
	return data, false, nil
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
	case "license", "license.md", "license.txt", "license.rst",
		"licence", "licence.md", "licence.txt", "licence.rst",
		"copying", "copying.md", "copying.txt":
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

// NearestLicense returns the license file that governs a candidate directory:
// its own LICENSE first, then the nearest ancestor up to the repository root.
func (a *Archive) NearestLicense(dir string) (string, []byte, bool) {
	return nearestNamed(a.Licenses, dir)
}

// NearestNotice returns the nearest ancestor NOTICE file, if any.
func (a *Archive) NearestNotice(dir string) (string, []byte, bool) {
	return nearestNamed(a.Notices, dir)
}

func nearestNamed(files map[string][]byte, dir string) (string, []byte, bool) {
	if len(files) == 0 {
		return "", nil, false
	}
	for {
		if name, content, ok := preferredInDir(files, dir); ok {
			return name, content, true
		}
		if dir == "" {
			return "", nil, false
		}
		dir = path.Dir(dir)
		if dir == "." {
			dir = ""
		}
	}
}

func preferredInDir(files map[string][]byte, dir string) (string, []byte, bool) {
	var names []string
	for name := range files {
		if path.Dir(name) != dir && !(dir == "" && !strings.Contains(name, "/")) {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "", nil, false
	}
	sort.Slice(names, func(i, j int) bool {
		rank := func(s string) int {
			switch strings.ToLower(path.Base(s)) {
			case "license":
				return 0
			case "license.md":
				return 1
			case "license.txt":
				return 2
			case "copying":
				return 3
			case "notice":
				return 0
			case "notice.md":
				return 1
			case "notice.txt":
				return 2
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
	return name, files[name], true
}

// PrimaryLicense returns the preferred root license file, if any.
func (a *Archive) PrimaryLicense() (string, []byte, bool) {
	return a.NearestLicense("")
}

// PrimaryNotice returns the preferred root notice file, if any.
func (a *Archive) PrimaryNotice() (string, []byte, bool) {
	return a.NearestNotice("")
}
