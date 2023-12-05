package main

import (
	"errors"
	"github.com/inancgumus/screen"
	"os"
	"os/exec"
	"strings"
)

type fsex struct {
	// Configuration Parameters
	cmd                      []string
	flagClearScreenOnChanges bool
	flagSuppressStdout       bool
	flagSuppressStderr       bool
	// Constants
	// Operational vars
}

func (f *fsex) execCommand() {
	// TODO: select line width respecting terminal properties
	const hrWidth = 48
	//headOpen := strings.Repeat("=", hrWidth) + "\n"
	headClose := strings.Repeat("-", hrWidth)
	bodyEndError := strings.Repeat("!", hrWidth) + "\n"
	bodyEndOk := strings.Repeat(".", hrWidth) + "\n"
	//hrBefore := headOpen + fmt.Sprintf("RUN %v\n", f.cmd) + headClose
	var cmd_ *exec.Cmd
	if len(f.cmd) == 1 {
		cmd_ = exec.Command(f.cmd[0])
	} else {
		cmd_ = exec.Command(f.cmd[0], f.cmd[1:]...)
	}
	if !f.flagSuppressStdout {
		cmd_.Stdout = os.Stdout
	}
	if !f.flagSuppressStderr {
		cmd_.Stderr = os.Stderr
	}
	if f.flagClearScreenOnChanges {
		Print("Clear the screen..")
		screen.Clear()
		screen.MoveTopLeft()
	}
	Printf("RUN %v\n"+headClose, f.cmd)
	err := cmd_.Run()
	if err != nil {
		var ee *exec.ExitError
		if loggingInstance.Quietness <= 0 {
			print(bodyEndError)
		}
		if errors.As(err, &ee) {
			Printf("Command returned %d", ee.ExitCode())
		} else {
			Printf("Fail to run command: %s", err)
		}
	} else {
		if loggingInstance.Quietness <= 0 {
			print(bodyEndOk)
		}
		Print("Command completed successfully")
	}
}
