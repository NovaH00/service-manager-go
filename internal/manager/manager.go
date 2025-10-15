package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ServiceManager struct {
	services         map[string]*service
	logsDir          string
	servicesDataPath string
	readWriteMutex   sync.RWMutex
}

func (sm *ServiceManager) GetServiceStatus(serviceID string) (ServiceStatus, error) {
	sm.readWriteMutex.RLock()
	defer sm.readWriteMutex.RUnlock()
	if _, ok := sm.services[serviceID]; !ok {
		return SERVICE_UNKNOWN, fmt.Errorf("service not found")
	}
	return sm.services[serviceID].GetStatus(), nil
}

func (sm *ServiceManager) ServiceExists(serviceID string) bool {
	sm.readWriteMutex.RLock()
	defer sm.readWriteMutex.RUnlock()

	_, ok := sm.services[serviceID]

	return ok
}

func (sm *ServiceManager) stdoutHandler(service *service, line string) {
	// Full log path: <logsDir>/<service ID>/stdout
	subDir := service.ID
	fileName := "stdout"
	fullPath := filepath.Join(sm.logsDir, subDir, fileName)

	file, err := openAppend(fullPath)
	if err != nil {
		log.Println("failed to open file: ", err)
	}
	defer file.Close()

	formattedLine := fmt.Sprintf("%s\n", strings.TrimSpace(line))

	if _, err := file.Write([]byte(formattedLine)); err != nil {
		log.Println("failed to write to file: ", err)
	}
}

func (sm *ServiceManager) stderrHandler(service *service, line string) {
	// Full log path: <logsDir>/<service ID>/stderr
	subDir := service.ID
	fileName := "stderr"
	fullPath := filepath.Join(sm.logsDir, subDir, fileName)

	file, err := openAppend(fullPath)
	if err != nil {
		log.Println("failed to open file: ", err)
	}
	defer file.Close()

	formattedLine := fmt.Sprintf("%s\n", strings.TrimSpace(line))

	if _, err := file.Write([]byte(formattedLine)); err != nil {
		log.Println("failed to write to file: ", err)
	}

}

func (sm *ServiceManager) GetAllServices() []*service {
	sm.readWriteMutex.RLock()
	defer sm.readWriteMutex.RUnlock()

	// Create a slice with a pre-defined capacity.
	servicesSlice := make([]*service, 0, len(sm.services))

	for _, value := range sm.services {
		servicesSlice = append(servicesSlice, value)
	}

	return servicesSlice
}

func (sm *ServiceManager) RegisterService(
	serviceName string,
	commandName string,
	commandArgs []string,
	executeDirectory string,
) error {
	sm.readWriteMutex.Lock()
	defer sm.readWriteMutex.Unlock()

	service, err := newService(
		"",
		serviceName,
		commandName,
		commandArgs,
		executeDirectory,
		sm.stdoutHandler,
		sm.stderrHandler,
	)
	if err != nil {
		return fmt.Errorf("register service: %w", err)
	}

	sm.services[service.ID] = service

	return sm.updateServicesFile()
}

func (sm *ServiceManager) loadService(serviceID, serviceName, executeDirectory string, command Command) error {
	sm.readWriteMutex.Lock()
	defer sm.readWriteMutex.Unlock()

	if _, ok := sm.services[serviceID]; ok {
		return fmt.Errorf("service id is already existed")
	}

	service, err := newService(
		serviceID,
		serviceName,
		command.Name,
		command.Arguments,
		executeDirectory,
		sm.stdoutHandler,
		sm.stderrHandler,
	)
	if err != nil {
		return fmt.Errorf("load service: %w", err)
	}

	sm.services[serviceID] = service

	return nil
}

func (sm *ServiceManager) LoadServices() error {
	file, err := os.Open(sm.servicesDataPath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var servicesData []serviceData

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&servicesData); err != nil {
		return fmt.Errorf("error decoding json: %w", err)
	}

	for _, serviceData := range servicesData {
		err := sm.loadService(
			serviceData.ID,
			serviceData.Name,
			serviceData.ExecuteDirectory,
			serviceData.Cmd,
		)
		if err != nil {
			return fmt.Errorf("error loading service: %w", err)
		}
	}

	return nil
}

func (sm *ServiceManager) updateServicesFile() error {
	file, err := os.Create(sm.servicesDataPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	var servicesData []serviceData

	for _, service := range sm.services {
		serviceData := serviceData{
			ID:   service.ID,
			Name: service.Name,
			Cmd:  service.Cmd,
		}

		servicesData = append(servicesData, serviceData)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(servicesData); err != nil {
		return fmt.Errorf("error encoding json: %w", err)
	}

	return nil
}

func (sm *ServiceManager) RemoveService(serviceID string) error {
	sm.readWriteMutex.Lock()
	defer sm.readWriteMutex.Unlock()

	if _, ok := sm.services[serviceID]; !ok {
		return fmt.Errorf("service not found")
	}

	service := sm.services[serviceID]

	if service.GetStatus() != SERVICE_STOPPED {
		return fmt.Errorf("service '%s' (ID: '%s') is running, cannot remove", service.Name, service.ID)
	}

	delete(sm.services, serviceID)

	serviceLogDir := filepath.Join(sm.logsDir, serviceID)

	err := os.RemoveAll(serviceLogDir)
	if err != nil {
		return fmt.Errorf("remove service log folder: %w", err)
	}

	return sm.updateServicesFile()
}

func (sm *ServiceManager) StartService(serviceID string) error {
	sm.readWriteMutex.RLock()
	defer sm.readWriteMutex.RUnlock()

	service, ok := sm.services[serviceID]

	if !ok {
		return fmt.Errorf("service with ID '%s' not found", serviceID)
	}

	err := service.Start(context.Background())
	if err != nil {
		return fmt.Errorf("error starting service '%s' (ID: '%s'). error: %v", service.Name, service.ID, err)
	}

	return nil
}

func (sm *ServiceManager) StopService(serviceID string) error {
	sm.readWriteMutex.RLock()
	defer sm.readWriteMutex.RUnlock()

	service, ok := sm.services[serviceID]

	if !ok {
		return fmt.Errorf("service with ID '%s' not found", serviceID)
	}

	err := service.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop service '%s' (ID: '%s'). Error: %v", service.Name, service.ID, err)
	}

	return nil
}

func (sm *ServiceManager) StopAllServices() {
	var stopServiceWG sync.WaitGroup

	for serviceID := range sm.services {
		stopServiceWG.Add(1)

		go func(serviceID string) {
			defer stopServiceWG.Done()

			sm.StopService(serviceID)
		}(serviceID)
	}

	stopServiceWG.Wait()
}

func NewServiceManager(logsDir, servicesDataPath string) *ServiceManager {
	return &ServiceManager{
		services:         make(map[string]*service),
		logsDir:          logsDir,
		servicesDataPath: servicesDataPath,
	}
}
