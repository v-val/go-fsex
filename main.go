// https://github.com/fsnotify/fsnotify
package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/v-val/go-fsex/build-vars"
	"os"
	"path/filepath"
	"time"
)

const FirstPublicationYear = "2022"

// Init - get config parameters from env
func init() {
	//const ENV_VERBOSITY = "FSEX_VERBOSITY"
	//s := os.Getenv(ENV_VERBOSITY)
	//if s != "" {
	//
	//}
}

// Main
func main() {
	// Flag instructing to exit after detecting first changes
	runOnce := false
	// Flag instructing to clear the screen before executing command
	needClearScreenOnChanges := false
	// Disable watching subdirectories
	flagEnabledRecursiveWatch := true
	// Print version and exit
	flagPrintVersionAndExit := false
	// Print about and exit
	flagPrintAboutAndExit := false
	// Diagnostics quietness level
	flagSuppressDiagnostics := false
	// Hide command stdout
	flagSuppressStdout := false
	// Hide command stderr
	flagSuppressStderr := false
	// Get list of filesystem entities to watch from CLI
	var fsEntities stringListFlag
	// Patterns of pathnames to ignore
	var ignorePatterns stringListFlag
	// Files with patterns of pathnames to ignore (one per line)
	var ignoreFiles stringListFlag
	// TODO: verify uniqueness
	flag.Var(&fsEntities, "f", "File or dir to watch after")
	flag.BoolVar(&needClearScreenOnChanges, "c", needClearScreenOnChanges, "Clear screen before running command")
	flag.BoolVar(&runOnce, "1", runOnce, "Exit after executing command once")
	flag.BoolVar(&flagSuppressDiagnostics, "q", flagSuppressDiagnostics, "Suppress diagnostics")
	flag.BoolVar(&flagSuppressStdout, "O", flagSuppressStdout, "Hide command STDOUT")
	flag.BoolVar(&flagSuppressStderr, "E", flagSuppressStderr, "Hide command STDERR")
	flag.BoolVar(&flagPrintVersionAndExit, "version", flagPrintVersionAndExit, "Print version and exit")
	flag.BoolVar(&flagPrintAboutAndExit, "about", flagPrintAboutAndExit, "Print about info and exit")
	flag.Var(&ignorePatterns, "x", "Pattern to ignore.")
	flag.Var(&ignoreFiles, "X", "Files with patterns to ignore.\n"+
		`Please note: ".zzup.ignore" and ".rsync.ignore" auto-included`)
	flag.Parse()
	if flagSuppressDiagnostics {
		SetQuietness(incrementableInt(1))
	}
	if flagPrintVersionAndExit {
		fmt.Println(build_vars.GitRef)
		return
	}
	if flagPrintAboutAndExit {
		years := FirstPublicationYear
		currentYear := time.Now().Format("2006")
		if currentYear > years {
			years = fmt.Sprintf("%s-%s", years, currentYear)
		}
		fmt.Printf("%s version %s © %s %s\n", build_vars.AppName, build_vars.Version, years, build_vars.HomePage)
		return
	}
	//Printf("XXX Run once: %v", runOnce)
	// Check that at least one FS entity and at least one word command are passed
	if len(fsEntities) < 1 || len(flag.Args()) < 1 {
		Fatalf("Usage: fsex [options] -f<path> <command>")
	}
	//
	// Build ignore filters
	filter := NewFileFilter(ignorePatterns...)
	for _, d := range fsEntities {
		if isDir, _ := IsDir(d); isDir {
			for _, i := range []string{".zzup.ignore", ".rsync.ignore"} {
				i = filepath.Join(d, i)
				_, err := os.Stat(i)
				if err == nil {
					ignoreFiles = append(ignoreFiles, i)
				}
			}
		}
	}
	for _, f := range ignoreFiles {
		if err := filter.LoadFile(f); err != nil {
			Fatalf(`Error parsing ignore file "%s": %s`, f, err)
		}
		Debugf(`ignore file "%s" loaded`, f)
	}
	//
	Printf("Dir %v", fsEntities)
	// Check that watch and ignore lists do not intersect
	// TODO: use not straightforward matching, but:
	// (A) when pattern contains path separator, it's applied to entire pathname
	// (B) when pattern ends with path separator, it's applied to dirs only
	// (C) otherwise pattern applied to name only. NB: now we support only this case.
	// TODO: work through various ways to specify
	// * path: absolute, all relatives
	// * pattern: absolute, relative, all possible meanings
	// For now we assume that pattern is to match only name
	for _, f := range fsEntities {
		if filter.Match(f) {
			Fatalf(`"%s" is to both watch and ignore`, f)
		}
	}

	// Remaining CLi args treated as command
	cmd := flag.Args()
	Printf("Cmd %v", cmd)

	// End of CLI args parsing

	app := fsex{
		cmd:                      cmd,
		flagClearScreenOnChanges: needClearScreenOnChanges,
		flagSuppressStdout:       flagSuppressStdout,
		flagSuppressStderr:       flagSuppressStderr,
	}

	// Create FS watcher
	var watcher *fsnotify.Watcher
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Fatal(err)
	}
	defer func() {
		err = watcher.Close()
		if err != nil {
			Fatal(err)
		}
	}()

	// Pass FS entities to watcher
	for _, f := range fsEntities {
		// Top level entities already checked vs. ignore patterns
		err = watcher.Add(f)
		if err != nil {
			Fatal(err)
		}
		if flagEnabledRecursiveWatch {
			var dirs []string
			dirs, err = app.GetSubDirs(f)
			if err != nil {
				Fatalf(`Fail to recurse to "%s": %s`, f, err)
			}
			// list of subdirs is empty for non-directories
			for _, d := range dirs {
				if filter.Match(d) {
					Debugf(`"%s" ignored`, d)
				} else {
					err = watcher.Add(d)
					if err != nil {
						Fatal(err)
					}
				}
			}
		}
	}

	// Number of events in last detected bunch
	nevents := 0
	// Number of events in last detected bunch
	nerrors := 0
	// Number of idle loops since last detected event
	nidle := 0
	flagKeepRunning := true
	for flagKeepRunning {
		select {
		case event, ok := <-watcher.Events:
			if ok {
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					if filter.Match(event.Name) {
						Debugf(`"%s" ignored`, event.Name)
					} else {
						Tracef(`Got %s`, event)
						nevents++
						//Printf("E%06d %v", nevents, event)
						Debugf("E%06d", nevents)
						// TODO: delete for deleted dirs
						if flagEnabledRecursiveWatch && event.Op&fsnotify.Create != 0 {
							// Temp files can disappear faster than we check, so ignore errors
							if ok, err = IsDir(event.Name); err == nil && ok {
								err = watcher.Add(event.Name)
								if err != nil {
									Panic(err)
								}
							}
						}
						nidle = 0
					}
				}
			} else {
				Print("Events chan closed, finishing..")
				flagKeepRunning = false
			}
		case err, ok := <-watcher.Errors:
			if ok {
				Printf(`Watch error: %v`, err)
				nerrors++
				nidle = 0
			} else {
				Print("Errors chan closed, finishing..")
				flagKeepRunning = false
			}

		default:
			// Can be within short pause between two events / errors
			// Do number of short sleeps until total sleep time exceeds timeout
			if nidle < 10 {
				// + 10 x 10us sleeps
				// = 100us idle
				time.Sleep(10 * time.Microsecond)
			} else if nidle < 19 {
				// 100us idle
				// + 9 x 100us sleeps
				// = 1ms idle
				time.Sleep(100 * time.Microsecond)
			} else if nidle < 28 {
				// 1ms idle
				// + 9 x 1ms sleeps
				// = 10ms idle
				time.Sleep(time.Millisecond)
			} else if nidle < 37 {
				// 10ms idle
				// + 9 x 10ms sleeps
				// = 100ms idle
				time.Sleep(10 * time.Millisecond)
			} else if nidle < 45 {
				// 100ms idle
				// + 8 * 50ms sleeps
				// = 500ms idle
				time.Sleep(50 * time.Millisecond)
			} else {
				if nidle == 45 && (nevents > 0 || nerrors > 0) {
					app.execCommand()
					nevents = 0
					nerrors = 0
					if runOnce {
						flagKeepRunning = false
					}
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
			nidle++
		}
	}
	// When counters are non-zero
	// it means that main loop was interrupted without flush
	if nevents > 0 || nerrors > 0 {
		app.execCommand()
	}
}
