package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_0007_FsexApp_FileFilter(t *testing.T) {
	{
		const Pattern = "/.zyx"
		f := NewFileFilter(Pattern)
		assert.True(t, f.Match(Pattern))
		assert.True(t, f.Match(".zyx"))
		assert.False(t, f.Match(".zyxX"))
		// I think this one supposed to go
		assert.True(t, f.Match(".zyx/a"))
		assert.True(t, f.Match("/.zyx"))
		assert.False(t, f.Match("one/.zyx"))
	}
	{
		const Pattern = ".zy*"
		f := NewFileFilter(Pattern)
		assert.True(t, f.Match(Pattern))
		assert.True(t, f.Match(".zy"))
		assert.True(t, f.Match(".zyN"))
		assert.True(t, f.Match(".zyxXyZ"))
		assert.True(t, f.Match(".zyx/a"))
		assert.True(t, f.Match(".zyx/a"))
	}
	{
		const Pattern = "/.zy*"
		f := NewFileFilter(Pattern)
		assert.True(t, f.Match(Pattern))
		assert.True(t, f.Match(".zy"))
		assert.True(t, f.Match(".zyN"))
		assert.True(t, f.Match(".zyxXyZ"))
		assert.True(t, f.Match(".zyx/a"))
	}
}
