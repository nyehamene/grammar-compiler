package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Server struct {
	reader   *bufio.Reader
	writer   io.Writer
	log      *log.Logger // Add logger field
	shutdown bool        // Add shutdown field
	// Add more fields as needed, e.g., for managing open files, diagnostics, etc.
}

func NewServer(r io.Reader, w io.Writer) *Server {
	// Set up logging to a file
	logPath := filepath.Join(os.Getenv("HOME"), ".cache", "grammar")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		// Fallback to stderr if file logging fails
		return newServer(bufio.NewReader(r), w, os.Stderr)
	}
	logFilePath := filepath.Join(logPath, "lsp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		// Fallback to stderr if file logging fails
		return newServer(bufio.NewReader(r), w, os.Stderr)
	}
	return newServer(r, w, logFile)
}

func newServer(r io.Reader, w io.Writer, out io.Writer) *Server {
	logger := log.New(out, "lsp: ", log.Ldate|log.Ltime|log.Lshortfile)
	return &Server{
		reader:   bufio.NewReader(r),
		writer:   w,
		log:      logger,
		shutdown: false,
	}
}

func (s *Server) Start() {
	s.log.Println("LSP Server started")

	for {
		// Decode incoming message
		content, err := s.DecodeMessage(s.reader)
		if err != nil {
			s.log.Printf("Failed to decode message: %v", err)
			continue
		}

		// Unmarshal into a generic map to determine message type
		var msg map[string]any
		if err := json.Unmarshal(content, &msg); err != nil {
			s.log.Printf("Failed to unmarshal message: %v", err)
			s.sendErrorResponse(0, ParseError, "Failed to unmarshal message")
			continue
		}

		// Determine if it's a request or notification
		if id, ok := msg["id"].(float64); ok { // Requests have an ID
			s.handleRequest(int(id), msg)
		} else { // Notifications do not have an ID
			s.handleNotification(msg)
		}
	}
}

func (s *Server) handleRequest(id int, msg map[string]any) {
	method, ok := msg["method"].(string)
	if !ok {
		s.sendErrorResponse(id, InvalidRequest, "Method not found in request")
		return
	}

	go s.logMessage("Received request", msg)

	switch method {
	case "initialize":
		handleInitializeRequest(s, id, msg)
	case "shutdown":
		s.handleShutdown(id)
	default:
		// For now, just send a simple response
		s.sendResponse(id, fmt.Sprintf("Received method %s with params %v", method, msg["params"]), nil)
	}
}

func (s *Server) handleShutdown(id int) {
	s.shutdown = true
	s.sendResponse(id, nil, nil) // Respond with null result
	s.log.Printf("Shutdown request received. Server state: shutdown=%t", s.shutdown)
}

func (s *Server) handleNotification(msg map[string]any) {
	method, ok := msg["method"].(string)
	if !ok {
		s.log.Printf("Notification without method: %v", msg)
		return
	}

	go s.logMessage("Received notification", msg)

	switch method {
	case "exit":
		s.handleExit()
	default:
		// For now, just log other notifications
	}
}

func (s *Server) handleExit() {
	if s.shutdown {
		os.Exit(0)
	}
	os.Exit(1)
}

func (s *Server) sendResponse(id int, result any, errResp *ResponseError) {
	resp := ResponseMessage{
		Message: Message{JSONRPC: "2.0"},
		ID:      &id,
		Result:  &result,
		Error:   errResp,
	}
	encoded, err := s.EncodeMessage(resp)
	if err != nil {
		s.log.Printf("Failed to encode response: %v", err)
		return
	}
	_, err = s.writer.Write([]byte(encoded))
	if err != nil {
		s.log.Printf("Failed to write response: %v", err)
	}

	go s.logMessage("Sent response", resp)
}

func (s *Server) logMessage(title string, msg any) {
	loggedMsg, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		s.log.Printf("Failed to marshal message for logging: %v", err)
	} else {
		s.log.Printf("%s:\n%s", title, loggedMsg)
	}
}

func (s *Server) sendErrorResponse(id int, code ErrorCodes, message string) {
	errResp := &ResponseError{
		Code:    int(code),
		Message: message,
	}
	s.sendResponse(id, nil, errResp)
}
