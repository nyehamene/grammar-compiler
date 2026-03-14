package server_test

import (
	"bytes"
	"context"
	"fmt"
	"grammar/log"
	"grammar/server"
	"strings"
	"testing"
	"time"
)

func TestInitializedNotification(t *testing.T) {
	// Sample initialized notification from a client
	message := `{
		"jsonrpc": "2.0",
		"method": "initialized",
		"params": {}
	}`

	contentLength := len(message)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", contentLength)

	var in bytes.Buffer
	var out strings.Builder
	var logOut bytes.Buffer

	srv := server.NewServer(&in, &out, log.NewConsoleLogger(&logOut, log.DEBUG))
	defer func() {
		if t.Failed() {
			t.Log(logOut.String())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	in.WriteString(header)
	in.WriteString(message)

	<-ctx.Done()

	response := out.String()

	if len(response) != 0 {
		t.Fatalf("Server produced an unexpected response for notification: %s", response)
	}
}
