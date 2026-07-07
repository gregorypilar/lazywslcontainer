package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// WSLC is the wslc.exe client used by lazywslcontainer.
type WSLC struct {
	binaryPath string
}

// New returns a WSLC client pointing at the given wslc binary (defaults to "wslc").
func New(binaryPath string) *WSLC {
	if binaryPath == "" {
		binaryPath = "wslc"
	}
	return &WSLC{binaryPath: binaryPath}
}

// ErrNotFound is returned when wslc binary cannot be located.
var ErrNotFound = errors.New("wslc.exe not found on PATH; run `wsl --update` to install it")

func (w *WSLC) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, w.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return nil, fmt.Errorf("wslc %s: %w\nstderr: %s",
			strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// Version returns the wslc version string.
func (w *WSLC) Version(ctx context.Context) (string, error) {
	out, err := w.run(ctx, "version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Ping reports whether wslc is reachable. Best-effort: returns the version.
func (w *WSLC) Ping(ctx context.Context) (string, error) {
	return w.Version(ctx)
}

// runJSON runs a wslc subcommand and unmarshals JSON output (pulled via --format json
// where supported by wslc; falls back to nil when JSON parsing fails).
func (w *WSLC) runJSON(ctx context.Context, into any, args ...string) error {
	out, err := w.run(ctx, args...)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil
	}
	if jsonErr := json.Unmarshal(out, into); jsonErr != nil {
		// Some wslc commands emit text. Caller should fall back to text.
		return errTextOnly{out: out}
	}
	return nil
}

type errTextOnly struct{ out []byte }

func (e errTextOnly) Error() string { return "non-json output from wslc" }
func (e errTextOnly) Text() []byte  { return e.out }

// IsTextOnly reports whether err indicates wslc returned plain text (not JSON).
func IsTextOnly(err error) (bool, []byte) {
	var t errTextOnly
	if errors.As(err, &t) {
		return true, t.out
	}
	return false, nil
}

// Duration helper for context timeouts used by callers.
func withTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}
