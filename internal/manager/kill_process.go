package manager

import (
	"context"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
)

func killProcessOnCancel(killContext context.Context, serviceName string, PID int) {
	<-killContext.Done()

	log.Printf("Context cancelled for service %s", serviceName)

	switch runtime.GOOS {
	case "windows":
		// Have to run taskkill manually because fucking Windows refuses to play nicely with process kill
		// Fuck Microsoft
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(PID))
		if err := killCmd.Run(); err != nil {
			log.Printf("taskkill failed: %v", err)
			return
		}
		log.Printf("taskkill service %s successfully", serviceName)

	case "linux":
		process, err := os.FindProcess(PID)
		if err != nil {
			log.Printf("os.FindProcess(%d): %v", PID, err)
			return
		}

		err = process.Signal(syscall.SIGKILL)
		if err != nil {
			log.Printf("SIGKILL system call: %v", err)
			return
		}

		log.Printf("SIGKILL service %s successfully", serviceName)
	}
}
