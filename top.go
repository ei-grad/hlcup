package main

import (
	"os"
	"os/exec"
)

func init() {
	cmd := exec.Command("top", "-b", "-d10")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
