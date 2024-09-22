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

func (app *fsex) execCommand() {
	// TODO: select line width respecting terminal properties
	const hrWidth = 48
	//headOpen := strings.Repeat("=", hrWidth) + "\n"
	headClose := strings.Repeat("-", hrWidth)
	bodyEndError := strings.Repeat("!", hrWidth) + "\n"
	bodyEndOk := strings.Repeat(".", hrWidth) + "\n"
	//hrBefore := headOpen + fmt.Sprintf("RUN %v\n", app.cmd) + headClose
	var cmd_ *exec.Cmd
	if len(app.cmd) == 1 {
		cmd_ = exec.Command(app.cmd[0])
	} else {
		cmd_ = exec.Command(app.cmd[0], app.cmd[1:]...)
	}
	if !app.flagSuppressStdout {
		cmd_.Stdout = os.Stdout
	}
	if !app.flagSuppressStderr {
		cmd_.Stderr = os.Stderr
	}
	if app.flagClearScreenOnChanges {
		Print("Clear the screen..")
		screen.Clear()
		screen.MoveTopLeft()
	}
	Printf("RUN %v\n"+headClose, app.cmd)
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
