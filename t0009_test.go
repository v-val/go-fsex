package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func Test_0009_FsexApp_SingleEvent(t *testing.T) {
	// Create directory
	var dir string
	var clearFn func()
	var err error
	//
	dir, clearFn, err = MakeTestDir()
	assert.NoError(t, err, `fail to create test dir`)
	defer clearFn()
	// Run FSEx watching this directory, record app start time
	messageChan := make(chan int)
	go func() {
		os.Args = []string{`zzup`, `-1`, `-f`, dir, `true`}
		main()
		messageChan <- 1
	}()
	t.Logf(`Watch after "%s" started`, dir)
	// Give it a tick to initialize watch
	time.Sleep(100 * time.Millisecond)
	// Make one change in the directory
	testNewFile(t, path.Join(dir, "FILE"))
	t0 := time.Now()
	var duration time.Duration
	// Test that FSEx exited not less than 500ms after the change
	select {
	case n, ok := <-messageChan:
		t1 := time.Now()
		duration = t1.Sub(t0)
		t.Logf(`%v: got %v (%v), duration %v`, t1, n, ok, duration)
	}
	assert.GreaterOrEqual(t, duration, 500*time.Millisecond)
}

func Test_0009_FsexApp_ThousandEvents(t *testing.T) {
	// Create test directory
	dir, clearFn, err := MakeTestDir()
	assert.NoError(t, err, `fail to create test dir`)
	defer clearFn()

	// Run FsEx watching this directory
	appStatus := testRunApp(`-1`, `-f`, dir, `true`)

	// Update a file with intervals smaller than timeout 500ms
	testNewFile(t, path.Join(dir, "FILE"), func(t_ *testing.T, fd_ *os.File, args ...any) {
		for n := 0; n < 100; n++ {
			const s = "Hello, world!\n"
			_, err_ := fd_.WriteString(s)
			assert.NoError(t_, err_, `fail to write %d bytes`, len(s))
			time.Sleep(10 * time.Millisecond)
		}
	})
	t0 := time.Now()
	var duration time.Duration
	// Test that FSEx exited not less than 500ms after the change
	select {
	case n, ok := <-appStatus:
		t1 := time.Now()
		duration = t1.Sub(t0)
		t.Logf(`%v: got %v (%v), duration %v`, t1, n, ok, duration)
	}
	assert.GreaterOrEqual(t, duration, 500*time.Millisecond)
}

// FsExTestUpdateWaitMax max. expected latency of update detection.
// Planned value is 500ms, but it can be affected by factors that can not be accurately calculated or estimated.
const FsExTestUpdateWaitMax = 700 * time.Millisecond

// Function checking that [outFile] was updated
func waitOutFileUpdateF(t *testing.T, file string, origSize int) (isUpdated bool, wait time.Duration, size int) {
	const waitInterval = 10 * time.Millisecond
	begin := time.Now()
	for !isUpdated && wait < FsExTestUpdateWaitMax {
		if data, err := os.ReadFile(file); err != nil || len(data) == origSize {
			time.Sleep(waitInterval)
			wait += waitInterval
		} else {
			wait = time.Since(begin)
			size = len(data)
			t.Logf(`Command executed after %v, "%s" size %d`, wait, file, len(data))
			isUpdated = true
		}
	}

	return
}

