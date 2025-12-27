package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"grammar/check"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Server struct {
	reader       *bufio.Reader
	writer       io.Writer
	log          *log.Logger
	shutdown     bool
	documents    map[DocumentUri]string
	checker      *check.Checker
	fsFileLoader check.FileLoader
}

func NewServer(r io.Reader, w io.Writer) *Server {
	logPath := filepath.Join(os.Getenv("HOME"), ".cache", "grammar")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return newServer(bufio.NewReader(r), w, os.Stderr)
	}
	logFilePath := filepath.Join(logPath, "lsp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return newServer(bufio.NewReader(r), w, os.Stderr)
	}
	return newServer(r, w, logFile)
}

func newServer(r io.Reader, w io.Writer, out io.Writer) *Server {
	logger := log.New(out, "lsp: ", log.Ldate|log.Ltime|log.Lshortfile)
	srv := &Server{
		reader:    bufio.NewReader(r),
		writer:    w,
		log:       logger,
		shutdown:  false,
		documents: make(map[DocumentUri]string),
	}

	checker := check.NewChecker(
		check.SetFileLoader(srv),
		check.SetLogger(logger),
	)
	srv.checker = checker
	srv.fsFileLoader = &check.FileSystemFileLoader{}
	return srv
}

func (s *Server) GetDocumentContent(uri DocumentUri) (string, bool) {
	content, ok := s.documents[uri]
	return content, ok
}

func (s *Server) Start() {
	s.log.Println("LSP Server started")
	for {
		content, err := s.DecodeMessage(s.reader)
		if err == io.EOF {
			s.log.Println("client disconnected!")
			return
		}
		if err != nil {
			s.log.Printf("Failed to decode message: %v", err)
			continue
		}

		var msg map[string]any
		if err := json.Unmarshal(content, &msg); err != nil {
			s.log.Printf("Failed to unmarshal message: %v\n%s\n", err, content)
			s.sendErrorResponse(0, ParseError, "Failed to unmarshal message")
			break
		}

		if id, ok := msg["id"].(float64); ok {
			s.handleRequest(int(id), msg)
		} else {
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

	s.log.Printf("Received request %d-'%s'", id, method)

	switch method {
	case "initialize":
		handleInitializeRequest(s, id, msg)
	case "textDocument/formatting":
		s.handleTextDocumentFormatting(id, msg)
	case "shutdown":
		s.handleShutdown(id)
	case "textDocument/hover":
		s.handleHover(id, msg)
	case "textDocument/definition":
		s.handleDefinition(id, msg)
	default:
		s.sendResponse(id, fmt.Sprintf("Received method %s with params %v", method, msg["params"]), nil)
		s.log.Printf("unexpected method: %s", method)
	}
}

func (s *Server) handleShutdown(id int) {
	s.shutdown = true
	s.sendResponse(id, nil, nil)
	s.log.Printf("Shutdown request received. Server state: shutdown=%t", s.shutdown)
}

func (s *Server) handleNotification(msg map[string]any) {
	method, ok := msg["method"].(string)
	if !ok {
		s.log.Printf("Notification without method: %v", msg)
		return
	}

	s.log.Printf("Received notification: %s", method)

	ctx := context.Background()

	switch method {
	case "initialized":
		if err := s.handleInitialized(ctx, msg); err != nil {
			s.log.Printf("Failed to handle initialized: %v", err)
		}
	case "textDocument/didOpen":
		if err := s.handleDidOpen(ctx, msg); err != nil {
			s.log.Printf("Failed to handle textDocument/didOpen: %v", err)
		}
	case "textDocument/didChange":
		if err := s.handleDidChange(ctx, msg); err != nil {
			s.log.Printf("Failed to handle textDocument/didChange: %v", err)
		}
	case "textDocument/didClose":
		if err := s.handleDidClose(ctx, msg); err != nil {
			s.log.Printf("Failed to handle textDocument/didClose: %v", err)
		}
	case "exit":
		s.handleExit()
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
		return
	}
}

func (s *Server) sendErrorResponse(id int, code ErrorCodes, message string) {
	errResp := &ResponseError{
		Code:    int(code),
		Message: message,
	}
	s.sendResponse(id, nil, errResp)
	s.log.Printf("sent error: %s", message)
}

func (s *Server) notify(_ context.Context, method string, params any) {
	note := NotificationMessage{
		Message: Message{JSONRPC: "2.0"},
		Method:  method,
		Params:  &params,
	}
	encoded, err := s.EncodeMessage(note)
	if err != nil {
		s.log.Printf("Failed to encode notification: %v", err)
		return
	}
	_, err = s.writer.Write([]byte(encoded))
	if err != nil {
		s.log.Printf("Failed to write notification: %v", err)
		return
	}
	s.log.Printf("Sent notification '%s'", method)
}
