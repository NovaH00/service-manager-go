package api

type RegisterServiceRequest struct {
	ServiceName      string   `json:"service_name" binding:"required"`
	CommandName      string   `json:"command_name" binding:"required"`
	CommandArgs      []string `json:"command_args"`
	ExecuteDirectory string   `json:"execute_directory"`
}

type ServiceIDRequest struct {
	ServiceID string `json:"service_id"`
}
