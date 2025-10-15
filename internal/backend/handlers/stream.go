package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"service-manager/internal/backend/api"
	"service-manager/internal/backend/utils"
	"service-manager/internal/manager"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

const INITIAL_LINES_OF_LOG = 10000

type StreamHandler struct {
	ServiceManager *manager.ServiceManager
	LogsDir        string
}

func NewStreamHandler(sm *manager.ServiceManager, logsDir string) *StreamHandler {
	return &StreamHandler{
		ServiceManager: sm,
		LogsDir:        logsDir,
	}
}

func sendSSE(c *gin.Context, msg any) bool {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		// This error is a server-side issue, so we can't easily send it to the client.
		// We'll log it and close the connection by returning false.
		log.Printf("Error marshalling SSE message: %v", err)
		return false
	}

	_, err = fmt.Fprintf(c.Writer, "data: %s\n\n", jsonBytes)
	if err != nil {
		// This likely means the client has closed the connection.
		log.Printf("Error writing to SSE stream: %v", err)
		return false
	}

	c.Writer.Flush()
	return true
}

func sendSSEError(c *gin.Context, title, detail string) bool {
	apiErr := api.NewError(title, detail)
	errMsg := api.StreamMessage{
		Type: api.EVENT_ERROR,
		Data: apiErr,
	}
	return sendSSE(c, errMsg)
}

// StreamStdout godoc
// @Summary      Stream service stdout logs
// @Description  Streams the standard output log of a service using Server-Sent Events (SSE).
// @Tags         stream
// @Produce      text/event-stream
// @Param        serviceID  path      string  true  "Service ID"
// @Success      200        {object}  api.StreamMessage  "SSE stream of log data"
// @Router       /stream/stdout/{serviceID} [get]
func (h *StreamHandler) StreamStdout(c *gin.Context) {
	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	serviceID := c.Param("serviceID")

	if !h.ServiceManager.ServiceExists(serviceID) {
		sendSSEError(c, "Service does not exist", fmt.Sprintf("Could not find service with id '%s'", serviceID))
		return
	}

	fullFilePath := filepath.Join(h.LogsDir, serviceID, "stdout")

	// Read initial lines, but handle "file not found" gracefully.
	lines, err := utils.ReadLines(fullFilePath, INITIAL_LINES_OF_LOG)
	if err != nil && !os.IsNotExist(err) {
		sendSSEError(c, "Error reading initial lines", err.Error())
		return
	}

	initialMessage := api.StreamMessage{
		Type: api.EVENT_INITIAL,
		Data: lines, // lines will be nil if the file didn't exist, which is fine.
	}
	if !sendSSE(c, initialMessage) {
		return // Client disconnected
	}

	// Stream message on file modification
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		sendSSEError(c, "Error creating file watcher", err.Error())
		return
	}
	defer watcher.Close()

	// It's crucial to add the directory to the watcher, not the file.
	// This allows us to detect when the file is created.
	logDir := filepath.Dir(fullFilePath)
	err = watcher.Add(logDir)
	if err != nil {
		sendSSEError(c, "Error adding file path to watcher", err.Error())
		return
	}

	for {
		select {
		case <-c.Request.Context().Done():
			log.Println("Client disconnected")
			return
		case event := <-watcher.Events:
			// We only care about events for our specific file.
			if event.Name != fullFilePath {
				continue
			}

			// We care about the file being written to or created.
			if !(event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				continue
			}

			lastLine, err := utils.ReadLastLine(fullFilePath)
			if err != nil {
				// Don't send an error here, as it could be a transient read issue.
				// We can just log it and wait for the next event.
				log.Printf("Error reading last line: %v", err)
				continue
			}

			appendMessage := api.StreamMessage{
				Type: api.EVENT_APPEND,
				Data: lastLine,
			}
			if !sendSSE(c, appendMessage) {
				return // Client disconnected
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
			// Optionally, send this error to the client.
			sendSSEError(c, "File watcher error", err.Error())
			return // Client disconnected

		}
	}
}

// StreamStderr godoc
// @Summary      Stream service stderr logs
// @Description  Streams the standard error log of a service using Server-Sent Events (SSE).
// @Tags         stream
// @Produce      text/event-stream
// @Param        serviceID  path      string  true  "Service ID"
// @Success      200        {object}  api.StreamMessage  "SSE stream of log data"
// @Router       /stream/stderr/{serviceID} [get]
func (h *StreamHandler) StreamStderr(c *gin.Context) {
	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	serviceID := c.Param("serviceID")

	if !h.ServiceManager.ServiceExists(serviceID) {
		sendSSEError(c, "Service does not exist", fmt.Sprintf("Could not find service with id '%s'", serviceID))
		return
	}

	fullFilePath := filepath.Join(h.LogsDir, serviceID, "stderr")

	// Read initial lines, but handle "file not found" gracefully.
	lines, err := utils.ReadLines(fullFilePath, INITIAL_LINES_OF_LOG)
	if err != nil && !os.IsNotExist(err) {
		sendSSEError(c, "Error reading initial lines", err.Error())
		return
	}

	initialMessage := api.StreamMessage{
		Type: api.EVENT_INITIAL,
		Data: lines, // lines will be nil if the file didn't exist, which is fine.
	}
	if !sendSSE(c, initialMessage) {
		return // Client disconnected
	}

	// Stream message on file modification
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		sendSSEError(c, "Error creating file watcher", err.Error())
		return
	}
	defer watcher.Close()

	// It's crucial to add the directory to the watcher, not the file.
	// This allows us to detect when the file is created.
	logDir := filepath.Dir(fullFilePath)
	err = watcher.Add(logDir)
	if err != nil {
		sendSSEError(c, "Error adding file path to watcher", err.Error())
		return
	}

	for {
		select {
		case <-c.Request.Context().Done():
			log.Println("Client disconnected")
			return
		case event := <-watcher.Events:
			// We only care about events for our specific file.
			if event.Name != fullFilePath {
				continue
			}

			// We care about the file being written to or created.
			if !(event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				continue
			}

			lastLine, err := utils.ReadLastLine(fullFilePath)
			if err != nil {
				// Don't send an error here, as it could be a transient read issue.
				// We can just log it and wait for the next event.
				log.Printf("Error reading last line: %v", err)
				continue
			}

			appendMessage := api.StreamMessage{
				Type: api.EVENT_APPEND,
				Data: lastLine,
			}
			if !sendSSE(c, appendMessage) {
				return // Client disconnected
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
			// Optionally, send this error to the client.
			sendSSEError(c, "File watcher error", err.Error())
			return // Client disconnected

		}
	}
}
