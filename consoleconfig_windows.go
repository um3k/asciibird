package main

import (
	"os/exec"
)

func SetupConsole() {
	cmd := exec.Command("mode", "con", "lines=40", "cols=80")
	err := cmd.Run()
	if err != nil {
		return
	}
}