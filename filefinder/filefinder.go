// Package filefinder helps to find files in absolute or relative paths.
package filefinder

import (
	"fmt"
	"os"
	"path/filepath"
)

// Finder is a helper to find files in a slice of directories.
type Finder struct {
	paths []string
}

// New creates a new Finder.
func New(paths ...string) Finder {
	if len(paths) == 0 {
		paths = []string{"./"}
	}
	return Finder{paths}
}

// Find searches for files in a set of directories.
// All files should have relative names.
// If a file is found in some directory all other files
// should also exists in the same directory.
func (f Finder) Find(files ...string) ([]string, error) {
	var fps []string
	for _, path := range files {
		f := filepath.Clean(path)
		if filepath.IsAbs(f) {
			return nil, fmt.Errorf("file path %q is absolute", path)
		}
		fps = append(fps, f)
	}
	for _, path := range f.paths {
		path = filepath.Clean(path)
		fps, err := findAll(path, fps)
		if !os.IsNotExist(err) {
			return fps, err
		}
	}
	return nil, fmt.Errorf("files %v are not found in any of %v", files, f.paths)
}

func findAll(dir string, files []string) ([]string, error) {
	var fps []string
	for _, fp := range files {
		fp = filepath.Join(dir, fp)
		_, err := os.Stat(fp)
		if err != nil {
			return nil, err
		}
		fps = append(fps, fp)
	}
	return fps, nil
}

// Must returns the list of found files, or panics if they are not found.
// It is useful for program initialization.
func (f Finder) Must(files ...string) []string {
	files, e := f.Find(files...)
	if e != nil {
		panic(e)
	}
	return files
}
