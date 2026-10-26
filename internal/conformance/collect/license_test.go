package collect

import "testing"

func TestDetectLicense(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"apache", "Apache License\nVersion 2.0, January 2004\nhttp://www.apache.org/licenses/LICENSE-2.0", "Apache-2.0"},
		{"mit", "MIT License\n\nPermission is hereby granted, free of charge\nThe above copyright notice and this permission notice shall be included", "MIT"},
		{"mit0", "MIT No Attribution\n\nPermission is hereby granted, free of charge", "MIT-0"},
		{"bsd4", "Redistribution and use in source and binary forms...\nAll advertising materials mentioning features or use of this software must display", ""},
		{"commons-clause", "Apache License\nVersion 2.0\n\nCommons Clause", ""},
		{"json", "The Software shall be used for Good, not Evil.", ""},
		{"bsd3", "Redistribution and use in source and binary forms...\nNeither the name of the copyright holder", "BSD-3-Clause"},
		{"bsd2", "Redistribution and use in source and binary forms, with or without modification", "BSD-2-Clause"},
		{"isc", "ISC License\nPermission to use, copy, modify, and/or distribute this software for any purpose", "ISC"},
		{"mpl", "Mozilla Public License Version 2.0", "MPL-2.0"},
		{"mpl-incompatible", "Mozilla Public License Version 2.0\nIncompatible with Secondary Licenses", ""},
		{"unlicense", "This is free and unencumbered software released into the public domain.", "Unlicense"},
		{"cc0", "CC0 1.0 Universal", "CC0-1.0"},
		{"gpl", "GNU GENERAL PUBLIC LICENSE Version 3", ""},
		{"unknown", "All rights reserved.", ""},
		{"bare spdx mit", "MIT", "MIT"},
		{"bare spdx apache", "Apache-2.0", "Apache-2.0"},
		{"bare spdx bsd3", "BSD-3-Clause", "BSD-3-Clause"},
		{"bare spdx not a license", "Proprietary", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectLicense([]byte(tc.text)); got != tc.want {
				t.Errorf("DetectLicense() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectLicenseRejectsMixedCopyleft(t *testing.T) {
	text := "GNU GENERAL PUBLIC LICENSE Version 3\n\n" +
		"Permission is hereby granted, free of charge, to any person obtaining a copy.\n" +
		"Licensed under the Apache License, Version 2.0."
	if got := DetectLicense([]byte(text)); got != "" {
		t.Errorf("copyleft text classified as %q", got)
	}
}

func TestDetectLicenseIgnoresBareMention(t *testing.T) {
	text := "All rights reserved. This project is not released under the MIT License."
	if got := DetectLicense([]byte(text)); got != "" {
		t.Errorf("bare mention classified as %q", got)
	}
}
