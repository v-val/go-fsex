package main

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

// Line-based conf format looks most appropriate for the moment.
// Blank character in line opening and ending are ignored
// Hash `#` and subsequent characters through the end of the line are ignored.
// Empty lines ignored.
// Remaining data lines treated as follows:
// * if the first symbol is dash, line treated as configuration directive
// * otherwise it's takes as a path to watch
// TODO: expansion of shell patterns, e.g. "~"
// TODO: ? expansion of environment variables
var lineBasedConfComment = regexp.MustCompile(`\s*#.*`)

func LoadConfData(reader *bytes.Reader) (targets []string, err error) {
	if reader == nil {
		err = errors.New("nil reader")
		return
	}
	for scanner := bufio.NewScanner(reader); scanner.Scan(); {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			line = lineBasedConfComment.ReplaceAllString(line, "")
		}
		if line != "" {
			if line[0] == '-' {
				Trace(`Conf "%s"`, line)
			} else {
				Trace(`Path "%s"`, line)
				targets = append(targets, line)
			}
		}
	}
	return
}
