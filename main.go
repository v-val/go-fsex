// https://github.com/fsnotify/fsnotify
package main

import (
	"errors"
	"flag"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Type to store list of strings passed with repeating CLI flag
type stringListFlag []string

// Converter to string reqd to use type with flag package
func (s *stringListFlag) String() string {
	r := ""
	for _, v := range *s {
		if len(r) > 0 {
			r += ", "
		}
		r += v
	}
	return r
}

// Setter reqd to use type with flag package
func (d *stringListFlag) Set(value string) error {
	*d = append(*d, value)
	return nil
}

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
	// Get list of filesystem entities to watch from CLI
	var fsEntities stringListFlag
	flag.Var(&fsEntities, "f", "File or dir to watch after")
	flag.BoolVar(&needClearScreenOnChanges, "c", needClearScreenOnChanges, "Clear screen before running command")
	flag.BoolVar(&runOnce, "1", runOnce, "Exit on first event")
	flag.Parse()
	//log.Printf("XXX Run once: %v", runOnce)
	log.Printf(`XXX clear: %v`, needClearScreenOnChanges)
	// Check that at least one FS entity and at least one word command are passed
	if len(fsEntities) < 1 || len(flag.Args()) < 1 {
		log.Fatalf("Usage: fsex -f<path> [-f<path2> ...] <command>")
	}
	log.Printf("Dir %v", fsEntities)

	// Remaining CLi args treated as command
	cmd := flag.Args()
	log.Printf("Cmd %v", cmd)

	// Create FS watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Channel to notify about detected FS events
	notifications := make(chan bool)
	// Channel to pass halt instruction
	done := make(chan bool)

	// Coro that gets and filters events and passes notification to main thread
	go func() {
		//lastEventTime := time.Now()
		for {
			select {
			case event, ok := <-watcher.Events:
				//log.Printf("XXX event: %v", event)
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					//t := time.Now()
					//if t.Sub(lastEventTime) < 100 * time.Millisecond {}
					//log.Println("modified file:", event.Name, event.Op)
					notifications <- true
					if runOnce {
						done <- true
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Pass FS entities to watcher
	for _, f := range fsEntities {
		//log.Printf("XXX Add %v", f)
		err = watcher.Add(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO: make constants
	waitDuration := 100 * time.Millisecond
	actionAfter := time.Second
	idleLoopsMax := int(actionAfter / waitDuration)
	//log.Printf("Idle Loops Max %v", idleLoopsMax)

	// Number of events in last detected bunch
	nevents := 0
	// Number of idle loops since last detected event
	nidle := 0
	for {
		select {
		case _ = <-done:
			// Got halt command
			return
		case _ = <-notifications:
			// Event detected, do nothing
			nevents += 1
			nidle = 0
		default:
			// When no event detected we have two options
			// * Execute command once actionAfter since last event
			// * idle otherwise
			if nevents == 0 || nidle < idleLoopsMax {
				time.Sleep(waitDuration)
				if nevents > 0 {
					nidle += 1
				}
			} else {
				// Print header, execute command
				headOpen := strings.Repeat("=", 48)
				headClose := strings.Repeat("-", 48)
				bodyEndError := strings.Repeat("+", 48)
				bodyEndOk := strings.Repeat(".", 48)
				if nevents > 0 && nidle >= idleLoopsMax {
					if needClearScreenOnChanges {
						//log.Println("Clear the screen..")
						//screen.Clear()
						//screen.MoveTopLeft()
					}
					println(headOpen)
					log.Printf("RUN %v", cmd)
					println(headClose)
					var cmd_ *exec.Cmd
					if len(cmd) == 1 {
						cmd_ = exec.Command(cmd[0])
					} else {
						cmd_ = exec.Command(cmd[0], cmd[1:]...)
					}
					cmd_.Stdout = os.Stdout
					cmd_.Stderr = os.Stderr
					err := cmd_.Run()
					if err != nil {
						println(bodyEndError)
						log.Printf("Command failed: %s", err)
					} else {
						var ee *exec.ExitError
						println(bodyEndOk)
						if errors.As(err, &ee) {
							log.Printf("Command returned %d", ee.ExitCode())
						} else {
							log.Printf("Command completed successfully")
						}
					}
					nevents = 0
					nidle = 0
				}
			}
		}
	}
}
