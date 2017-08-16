package main

import (
	"os"
	"os/exec"
)

func init() {
	cmd := exec.Command("top", "-b")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
