package handlers

import (
	"fmt"
	"net/http"
	"service-manager/internal/backend/api"
	"service-manager/internal/backend/helpers"
	"service-manager/internal/manager"

	"github.com/gin-gonic/gin"
)

type ServiceManagerHandler struct {
	ServiceManager *manager.ServiceManager
}

func NewServiceManagerHandler(sm *manager.ServiceManager) *ServiceManagerHandler {
	return &ServiceManagerHandler{
		ServiceManager: sm,
	}
}

// RegisterService godoc
// @Summary      Register a new service
// @Description  Registers a new service with the service manager.
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.RegisterServiceRequest  true  "Service Registration"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  api.ErrorResponse
// @Failure      422      {object}  api.ErrorResponse
// @Router       /manager/register [post]
func (h *ServiceManagerHandler) RegisterService(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.RegisterServiceRequest](c)
	if !ok {
		return
	}

	err := h.ServiceManager.RegisterService(
		req.ServiceName,
		req.CommandName,
		req.CommandArgs,
		req.ExecuteDirectory,
	)
	if err != nil {
		apiError := api.NewError(
			fmt.Sprintf("cannot register service '%s'", req.ServiceName),
			err.Error(),
		)

		c.JSON(
			http.StatusUnprocessableEntity,
			apiError,
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "register service successful"})
}

// RemoveService godoc
// @Summary      Remove a service
// @Description  Removes a service from the service manager.
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.ServiceIDRequest  true  "Service ID"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /manager/remove [delete]
func (h *ServiceManagerHandler) RemoveService(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.ServiceIDRequest](c)
	if !ok {
		return
	}

	err := h.ServiceManager.RemoveService(req.ServiceID)
	if err != nil {
		apiError := api.NewError(
			"failed to remove service",
			err.Error(),
		)

		c.JSON(
			http.StatusInternalServerError,
			apiError,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"message": "remove service successful"},
	)
}

// StartService godoc
// @Summary      Start a service
// @Description  Starts a registered service.
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.ServiceIDRequest  true  "Service ID"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /manager/start [post]
func (h *ServiceManagerHandler) StartService(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.ServiceIDRequest](c)
	if !ok {
		return
	}

	err := h.ServiceManager.StartService(req.ServiceID)
	if err != nil {
		apiError := api.NewError(
			"failed to start service",
			err.Error(),
		)

		c.JSON(
			http.StatusInternalServerError,
			apiError,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"message": "service started"},
	)
}

// StopService godoc
// @Summary      Stop a service
// @Description  Stops a running service.
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.ServiceIDRequest  true  "Service ID"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /manager/stop [post]
func (h *ServiceManagerHandler) StopService(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.ServiceIDRequest](c)
	if !ok {
		return
	}

	err := h.ServiceManager.StopService(req.ServiceID)
	if err != nil {
		apiError := api.NewError(
			"Failed to stop service",
			err.Error(),
		)

		c.JSON(
			http.StatusInternalServerError,
			apiError,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"message": "service stopped"},
	)
}

// GetServices godoc
// @Summary      Get all services
// @Description  Retrieves a list of all registered services and their statuses.
// @Tags         manager
// @Produce      json
// @Success      200  {array}   api.ServiceData
// @Router       /manager/services [get]
func (h *ServiceManagerHandler) GetServices(c *gin.Context) {
	services := h.ServiceManager.GetAllServices()

	response := make([]api.ServiceData, 0, len(services))

	for _, service := range services {
		IsRunning := service.GetStatus() == manager.SERVICE_RUNNING

		serviceData := api.ServiceData{
			ID:               service.ID,
			Name:             service.Name,
			Cmd:              service.Cmd,
			ExecuteDirectory: service.ExecuteDirectory,
			IsRunning:        IsRunning,
		}
		response = append(response, serviceData)
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// GetServiceMetrics godoc
// @Summary      Get metrics of an service
// @Description  Get the metrics such as cpu percentage, ram usage and uptime.
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.ServiceIDRequest  true  "Service ID"
// @Success      200      {object}  api.ServiceMetrics
// @Failure      400      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /manager/metrics [post]
func (h *ServiceManagerHandler) GetServiceMetrics(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.ServiceIDRequest](c)
	if !ok {
		return
	}

	service, err := h.ServiceManager.GetService(req.ServiceID)
	if err != nil {
		apiError := api.NewError(
			"error fetching service",
			err.Error(),
		)

		c.JSON(
			http.StatusInternalServerError,
			apiError,
		)
		return
	}

	serviceResourcesUsage := service.GetResourcesUsage()
	serviceMetrics := api.ServiceMetrics{
		Uptime:     service.GetUptime(),
		CPUPercent: serviceResourcesUsage.CPUPercent,
		RAMUsage:   serviceResourcesUsage.RAMUsage,
	}

	c.JSON(
		http.StatusOK,
		serviceMetrics,
	)
}

// GetNetworkInfo godoc
// @Summary      Get network information of a service
// @Description  Return info like ip address, port. May expand in the future
// @Tags         manager
// @Accept       json
// @Produce      json
// @Param        service  body      api.ServiceIDRequest  true  "Service ID"
// @Success      200      {object}  api.NetworkInfo
// @Failure      400      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /manager/network [post]
func (h *ServiceManagerHandler) GetNetworkInfo(c *gin.Context) {
	req, ok := helpers.BindOrAbort[api.ServiceIDRequest](c)
	if !ok {
		return
	}

	service, err := h.ServiceManager.GetService(req.ServiceID)
	if err != nil {
		apiError := api.NewError(
			"error fetching service",
			err.Error(),
		)

		c.JSON(
			http.StatusInternalServerError,
			apiError,
		)
		return
	}

	c.JSON(
		http.StatusOK,
		service.GetNetworkInfo(),
	)
}
