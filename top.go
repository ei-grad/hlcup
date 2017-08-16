package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func top() {

	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	cmd := exec.Command("top", "-b", "-d10")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()

	<-gracefulStop

	cmd.Process.Kill()

	cmd.Wait()

	os.Exit(0)

}
