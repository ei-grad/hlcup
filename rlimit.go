package main

import (
	"log"
	"syscall"
)

func init() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Print("Error Getting Rlimit ", err)
	}
	log.Print("Limits:", rLimit)
}
