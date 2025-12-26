package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// EncodeMessage encodes a JSON-RPC message into the LSP format (header + JSON content).
func EncodeMessage(msg any) (string, error) {
	content, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", len(content))
	return header + string(content), nil
}

// DecodeMessage reads an LSP message (headers + JSON content) from an io.Reader and decodes it.
func DecodeMessage(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)

	// Read headers
	headers := make(map[string]string)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}
		if len(line) == 0 { // Empty line separates headers from content
			break
		}

		parts := strings.SplitN(string(line), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header: %s", line)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Get Content-Length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("Content-Length header missing")
	}
	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, content); err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	return content, nil
}
