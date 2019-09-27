package filefinder

import (
	"fmt"
	"testing"
)

func equalLists(a, b []string) error {
	if len(a) != len(b) {
		return fmt.Errorf("different lengths: %d != %d", len(a), len(b))
	}
	for i, x := range a {
		if x != b[i] {
			return fmt.Errorf("element #%d is different: %q != %q", i, x, b[i])
		}
	}
	return nil
}

func TestFind(t *testing.T) {
	tests := []struct {
		desc   string
		paths  []string
		files  []string
		ok     bool
		want   []string
	}{
		{
			desc: "empty paths",
			paths: nil,
			files: nil,
			ok:    true,
			want:  nil,
		},
		{
			desc: "find an absolute file",
			paths: []string{"/etc/"},
			files: []string{"./passwd"},
			ok:    true,
			want:  []string{"/etc/passwd"},
		},
		{
			desc: "fail if any not found",
			paths: []string{"/etc"},
			files: []string{"passwd", "surely-does-not-exist-in-etc"},
			ok:    false,
		},
		{
			desc: "fail if files are absolute",
			paths: []string{"/"},
			files: []string{"/etc/passwd"},
			ok:    false,
		},
		{
			desc: "ok if found not in some later path",
			paths: []string{"/", "./some-unknown-dir", "filefinder/testdata", "testdata"},
			files: []string{"testfile.txt"},
			ok:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			f := New(tc.paths...)
			got, err := f.Find(tc.files...)
			if tc.ok != (err == nil) {
				t.Errorf("got %v, want %t", err, tc.ok)
			}
			if err != nil || len(tc.want) == 0 {
				return
			}
			if err := equalLists(got, tc.want); err != nil {
				t.Errorf("got %v, want %v => %v", got, tc.want, err)
			}
		})
	}
}
