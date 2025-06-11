package main

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func GetRandom(n uint) (string, error) {
	b := make([]byte, n)
	if c, err := rand.Read(b); err != nil {
		return "", err
	} else if c != int(n) {
		return "", fmt.Errorf(`read %d bytes, expected %d`, c, n)
	}
	return fmt.Sprintf("%X", b), nil
}

func MakeTestDir() (string, func(), error) {
	const EnvKeyTestDir = "FSEX_TEST_PREFIX_DIR"
	const TestDirBaseName = AppId + "-test"
	const NRandBytes = 2
	var f func()
	r, err := GetRandom(NRandBytes)
	t := time.Now()
	s := t.Format("060102150405") + fmt.Sprintf(".%d", int64(t.Nanosecond())/time.Millisecond.Nanoseconds())
	d, ok := os.LookupEnv(EnvKeyTestDir)
	if !ok {
		d = "/tmp"
	}
	r = filepath.Join(d, fmt.Sprintf("%s-%s-%s", TestDirBaseName, s, r))
	err = os.MkdirAll(r, fs.ModePerm)
	if err == nil {
		f = func() {
			if err := os.RemoveAll(r); err != nil {
				panic(err)
			} else {
				Print("Test dir \"%s\" deleted\n", r)
			}
		}
	}
	return r, f, err
}

func testNewFile(t *testing.T, file string, extra ...any) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0600)
	assert.NoError(t, err, `fail to create "%s"`, file)
	if len(extra) > 0 {
		t.Logf("extra[0] %v", extra[0])
		doExtraFn, ok := extra[0].(func(*testing.T, *os.File, ...any))
		assert.True(t, ok, `1st optional argument must be a function func(*testing.T, *os.File, ...any)`)
		if len(extra) > 1 {
			doExtraFn(t, fd, extra[1:]...)
		} else {
			doExtraFn(t, fd)
		}
	}
	//_, err = fd.WriteString("Hello, world!\n")
	//assert.NoError(t, err, `fail to write to "%s"`, file)
	err = fd.Close()
	assert.NoErrorf(t, err, `fail to close "%s"`, file)
}

func testRunApp(args ...string) chan int {
	statusChan := make(chan int)
	go func() {
		os.Args = append([]string{AppId}, args...)
		main()
		statusChan <- -1
	}()
	// Give it time to initialize
	time.Sleep(100 * time.Millisecond)
	return statusChan
}
