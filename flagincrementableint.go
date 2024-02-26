package main

import "strconv"

// Type to handle flag incrementing underlying parameter
type incrementableInt int

func (i *incrementableInt) String() string {
	return strconv.Itoa(int(*i))
}

func (i *incrementableInt) IsBoolFlag() bool {
	return true
}

func (i *incrementableInt) Set(string) error {
	(*i)++
	return nil
}
