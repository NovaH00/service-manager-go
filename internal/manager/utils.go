package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v3/process"
)

func openAppend(filePath string) (*os.File, error) {
	// Flags to open file in append mode, if file is not exist then create and open
	// in append mode
	parentDir := filepath.Dir(filePath)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("create parent dir: %w", err)
	}

	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY

	return os.OpenFile(filePath, flags, 0644)
}

// collectNetworkInfoRecursive gets all LISTEN connections for the process and its children
// and normalizes "::" to "0.0.0.0"
func collectNetworkInfoRecursive(proc *process.Process) []NetworkInfo {
	res := []NetworkInfo{}

	// Get connections for this process
	conns, err := proc.Connections()
	if err == nil {
		for _, c := range conns {
			if c.Status == "LISTEN" {
				ip := c.Laddr.IP
				if ip == "::" {
					ip = "0.0.0.0" // normalize IPv6 "all interfaces" to IPv4 style
				}

				res = append(res, NetworkInfo{
					IP:   ip,
					Port: c.Laddr.Port,
				})
			}
		}
	}

	// Recurse into children
	children, err := proc.Children()
	if err == nil {
		for _, child := range children {
			res = append(res, collectNetworkInfoRecursive(child)...)
		}
	}

	return res
}
