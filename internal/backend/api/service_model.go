package api

import "service-manager/internal/manager"

type ServiceData struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Cmd              manager.Command `json:"cmd"`
	ExecuteDirectory string          `json:"execute_directory"`
	IsRunning        bool            `json:"is_running"`
}
