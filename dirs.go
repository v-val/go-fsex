package main

import (
	"log"
	"os"
	"path/filepath"
)

func IsDir(path string) (bool, error) {
	var r bool
	i, err := os.Stat(path)
	if err == nil {
		r = i.IsDir()
	}
	return r, err
}

func (f *fsex) GetSubDirs(path string) ([]string, error) {
	var r []string
	isDir, err := IsDir(path)
	if err == nil && isDir {
		var t = []string{path}
		var entries []os.DirEntry
		entries, err = os.ReadDir(path)
		if err == nil {
			for _, d := range entries {
				if d.IsDir() {
					var subdirs []string
					subdirs, err = f.GetSubDirs(filepath.Join(path, d.Name()))
					if err != nil {
						log.Printf("Error: %s", err)
						break
					}
					t = append(t, subdirs...)
				}
			}
			if err == nil {
				r = t
			}
		}
	}
	return r, err
}
