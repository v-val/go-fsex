package main

import (
	"github.com/sabhiram/go-gitignore"
)

// FileFilter is an interface for file matching
type FileFilter interface {
	// LoadFile loads patterns from file
	LoadFile(ignoreFile string) error
	// Match tests pathname
	Match(pathname string) bool
}

// fileFilterImpl implements FileFilter interface
type fileFilterImpl struct {
	matchers []*ignore.GitIgnore
}

// Match tests pathname
func (filter *fileFilterImpl) Match(pathname string) bool {
	for _, m := range filter.matchers {
		if m.MatchesPath(pathname) {
			return true
		}
	}
	return false
}

// LoadFile loads patterns from file
func (filter *fileFilterImpl) LoadFile(f string) error {
	i, err := ignore.CompileIgnoreFile(f)
	if err == nil {
		filter.matchers = append(filter.matchers, i)
	}
	return err
}

// NewFileFilter returns FileFilter built from patterns passed in arguments
func NewFileFilter(patterns ...string) FileFilter {
	var m []*ignore.GitIgnore
	if len(patterns) > 0 {
		m = append(m, ignore.CompileIgnoreLines(patterns...))
	}
	return &fileFilterImpl{
		matchers: m,
	}
}
