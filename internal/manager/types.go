package manager

type ServiceStatus string

const (
	SERVICE_UNKNOWN ServiceStatus = "service_unknown"
	SERVICE_RUNNING ServiceStatus = "service_running"
	SERVICE_STOPPED ServiceStatus = "service_stopped"
)

type Command struct {
	// Name is the name of the executable/binary
	Name string `json:"name"`
	// Arguments is the arguments of the command
	Arguments []string `json:"args"`
}

type serviceData struct {
	ID               string
	Name             string
	Cmd              Command
	ExecuteDirectory string
}

type ResourcesData struct {
	CPUPercent float64
	RAMUsage   float64
}

type NetworkInfo struct {
	IP   string
	Port uint32
}
