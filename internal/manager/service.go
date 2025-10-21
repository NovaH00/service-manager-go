package manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
)

// service represents a command or process that is managed by the service manager.
// It holds all the necessary information to start, monitor, and stop the service.
type service struct {
	ID               string
	Name             string
	Cmd              Command
	ExecuteDirectory string
	pid              int
	startTime        time.Time
	status           ServiceStatus
	stdoutHandler    func(service *service, line string)
	stderrHandler    func(service *service, line string)
	cancelService    context.CancelFunc
	mutex            sync.Mutex
	commandWaitGroup sync.WaitGroup
}

func (s *service) streamOutput(reader io.ReadCloser, handler func(service *service, line string)) {
	lines := make(chan string, 100)

	// 64KB per lines, or roughly 65000 characters per line
	const SCANNER_MAX_CAPACITY = 64 * 1024

	go func() {
		defer close(lines)

		scanner := bufio.NewScanner(reader)

		buf := make([]byte, SCANNER_MAX_CAPACITY)
		scanner.Buffer(buf, SCANNER_MAX_CAPACITY)

		for scanner.Scan() {
			lines <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Printf("error reading: %v", err)
		}
	}()

	for line := range lines {
		handler(s, line)
	}
}

func (s *service) setStatus(status ServiceStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.status = status
}

func (s *service) GetStatus() ServiceStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.status
}

func (s *service) GetResourcesUsage() ResourcesData {
	proc, err := process.NewProcess(int32(s.pid))
	if err != nil {
		return ResourcesData{
			CPUPercent: 0.0,
			RAMUsage:   0.0,
		}
	}
	// We ignore errors since this doesn't affect service operation
	CPUPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()

	return ResourcesData{
		CPUPercent: CPUPercent,
		RAMUsage:   float64(memInfo.RSS) / 1024.0 / 1024.0,
	}
}

func (s *service) GetUptime() int64 {
	var zeroTime time.Time

	// Service not started == uptime zero
	if time.Time.Equal(s.startTime, zeroTime) {
		return 0
	}

	return int64(time.Since(s.startTime).Seconds())
}

func (s *service) GetNetworkInfo() []NetworkInfo {
	defaultNetworkInfo := []NetworkInfo{
		{IP: "", Port: 0},
	}

	proc, err := process.NewProcess(int32(s.pid))
	if err != nil {
		return defaultNetworkInfo
	}

	res := collectNetworkInfoRecursive(proc)

	if len(res) == 0 {
		return defaultNetworkInfo
	}

	return res
}

func (s *service) executeCommand(ctx context.Context) error {
	defer s.commandWaitGroup.Done()

	cmd := exec.Command(s.Cmd.Name, s.Cmd.Arguments...)
	cmd.Dir = s.ExecuteDirectory
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags:    windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
		NoInheritHandles: true,
	}

	s.setStatus(SERVICE_RUNNING)
	s.startTime = time.Now()

	defer func() {
		s.setStatus(SERVICE_STOPPED)
		var zeroTime time.Time
		s.startTime = zeroTime
	}()

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	// Save pid to monitor resources usage
	s.pid = cmd.Process.Pid

	go killProcessOnCancel(ctx, s.Name, cmd.Process.Pid)

	go s.streamOutput(outReader, s.stdoutHandler)
	go s.streamOutput(errReader, s.stderrHandler)

	err = cmd.Wait()
	if ctx.Err() == context.Canceled {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("exit command: %w", err)
	}

	return nil
}

func (s *service) Start(serviceContext context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.status == SERVICE_RUNNING {
		return fmt.Errorf("service '%s' (ID: '%s') is already running", s.Name, s.ID)
	}

	ctx, cancel := context.WithCancel(serviceContext)

	s.commandWaitGroup.Add(1)
	go s.executeCommand(ctx)

	s.status = SERVICE_RUNNING
	s.cancelService = cancel

	return nil
}

func (s *service) Stop() error {
	s.mutex.Lock()

	if s.status == SERVICE_STOPPED {
		s.mutex.Unlock()
		return fmt.Errorf("service '%s' (ID: '%s') is not running", s.Name, s.ID)
	}

	if s.cancelService != nil {
		s.cancelService()
	}
	s.mutex.Unlock()

	// Wait for command to exit
	s.commandWaitGroup.Wait()

	s.mutex.Lock()
	s.cancelService = nil
	s.mutex.Unlock()

	return nil
}

func newService(
	serviceID string,
	serviceName string,
	commandName string,
	commandArgs []string,
	executeDirectory string,
	stdoutHandler func(service *service, line string),
	stderrHandler func(service *service, line string),

) (*service, error) {
	if serviceID == "" {
		serviceID = uuid.New().String()
	}

	if serviceName == "" {
		return nil, errors.New("service name cannot be empty")
	}

	if commandName == "" {
		return nil, errors.New("command name cannot be empty")
	}

	service := &service{
		ID:   serviceID,
		Name: serviceName,
		Cmd: Command{
			Name:      commandName,
			Arguments: commandArgs,
		},
		ExecuteDirectory: executeDirectory,
		status:           SERVICE_STOPPED,
		stdoutHandler:    stdoutHandler,
		stderrHandler:    stderrHandler,
	}

	return service, nil
}
