//go:build windows

package manager

import (
	"context"
	"log"
	"os/exec"
	"strconv"
)

func killProcessOnCancel(killContext context.Context, serviceName string, PID int) {
	<-killContext.Done()

	log.Printf("Context cancelled for service %s", serviceName)

	// Have to run taskkill manually because fucking Windows refuses to play nicely with process kill
	// Fuck Microsoft
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(PID))
	if err := killCmd.Run(); err != nil {
		log.Printf("taskkill failed: %v", err)
		return
	}
	log.Printf("taskkill service %s successfully", serviceName)
}
