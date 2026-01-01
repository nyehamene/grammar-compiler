package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"grammar/check"
	grammar_log "grammar/log"
	"io"
	"os"
)

type Server struct {
	reader       *bufio.Reader
	writer       io.Writer
	logger       grammar_log.Logger
	shutdown     bool
	documents    map[DocumentUri]*document
	checker      *check.Checker
	fsFileLoader check.FileLoader
}

// document is an in-memory representation of a document.
type document struct {
	text []rune
}

func NewServer(in io.Reader, out io.Writer, logOut io.Writer) *Server {
	logger := NewLineLogger(logOut) // Use our new line logger
	srv := &Server{
		reader:    bufio.NewReader(in),
		writer:    out,
		logger:    logger,
		shutdown:  false,
		documents: make(map[DocumentUri]*document),
	}

	cu := check.NewCompilationUnit(srv, logger)
	checker := check.NewChecker(cu, logger)

	srv.checker = checker
	srv.fsFileLoader = &check.FileSystemFileLoader{}
	return srv
}

func (s *Server) GetDocumentContent(uri DocumentUri) (string, bool) {
	doc, ok := s.documents[uri]
	if !ok {
		return "", false
	}
	return string(doc.text), true
}

func (s *Server) Start() {
	s.logger.Print("LSP Server started")
	for {
		content, err := s.DecodeMessage(s.reader)
		if err == io.EOF {
			s.logger.Print("client disconnected!")
			return
		}
		if err != nil {
			s.logger.Printf("Failed to decode message: %v", err)
			continue
		}

		// Parse the raw JSON content into a generic map first to determine if it's a request or notification.
		var msg map[string]any
		if err := json.Unmarshal(content, &msg); err != nil {
			s.logger.Printf("Failed to unmarshal message: %v\n%s\n", err, content)
			// This is a parse error of the raw JSON, so we can't send a proper response.
			continue
		}

		// Now, try to unmarshal it into our specific RequestMessage for structured logging and handling.
		var reqMsg RequestMessage
		if err := json.Unmarshal(content, &reqMsg); err != nil {
			s.logger.Printf("Failed to unmarshal into RequestMessage: %v\n%s\n", err, content)
			// This should not happen if the first unmarshal succeeded and content is valid JSON.
			continue
		}

		if reqMsg.ID != nil { // It's a request
			s.handleRequest(*reqMsg.ID, msg)
		} else { // It's a notification
			s.handleNotification(msg)
		}
	}
}

func (s *Server) handleRequest(id int, rawMsg map[string]any) {
	s.logger.Print(rawMsg) // Log the full request message struct
	method, ok := rawMsg["method"].(string)
	if !ok {
		s.sendErrorResponse(id, InvalidRequest, "method not found", "unknown")
		return
	}
	switch method {
	case "initialize":
		handleInitializeRequest(s, id, rawMsg)
	case "textDocument/formatting":
		s.handleTextDocumentFormatting(id, rawMsg)
	case "shutdown":
		s.handleShutdown(id, rawMsg)
	case "textDocument/hover":
		s.handleHover(id, rawMsg)
	case "textDocument/completion":
		s.handleCompletion(id, rawMsg)
	case "textDocument/definition":
		s.handleDefinition(id, rawMsg)
	case "textDocument/references":
		s.handleReferences(id, rawMsg)
	case "textDocument/documentSymbol":
		s.handleDocumentSymbol(id, rawMsg)
	case "workspace/symbol":
		s.handleWorkspaceSymbol(id, rawMsg)
	case "textDocument/prepareRename":
		s.handlePrepareRename(id, rawMsg)
	case "textDocument/rename":
		s.handleRename(id, rawMsg)
	case "textDocument/diagnostic":
		s.handleDocumentDiagnostic(id, rawMsg)
	case "textDocument/documentLink":
		s.handleDocumentLink(id, rawMsg)
	case "textDocument/documentHighlight":
		handleDocumentHighlight(s, id, rawMsg)
	default:
		s.sendErrorResponse(id, MethodNotFound, fmt.Sprintf("unexpected method: %s", method), method)
	}
}

func (s *Server) handleShutdown(id int, rawMsg map[string]any) {
	s.shutdown = true
	method := "shutdown"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}
	s.sendResponse(id, method, nil, nil)
	s.logger.Printf("Shutdown request received. Server state: shutdown=%t", s.shutdown)
}

func (s *Server) handleNotification(rawMsg map[string]any) {
	s.logger.Print(rawMsg) // Log the full notification message struct
	method, ok := rawMsg["method"].(string)
	if !ok {
		s.logger.Printf("Notification method not found in raw message: %#v", rawMsg)
		return
	}

	ctx := context.Background()

	switch method {
	case "initialized":
		if err := s.handleInitialized(ctx, rawMsg); err != nil {
			s.logger.Printf("Failed to handle initialized: %v", err)
		}
	case "textDocument/didOpen":
		if err := s.handleDidOpen(ctx, rawMsg); err != nil {
			s.logger.Printf("Failed to handle textDocument/didOpen: %v", err)
		}
	case "textDocument/didChange":
		if err := s.handleDidChange(ctx, rawMsg); err != nil {
			s.logger.Printf("Failed to handle textDocument/didChange: %v", err)
		}
	case "textDocument/didClose":
		if err := s.handleDidClose(ctx, rawMsg); err != nil {
			s.logger.Printf("Failed to handle textDocument/didClose: %v", err)
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

func (s *Server) sendResponse(id int, method string, result any, errResp *ResponseError) {
	resp := ResponseMessage{
		Message: Message{JSONRPC: "2.0"},
		ID:      &id,
		Result:  &result,
		Error:   errResp,
	}
	s.logger.Print(resp) // Log the full response message struct

	encoded, err := s.EncodeMessage(resp)
	if err != nil {
		s.logger.Printf("Failed to encode response: %v", err)
		return
	}
	_, err = s.writer.Write([]byte(encoded))
	if err != nil {
		s.logger.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) sendErrorResponse(id int, code ErrorCodes, message string, method string) {
	errResp := &ResponseError{
		Code:    int(code),
		Message: message,
	}
	// The sendResponse will log the actual response message.
	s.sendResponse(id, method, nil, errResp)
}

func (s *Server) notify(_ context.Context, method string, params any) {
	note := NotificationMessage{
		Message: Message{JSONRPC: "2.0"},
		Method:  method,
		Params:  &params,
	}
	s.logger.Print(note) // Log the full notification struct

	encoded, err := s.EncodeMessage(note)
	if err != nil {
		s.logger.Printf("Failed to encode notification: %v", err)
		return
	}
	_, err = s.writer.Write([]byte(encoded))
	if err != nil {
		s.logger.Printf("Failed to write notification: %v", err)
	}
}
