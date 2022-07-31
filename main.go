// https://github.com/fsnotify/fsnotify
package main

import (
	"log"
	"flag"
	"github.com/fsnotify/fsnotify"
//    "os"
)

func main() {
    flag.Parse()
    if len(flag.Args()) == 0 {
	    log.Fatalf("Usage: onfse <path>")
    }
    dir := flag.Arg(0)
    log.Printf("Dir %v", dir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
	 	log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op & (fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename ) != 0 {
					log.Println("modified file:", event.Name)
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
	<-done
}
