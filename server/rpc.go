package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// EncodeMessage encodes a JSON-RPC message into the LSP format (header + JSON content).
func (s *Server) EncodeMessage(msg any) (string, error) {
	content, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", len(content))
	return header + string(content), nil
}

// DecodeMessage reads an LSP message (headers + JSON content) from an io.Reader and decodes it.
func (s *Server) DecodeMessage(r io.Reader) ([]byte, error) {
	br := bufio.NewReader(r)
	headers := map[string]string{}

	var contentLength int

	// --- Read headers ---
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line -> headers done
			break
		}

		// s.log.Printf("received header: %s\n", line)

		parts := strings.Split(line, ": ")
		if len(parts) != 2 {
			log.Printf("invalid header: %s\n", line)
			continue
		}

		header := strings.ToLower(parts[0])
		value := parts[1]
		headers[header] = value
	}

	// Parse "Content-Length: N"
	if v, ok := headers["content-length"]; ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %w", err)
		}
		contentLength = n
	} else {
		return nil, errors.New("missing Content-Length header")
	}

	// --- Read body ---
	body := make([]byte, contentLength)
	_, err := io.ReadFull(br, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
