package main

import (
	L "log"
)

type logging_ struct {
	Quietness incrementableInt
}

var loggingInstance = logging_{
	Quietness: incrementableInt(0),
}

func SetQuietness(q incrementableInt) {
	loggingInstance.Quietness = q
}

func Debug(message any) {
	if loggingInstance.Quietness < 0 {
		L.Println(message)
	}
}

func Debugf(format string, args ...any) {
	if loggingInstance.Quietness < 0 {
		L.Printf(format, args...)
	}
}

func Print(message any) {
	if loggingInstance.Quietness < 1 {
		L.Println(message)
	}
}

func Printf(format string, args ...any) {
	if loggingInstance.Quietness < 1 {
		L.Printf(format, args...)
	}
}

func Fatal(message any) {
	L.Fatalln(message)
}

func Fatalf(format string, args ...any) {
	L.Fatalf(format, args...)
}

func Panic(message any) {
	L.Panicln(message)
}

func Panicf(format string, args ...any) {
	L.Panicf(format, args...)
}
