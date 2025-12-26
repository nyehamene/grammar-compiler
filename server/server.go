package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type Server struct {
	reader *bufio.Reader
	writer io.Writer
	// Add more fields as needed, e.g., for managing open files, diagnostics, etc.
}

func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

func (s *Server) Start() {
	log.Println("LSP Server started")

	for {
		// Decode incoming message
		content, err := DecodeMessage(s.reader)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			continue
		}

		// Unmarshal into a generic map to determine message type
		var msg map[string]any
		if err := json.Unmarshal(content, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			s.sendErrorResponse(nil, ParseError, "Failed to unmarshal message")
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
		s.sendErrorResponse(&id, InvalidRequest, "Method not found in request")
		return
	}

	params := msg["params"]

	log.Printf("Received request: id=%d, method=%s, params=%v", id, method, params)

	// For now, just send a simple response
	s.sendResponse(id, fmt.Sprintf("Received method %s with params %v", method, params), nil)
}

func (s *Server) handleNotification(msg map[string]any) {
	method, ok := msg["method"].(string)
	if !ok {
		log.Printf("Notification without method: %v", msg)
		return
	}
	params := msg["params"]

	log.Printf("Received notification: method=%s, params=%v", method, params)

	// For now, just log notifications
}

func (s *Server) sendResponse(id int, result any, errResp *ResponseError) {
	resp := ResponseMessage{
		Message: Message{JSONRPC: "2.0"},
		ID:      &id,
		Result:  &result,
		Error:   errResp,
	}
	encoded, err := EncodeMessage(resp)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}
	_, err = s.writer.Write([]byte(encoded))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) sendErrorResponse(id *int, code ErrorCodes, message string) {
	errResp := &ResponseError{
		Code:    int(code),
		Message: message,
	}
	s.sendResponse(*id, nil, errResp)
}
