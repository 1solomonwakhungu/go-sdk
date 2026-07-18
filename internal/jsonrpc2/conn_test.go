// Copyright 2026 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package jsonrpc2

import (
	"errors"
	"io"
	"testing"
)

// TestShuttingDownWrapsErrors verifies that errors returned from
// inFlightState.shuttingDown wrap both the closing error and the
// underlying read/write error, so that errors.Is works for both.
//
// Before the fix for #1098, the underlying error was formatted with
// %v instead of %w, which meant errors.Is(err, underlyingErr) always
// returned false. Consumers that relied on errors.Is(err, io.EOF) to
// detect a broken connection could not do so.
func TestShuttingDownWrapsErrors(t *testing.T) {
	tests := []struct {
		name      string
		state     inFlightState
		wantInner error
	}{
		{
			name:      "writeErr wrapped",
			state:     inFlightState{writeErr: io.EOF},
			wantInner: io.EOF,
		},
		{
			name:      "readErr wrapped",
			state:     inFlightState{readErr: io.ErrUnexpectedEOF},
			wantInner: io.ErrUnexpectedEOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.shuttingDown(ErrServerClosing)
			if err == nil {
				t.Fatal("shuttingDown returned nil, expected an error")
			}
			// The closing error should be matchable via errors.Is.
			if !errors.Is(err, ErrServerClosing) {
				t.Errorf("errors.Is(err, ErrServerClosing) = false, want true; err = %v", err)
			}
			// The underlying read/write error should be matchable.
			// Before the fix, %v was used instead of %w, so this
			// comparison would fail.
			if !errors.Is(err, tt.wantInner) {
				t.Errorf("errors.Is(err, %v) = false, want true; err = %v", tt.wantInner, err)
			}
		})
	}
}
