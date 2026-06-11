package backend

import "testing"

func TestParseContainerVersion(t *testing.T) {
	tests := map[string]struct {
		ok                  bool
		major, minor, patch int
	}{
		"container CLI version 1.0.0 (build: release)": {true, 1, 0, 0},
		"0.9.0": {true, 0, 9, 0},
		"nope":  {false, 0, 0, 0},
	}
	for input, want := range tests {
		major, minor, patch, ok := parseContainerVersion(input)
		if ok != want.ok {
			t.Errorf("%q: ok=%v want %v", input, ok, want.ok)
			continue
		}
		if !ok {
			continue
		}
		if major != want.major || minor != want.minor || patch != want.patch {
			t.Errorf("%q: got %d.%d.%d want %d.%d.%d", input, major, minor, patch, want.major, want.minor, want.patch)
		}
	}
}
