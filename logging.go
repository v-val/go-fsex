package main

import (
	L "log"
	"os"
	"strconv"
)

type logging_ struct {
	Quietness incrementableInt
}

var loggingInstance logging_

func SetQuietness(q incrementableInt) {
	loggingInstance.Quietness = q
}

func Trace(message any) {
	if loggingInstance.Quietness < -1 {
		L.Println(message)
	}
}

func Tracef(format string, args ...any) {
	if loggingInstance.Quietness < -1 {
		L.Printf(format, args...)
	}
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

func init() {
	const EnvVerbosity = "FSEX_VERBOSITY"
	s := os.Getenv(EnvVerbosity)
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			L.Panicf(`Environment "%s" must be int`, EnvVerbosity)
		}
		SetQuietness(incrementableInt(i))
	}
}
