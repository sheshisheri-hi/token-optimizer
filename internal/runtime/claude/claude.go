package claude

import (
	"context"

	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
)

type Runtime struct{}

func New() *Runtime { return &Runtime{} }

func (r *Runtime) ID() string          { return "claude" }
func (r *Runtime) DisplayName() string { return "Claude Code CLI" }

func (r *Runtime) DataHome() (string, error) {
	return "", runtime.ErrNotImplemented
}

func (r *Runtime) Setup(ctx context.Context, opts runtime.SetupOpts) (*runtime.SetupResult, error) {
	return nil, runtime.ErrNotImplemented
}

func (r *Runtime) Off(ctx context.Context, opts runtime.OffOpts) (*runtime.OffResult, error) {
	return nil, runtime.ErrNotImplemented
}

func (r *Runtime) Status(ctx context.Context) (*runtime.StatusReport, error) {
	return nil, runtime.ErrNotImplemented
}

func (r *Runtime) Doctor(ctx context.Context, opts runtime.DoctorOpts) (*runtime.DoctorReport, error) {
	return nil, runtime.ErrNotImplemented
}

func (r *Runtime) Probe(ctx context.Context) (*runtime.ProbeReport, error) {
	return nil, runtime.ErrNotImplemented
}

func (r *Runtime) HookCommand(event string) []string { return nil }
