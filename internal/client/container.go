package client

import (
	"context"
	"fmt"
	"time"
)

type ContainerState int

const (
	ContainerStateCreated ContainerState = iota + 1
	ContainerStateRunning
	ContainerStatePaused
	ContainerStateExited
	ContainerStateUnknown
)

func (s ContainerState) String() string {
	switch s {
	case ContainerStateCreated:
		return "created"
	case ContainerStateRunning:
		return "running"
	case ContainerStatePaused:
		return "paused"
	case ContainerStateExited:
		return "exited"
	default:
		return "unknown"
	}
}

type ContainerPort struct {
	BindingAddress string `json:"BindingAddress"`
	ContainerPort  int    `json:"ContainerPort"`
	HostPort       int    `json:"HostPort"`
	Protocol       int    `json:"Protocol"`
}

func (p ContainerPort) String() string {
	proto := "tcp"
	if p.Protocol == 17 {
		proto = "udp"
	}
	return fmt.Sprintf("%s:%d->%d/%s", p.BindingAddress, p.HostPort, p.ContainerPort, proto)
}

type Container struct {
	ID             string          `json:"Id"`
	Name           string          `json:"Name"`
	Image          string          `json:"Image"`
	State          ContainerState  `json:"State"`
	CreatedAt      int64           `json:"CreatedAt"`
	StateChangedAt int64           `json:"StateChangedAt"`
	Ports          []ContainerPort `json:"Ports"`
}

// Containers calls `wslc container list --all --format json`. Returns
// parsed containers when JSON is available; otherwise an empty list with err
// (caller may retry text parsing).
func (w *WSLC) Containers(ctx context.Context, all bool) ([]Container, error) {
	args := []string{"container", "list", "--format", "json"}
	if all {
		args = append(args, "--all")
	}
	var cs []Container
	err := w.runJSON(ctx, &cs, args...)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// ContainerLogs returns logs for the given container id/name.
func (w *WSLC) ContainerLogs(ctx context.Context, id string, since time.Duration) ([]byte, error) {
	args := []string{"container", "logs", id}
	if since > 0 {
		args = append(args, "--since", fmt.Sprintf("%d", time.Now().Add(-since).Unix()))
	}
	return w.run(ctx, args...)
}

// ContainerInspect returns raw inspect output for the given container.
func (w *WSLC) ContainerInspect(ctx context.Context, id string) ([]byte, error) {
	return w.run(ctx, "container", "inspect", id)
}

// ContainerStop stops a container by id or name.
func (w *WSLC) ContainerStop(ctx context.Context, id string) error {
	_, err := w.run(ctx, "container", "stop", id)
	return err
}

// ContainerStart starts a container by id or name.
func (w *WSLC) ContainerStart(ctx context.Context, id string) error {
	_, err := w.run(ctx, "container", "start", id)
	return err
}

// ContainerRestart restarts a container by id or name.
func (w *WSLC) ContainerRestart(ctx context.Context, id string) error {
	_, err := w.run(ctx, "container", "restart", id)
	return err
}

// ContainerRemove removes a container by id or name.
func (w *WSLC) ContainerRemove(ctx context.Context, id string, force bool) error {
	args := []string{"container", "rm", id}
	if force {
		args = append(args, "--force")
	}
	_, err := w.run(ctx, args...)
	return err
}

// ContainerExec runs an arbitrary command inside the given container and
// returns combined stdout. Useful for attaching / one-off commands.
func (w *WSLC) ContainerExec(ctx context.Context, id string, cmd ...string) ([]byte, error) {
	args := append([]string{"exec", id}, cmd...)
	return w.run(ctx, args...)
}

// ContainerPrune removes all stopped containers.
func (w *WSLC) ContainerPrune(ctx context.Context) ([]byte, error) {
	return w.run(ctx, "container", "prune")
}

// RunOptions configures a `wslc run` invocation.
type RunOptions struct {
	Name   string
	Ports  []string
	Detach bool
	Remove bool
	Cmd    []string
}

// Run starts a container from an image. Returns the container id (when
// detached) or combined output (otherwise).
func (w *WSLC) Run(ctx context.Context, image string, opts RunOptions) ([]byte, error) {
	args := []string{"run"}
	if opts.Remove {
		args = append(args, "--rm")
	}
	if opts.Detach {
		args = append(args, "-d")
	}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	for _, p := range opts.Ports {
		args = append(args, "-p", p)
	}
	args = append(args, image)
	args = append(args, opts.Cmd...)
	return w.run(ctx, args...)
}

func (c Container) PortList() []string {
	if len(c.Ports) == 0 {
		return nil
	}
	out := make([]string, 0, len(c.Ports))
	for _, p := range c.Ports {
		out = append(out, p.String())
	}
	return out
}

func (c Container) StateString() string {
	return c.State.String()
}

func (c Container) CreatedTime() time.Time {
	return time.Unix(c.CreatedAt, 0)
}
