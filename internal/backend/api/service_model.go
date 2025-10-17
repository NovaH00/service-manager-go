package api

import "service-manager/internal/manager"

type ServiceData struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Cmd              manager.Command `json:"cmd"`
	ExecuteDirectory string          `json:"execute_directory"`
	IsRunning        bool            `json:"is_running"`
}

type ServiceMetrics struct {
	Uptime     int64   `json:"uptime"`
	CPUPercent float64 `json:"cpu_percent"`
	RAMUsage   float64 `json:"ram_usage"`
}

type NetworkInfo struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
}