func Test_0009_FsexApp_ReloadOnConfChange(t *testing.T) {
	// Create top test dir
	dir, clearOnSuccess, err := MakeTestDir()
	assert.NoError(t, err, `fail to create test prefix dir`)
	defer clearOnSuccess()

	// Create 2 target directories
	dirA := filepath.Join(dir, "A")
	err = os.Mkdir(dirA, os.ModePerm)
	assert.NoError(t, err, `fail to create "%s"`, dirA)
	dirB := filepath.Join(dir, "B")
	err = os.Mkdir(dirB, os.ModePerm)
	assert.NoError(t, err, `fail to create "%s"`, dirB)

	// Create conf file enabling watch after dir A
	confFile := filepath.Join(dir, DefaultConfFileName)
	{
		var confFd *os.File
		confFd, err = os.OpenFile(confFile, os.O_CREATE|os.O_WRONLY, 0640)
		assert.NoError(t, err, `fail to create conf file`)
		_, err = confFd.WriteString("# Test " + AppId + "\n" + dirA)
		assert.NoError(t, err, `fail to insert target "%s" to conf file`, dirA)
		err = confFd.Close()
		assert.NoError(t, err, `fail to close conf file`)
	}

	// Start app
	var outFile string
	outFile, err = GetRandom(6)
	assert.NoError(t, err, `fail to get random string for file name`)
	outFile += `.out`
	outFile = filepath.Join(dir, outFile)
	_ = testRunApp(`-F`, confFile, `bash`, `-Eeuo`, `pipefail`, `-c`, `echo >> `+outFile)

	//
	// Test body
	//

	// Do something in [dirA] and verify that app detected it after 500ms
	outFileSize := 0
	var isUpdated bool
	var updateDetectedIn time.Duration
	testNewFile(t, filepath.Join(dirA, "Foo"))
	isUpdated, updateDetectedIn, outFileSize = waitOutFileUpdateF(t, outFile, outFileSize)
	assert.True(t, isUpdated, `No update in %v`, updateDetectedIn)
	assert.Lessf(t, updateDetectedIn, FsExTestUpdateWaitMax, `Update detection time %v exceeds ETA %v`, updateDetectedIn, FsExTestUpdateWaitMax)

	// Append [dirB] to config
	// App expected to reload conf and monitor both [dirA] and [dirB]
	{
		var confFd *os.File
		confFd, err = os.OpenFile(confFile, os.O_APPEND|os.O_WRONLY, 0640)
		assert.NoError(t, err, `fail to re-open conf file`)
		_, err = confFd.WriteString("# Test " + AppId + "\n" + dirB)
		assert.NoError(t, err, `fail to append target "%s" to conf file`, dirB)
		err = confFd.Close()
		assert.NoError(t, err, `fail to close conf file`)
	}
	// Give app a tick to detect conf update and reload
	time.Sleep(FsExTestUpdateWaitMax)
	Trace(`Expected conf reloaded with 2 targets "%s" and "%s"`, dirA, dirB)

	// Do something in [dirA] and verify that app detected it after 500ms
	testNewFile(t, filepath.Join(dirA, "Bar"))
	isUpdated, updateDetectedIn, outFileSize = waitOutFileUpdateF(t, outFile, outFileSize)
	assert.True(t, isUpdated, `No update in %v`, updateDetectedIn)
	assert.Lessf(t, updateDetectedIn, FsExTestUpdateWaitMax, `Update detection time %v exceeds ETA %v`, updateDetectedIn, FsExTestUpdateWaitMax)

	// Do something in [dirB] and verify that app detected it after 500ms
	testNewFile(t, filepath.Join(dirB, "Foo"))
	isUpdated, updateDetectedIn, outFileSize = waitOutFileUpdateF(t, outFile, outFileSize)
	assert.True(t, isUpdated, `No update in %v`, updateDetectedIn)
	assert.Lessf(t, updateDetectedIn, FsExTestUpdateWaitMax, `Update detection time %v exceeds ETA %v`, updateDetectedIn, FsExTestUpdateWaitMax)

	// Rewrite config, leave only [dirB]
	{
		var fd *os.File
		fd, err = os.OpenFile(confFile, os.O_TRUNC|os.O_WRONLY, 0640)
		assert.NoError(t, err, `fail to re-open conf file`)
		_, err = fd.WriteString("# Test " + AppId + "\n" + dirB)
		assert.NoError(t, err, `fail to insert target "%s" to conf file`, dirB)
		err = fd.Close()
		assert.NoError(t, err, `fail to close conf file`)
	}
	// Give app a tick to detect conf update and reload
	time.Sleep(FsExTestUpdateWaitMax)
	Trace(`Expected conf reloaded with 1 target "%s"`, dirB)

	// Do something in [dirB] and verify that app detected it after 500ms
	testNewFile(t, filepath.Join(dirB, "Bar"))
	isUpdated, updateDetectedIn, outFileSize = waitOutFileUpdateF(t, outFile, outFileSize)
	assert.True(t, isUpdated, `No update in %v`, updateDetectedIn)
	assert.Lessf(t, updateDetectedIn, FsExTestUpdateWaitMax, `Update detection time %v exceeds ETA %v`, updateDetectedIn, FsExTestUpdateWaitMax)

	// Do something in [dirA] and verify that app NOT detects these changes
	testNewFile(t, filepath.Join(dirA, "Baz"))
	isUpdated, updateDetectedIn, outFileSize = waitOutFileUpdateF(t, outFile, outFileSize)
	assert.False(t, isUpdated, `No update in %v`, updateDetectedIn)
	assert.GreaterOrEqualf(t, updateDetectedIn, FsExTestUpdateWaitMax, `Update detection time %v exceeds ETA %v`, updateDetectedIn, FsExTestUpdateWaitMax)
}
