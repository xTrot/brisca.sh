package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"brisca.sh/server"
	"github.com/charmbracelet/log"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		server.Start()
	}()

	sig := <-sigChan
	log.Debug("Initiating graceful shutdown, Received signal: ", "sig", sig)

	log.Debug("Performing cleanup operations...")
	server.ShutdownMark = true
	time.Sleep(60 * time.Second)

	log.Debug("Application shut down gracefully.")
	os.Exit(0)
}
