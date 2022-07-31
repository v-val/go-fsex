// https://github.com/fsnotify/fsnotify
package main

import (
	"log"
    "flag"
	"github.com/fsnotify/fsnotify"
)

func main() {
    log.Fatal("Arguments: %v", flag.Args())

	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer watcher.Close()

	// done := make(chan bool)
	// go func() {
	// 	for {
	// 		select {
	// 		case event, ok := <-watcher.Events:
	// 			if !ok {
	// 				return
	// 			}
	// 			log.Println("event:", event)
	// 			if event.Has(fsnotify.Write) {
	// 				log.Println("modified file:", event.Name)
	// 			}
	// 		case err, ok := <-watcher.Errors:
	// 			if !ok {
	// 				return
	// 			}
	// 			log.Println("error:", err)
	// 		}
	// 	}
	// }()

	// err = watcher.Add("/tmp")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// <-done
}
