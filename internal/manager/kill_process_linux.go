//go:build linux

package manager

import (
	"context"
	"log"
	"syscall"
)

func killProcessOnCancel(killContext context.Context, serviceName string, PID int) {
	<-killContext.Done()

	log.Printf("Context cancelled for service %s", serviceName)

	// Kill the whole process group (negative PID kills PGID)
	pgid, err := syscall.Getpgid(PID)
	if err != nil {
		log.Printf("Getpgid failed: %v", err)
		return
	}

	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		log.Printf("Failed to kill process group: %v", err)
		return
	}

	log.Printf("Killed service %s (pid %d, pgid %d)", serviceName, PID, pgid)
}
