package server

import (
	"context"
	"encoding/json"
	"grammar/log"
)

type InitializedParams struct {
	// The initialized notification has no parameters.
}

func (s *Server) handleInitialized(ctx context.Context, rawMsg map[string]any) error {
	var params InitializedParams
	if rawMsg["params"] != nil { // Check if params exist before attempting to marshal
		encodedParams, err := json.Marshal(rawMsg["params"])
		if err != nil {
			return err
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			return err
		}
	}

	s.logger.Info("Initialized notification received. Client is ready.", log.Fields{})
	// In a real LSP server, you might trigger initial workspace scans,
	// send initial diagnostics, or register capabilities here.

	return nil
}
