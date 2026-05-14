package main

import (
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func Test_0000_MakeTestDir(t *testing.T) {
	const NTests = 5
	//for i := 0; i < NTests; i++ {
	//	if d, f, err := MakeTestDir(); err != nil {
	//		t.Fatal(`Can't create test dir: %s`, err)
	//	} else {
	//		t.Logf(`Test dir "%s" created`, d)
	//		defer f()
	//	}
	//}
	{
		var d string
		var err error
		var f func()
		{
			d, f, err = MakeTestDir()
			if err != nil {
				t.Fatal(err)
			}
			defer f()
			if _, err := os.Stat(d); err != nil {
				// Stat error is very informative
				t.Fatal(err)
			}
		}
		if _, err := os.Stat(d); err != nil {
			// Stat error is very informative
			if !os.IsNotExist(err) {
				t.Fatalf(`Unexpectedly "%s" exists: %s`, d, err)
			} else {
				t.Fatal(err)
			}
		}
	}
}

type fsEntryT uint8

const (
	FsEntryFile fsEntryT = iota
	FsEntrySymlink
	FsEntryDirectory
	FsEntryTypeCount
)

func getRandFsEntryType() fsEntryT {
	return fsEntryT(rand.Uint32() % uint32(FsEntryTypeCount))
}

const MaxSubSirs uint8 = 255

type dirSetNoneT_ struct{}

var dirSetNone_ dirSetNoneT_ = dirSetNoneT_{}

type dirSetT map[string]dirSetNoneT_

// makeTree creates hierarchy or dirs / symlinks / files
func makeTree(t *testing.T, allNodes *[]string, subdirs *dirSetT, path string, depth uint8, nleafs uint8) error {
	//t.Logf("makeTree([%v], [%v], %s, %d, %d)\n", len(*allNodes), len(*subdirs), dir, depth, nleafs)
	var err error
	for n := uint8(0); n < nleafs; n++ {
		p := filepath.Join(path, fmt.Sprintf("%02x-%02x-", depth, n))
		switch e := getRandFsEntryType(); e {
		case FsEntryFile:
			p += "F"
			_, err = os.Create(p)
			if err != nil {
				return err
			}
		case FsEntrySymlink:
			p += "H"
			var s string
			if len(*allNodes) == 0 {
				s = path
			} else {
				s = (*allNodes)[rand.Uint32()%uint32(len(*allNodes))]
			}
			err = os.Symlink(s, p)
			if err != nil {
				return err
			}
		case FsEntryDirectory:
			if depth > uint8(0) && len(*subdirs) < 256 {
				p += "D"
				err = os.Mkdir(p, fs.ModePerm)
				if err != nil {
					return err
				}
				(*subdirs)[p] = dirSetNone_
				err = makeTree(t, allNodes, subdirs, p, depth-1, nleafs)
				if err != nil {
					return fmt.Errorf(`can't makeTree(nodes=%v, subdirs=%v, dir="%s", depth=%d, leafs=%d): %v`,
						allNodes, subdirs, p, depth-1, nleafs, err)
				}
			}
		default:
			t.Fatalf(`Unexpected node type "%v"`, e)
		}
		*allNodes = append(*allNodes, p)
	}
	return err
}

func Test_0000_FsexApp_GetSubDirs(t *testing.T) {
	//
	const TreeDepth uint8 = 3
	const TreeLeafs uint8 = 11
	//
	testDir, cleanupFn, err := MakeTestDir()
	if err != nil {
		t.Fatalf(`fail to make test dir: %s`, err)
	}
	var nodes []string
	var dirsCreated dirSetT = make(dirSetT)
	err = makeTree(t, &nodes, &dirsCreated, testDir, TreeDepth, TreeLeafs)
	if err != nil {
		t.Fatalf(`fail to make tree: %s`, err)
	}
	//t.Logf("Tree created\n%d nodes\n%d subdirs:\n  %s", len(nodes), len(dirsCreated), strings.Join(dirsCreated, "\n  "))
	t.Logf("Tree created: %d nodes, %d subdirs", len(nodes), len(dirsCreated))

	app := fsex{}
	var dirsFound []string
	dirsFound, err = app.GetSubDirs(testDir, nil)
	if err != nil {
		t.Fatalf(`Fail to get subdirs of "%s"`, testDir)
	}
	if dirsFound[0] != testDir {
		t.Errorf(`First dir found "%s" is not top dir "%s"`, dirsFound[0], testDir)
	} else {
		dirsFound = dirsFound[1:]
	}
	t.Logf("Subdirs found: %d", len(dirsFound))
	for _, d := range dirsFound {
		_, ok := dirsCreated[d]
		if !ok {
			t.Errorf(`Unknown subdir "%s" found`, d)
		} else {
			delete(dirsCreated, d)
		}
	}
	if len(dirsCreated) > 0 {
		t.Fatalf(`%d dirs weren't found: %v`, len(dirsCreated), dirsCreated)
	}
	defer cleanupFn()
}
