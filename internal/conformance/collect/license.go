package collect

import "strings"

// DetectLicense identifies common permissive licenses from their text. It
// returns an empty string for anything it cannot identify confidently,
// including licenses that are not on the allowlist. Unknown is safer than
// guessing.
func DetectLicense(text []byte) string {
	s := strings.ToLower(string(text))
	s = strings.Join(strings.Fields(s), " ")

	if strings.Contains(s, "incompatible with secondary licenses") {
		return ""
	}
	switch {
	case strings.Contains(s, "mozilla public license version 2.0"),
		strings.Contains(s, "mozilla public license, version 2.0"):
		return "MPL-2.0"
	case strings.Contains(s, "apache license") && strings.Contains(s, "version 2.0"):
		return "Apache-2.0"
	case strings.Contains(s, "apache-2.0"):
		return "Apache-2.0"
	case strings.Contains(s, "mit license"),
		strings.Contains(s, "permission is hereby granted, free of charge, to any person obtaining a copy"):
		return "MIT"
	case strings.Contains(s, "redistribution and use in source and binary forms") && strings.Contains(s, "neither the name"):
		return "BSD-3-Clause"
	case strings.Contains(s, "redistribution and use in source and binary forms"):
		return "BSD-2-Clause"
	case strings.Contains(s, "isc license"),
		strings.Contains(s, "permission to use, copy, modify, and/or distribute this software for any purpose"):
		return "ISC"
	case strings.Contains(s, "this is free and unencumbered software released into the public domain"):
		return "Unlicense"
	case strings.Contains(s, "cc0 1.0 universal"),
		strings.Contains(s, "creative commons zero v1.0 universal"):
		return "CC0-1.0"
	}
	return ""
}
