package collect

import "strings"

// copyleftMarkers reject a license text even when it also mentions a
// permissive license. A dual-license notice that offers a permissive option
// usually names the permissive license first; the conservative choice is to
// reject rather than select a license from ambiguous text.
var copyleftMarkers = []string{
	"gnu general public license",
	"gnu lesser general public license",
	"gnu affero general public license",
	"gpl-2.0",
	"gpl-3.0",
	"gplv2",
	"gplv3",
	"lgpl-2",
	"lgpl-3",
	"agpl-3",
	"european union public licence",
	"creative commons attribution-sharealike",
	"cc by-sa",
	"server side public license",
	"business source license",
	"sspl",
	// Texts that quote a permissive license but add restrictions are not that
	// license, and several are incompatible with GPL-3.0.
	"the software shall be used for good, not evil",
	"commons clause",
	"hippocratic license",
	"all advertising materials mentioning features or use of this software",
	"all advertising materials mentioning features or use of the software",
}

// DetectLicense identifies a permissive license from its full text. It returns
// an empty string when the text is unknown, copyleft, or merely mentions a
// permissive license without being one.
func DetectLicense(text []byte) string {
	s := strings.ToLower(string(text))
	s = strings.Join(strings.Fields(s), " ")

	for _, marker := range copyleftMarkers {
		if strings.Contains(s, marker) {
			return ""
		}
	}

	// A LICENSE that contains only an SPDX identifier is a valid license file.
	switch strings.TrimSpace(s) {
	case "apache-2.0":
		return "Apache-2.0"
	case "mit":
		return "MIT"
	case "mit-0":
		return "MIT-0"
	case "bsd-2-clause":
		return "BSD-2-Clause"
	case "bsd-3-clause":
		return "BSD-3-Clause"
	case "isc":
		return "ISC"
	case "mpl-2.0":
		return "MPL-2.0"
	case "unlicense":
		return "Unlicense"
	case "cc0-1.0":
		return "CC0-1.0"
	}

	switch {
	case strings.Contains(s, "mozilla public license version 2.0"),
		strings.Contains(s, "mozilla public license, version 2.0"):
		if strings.Contains(s, "incompatible with secondary licenses") {
			return ""
		}
		return "MPL-2.0"
	case strings.Contains(s, "apache license") &&
		strings.Contains(s, "version 2.0") &&
		(strings.Contains(s, "apache.org/licenses") ||
			strings.Contains(s, "terms and conditions for use, reproduction, and distribution")):
		return "Apache-2.0"
	case strings.TrimSpace(s) == "apache-2.0":
		return "Apache-2.0"
	case strings.Contains(s, "mit no attribution"), strings.Contains(s, "mit-0"):
		return "MIT-0"
	case strings.Contains(s, "permission is hereby granted, free of charge") &&
		strings.Contains(s, "the above copyright notice and this permission notice shall be included"):
		return "MIT"
	case strings.Contains(s, "redistribution and use in source and binary forms") &&
		strings.Contains(s, "neither the name"):
		return "BSD-3-Clause"
	case strings.Contains(s, "redistribution and use in source and binary forms"):
		return "BSD-2-Clause"
	case strings.Contains(s, "permission to use, copy, modify, and/or distribute this software"):
		return "ISC"
	case strings.Contains(s, "this is free and unencumbered software released into the public domain"):
		return "Unlicense"
	case strings.Contains(s, "cc0 1.0 universal"),
		strings.Contains(s, "creative commons zero v1.0 universal"):
		return "CC0-1.0"
	}
	return ""
}
