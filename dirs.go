package main

import (
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

func (app *fsex) GetSubDirs(path string, filter FileFilter) ([]string, error) {
	var r []string
	isDir, err := IsDir(path)
	if err == nil && isDir {
		var t = []string{path}
		var entries []os.DirEntry
		entries, err = os.ReadDir(path)
		if err == nil {
			for _, d := range entries {
				if d.IsDir() {
					p := filepath.Join(path, d.Name())
					if filter == nil || !filter.Match(p) {
						var subdirs []string
						subdirs, err = app.GetSubDirs(p, filter)
						if err != nil {
							Printf("Error: %s", err)
							break
						}
						t = append(t, subdirs...)
					} else {
						Debugf(`ignore "%s"`, p)
					}
				}
			}
			if err == nil {
				r = t
			}
		}
	}
	return r, err
}
