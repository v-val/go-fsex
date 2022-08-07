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

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatalf("Usage: onfse <path> <command>")
	}
	dir := flag.Arg(0)
	log.Printf("Dir %v", dir)

	cmd := flag.Args()[1:]
	log.Printf("Cmd %v", cmd)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	notifications := make(chan bool)
	done := make(chan bool)
	go func() {
		//lastEventTime := time.Now()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//log.Println("event:", event)
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					//t := time.Now()
					//if t.Sub(lastEventTime) < 100 * time.Millisecond {}
					//log.Println("modified file:", event.Name, event.Op)
					notifications <- true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	waitDuration := 100 * time.Millisecond
	actionAfter := time.Second / 2
	idleLoopsMax := int(actionAfter / waitDuration)
	//log.Printf("Idle Loops Max %v", idleLoopsMax)

	nevents := 0
	nidle := 0
	for {
		select {
		case _ = <-done:
			return
		case _ = <-notifications:
			nevents += 1
			nidle = 0
		default:
			if nevents == 0 || nidle < idleLoopsMax {
				time.Sleep(waitDuration)
				if nevents > 0 {
					nidle += 1
				}
			} else {
				headOpen := strings.Repeat("=", 48)
				headClose := strings.Repeat("-", 48)
				bodyEndError := strings.Repeat("+", 48)
				bodyEndOk := strings.Repeat(".", 48)
				if nevents > 0 && nidle >= idleLoopsMax {
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
