package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"syscall"
	"time"
)

func memstats() {
	var memstats runtime.MemStats
	if os.Getenv("DEBUG_MEMSTATS") != "1" {
		return
	}
	go func() {
		for {
			runtime.ReadMemStats(&memstats)
			log.Printf("memstats: %+v", memstats)
			time.Sleep(10 * time.Second)
		}
	}()
}

func rlimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatal("error getting RLIMIT_NOFILE: ", err)
	}
	log.Printf("NOFILE: %+v", rLimit)
}

func whoami() {
	u, _ := user.Current()
	log.Printf("Who am I: %+v", u)
}

func top() {

	time.Sleep(10 * time.Second)

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
