package server

import (
	"context"
	"encoding/json"
)

type InitializedParams struct {
	// The initialized notification has no parameters.
}

func (s *Server) handleInitialized(ctx context.Context, msg map[string]any) error {
	var params InitializedParams
	if p, ok := msg["params"]; ok {
		encodedParams, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			return err
		}
	}

	s.log.Printf("Initialized notification received. Client is ready.")
	// In a real LSP server, you might trigger initial workspace scans,
	// send initial diagnostics, or register capabilities here.

	return nil
}
