package collect

import "testing"

func TestDetectLicense(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"apache", "Apache License\nVersion 2.0, January 2004", "Apache-2.0"},
		{"mit", "MIT License\n\nPermission is hereby granted, free of charge", "MIT"},
		{"bsd3", "Redistribution and use in source and binary forms...\nNeither the name of the copyright holder", "BSD-3-Clause"},
		{"bsd2", "Redistribution and use in source and binary forms, with or without modification", "BSD-2-Clause"},
		{"isc", "ISC License\nPermission to use, copy, modify, and/or distribute", "ISC"},
		{"mpl", "Mozilla Public License Version 2.0", "MPL-2.0"},
		{"mpl-incompatible", "Mozilla Public License Version 2.0\nIncompatible with Secondary Licenses", ""},
		{"unlicense", "This is free and unencumbered software released into the public domain.", "Unlicense"},
		{"cc0", "CC0 1.0 Universal", "CC0-1.0"},
		{"gpl", "GNU GENERAL PUBLIC LICENSE Version 3", ""},
		{"unknown", "All rights reserved.", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectLicense([]byte(tc.text)); got != tc.want {
				t.Errorf("DetectLicense() = %q, want %q", got, tc.want)
			}
		})
	}
}
