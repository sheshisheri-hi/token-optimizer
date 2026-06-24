package runtime

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("runtime not implemented for this target")

type FeatureStatus string

const (
	StatusOK        FeatureStatus = "OK"
	StatusDegraded  FeatureStatus = "DEGRADED"
	StatusOff       FeatureStatus = "OFF"
	StatusBlocked   FeatureStatus = "BLOCKED"
)

type SetupOpts struct {
	Modules []string
	All     bool
	DryRun  bool
}

type SetupResult struct {
	Actions []string
}

type OffOpts struct {
	Modules []string
	All     bool
}

type OffResult struct {
	Actions []string
}

type DoctorOpts struct {
	JSON   bool
	Probe  bool
	Audit  bool
	Module string
}

type FeatureReport struct {
	Name   string
	Status FeatureStatus
	Detail string
}

type StatusReport struct {
	Target   string
	DataHome string
	Modules  []FeatureReport
	Features []FeatureReport
	Tally    string
}

type DoctorReport struct {
	StatusReport
	Warnings []string
}

type ProbeReport struct {
	ProbedAt string
	Features map[string]FeatureStatus
}

type Runtime interface {
	ID() string
	DisplayName() string
	DataHome() (string, error)

	Setup(ctx context.Context, opts SetupOpts) (*SetupResult, error)
	Off(ctx context.Context, opts OffOpts) (*OffResult, error)

	Status(ctx context.Context) (*StatusReport, error)
	Doctor(ctx context.Context, opts DoctorOpts) (*DoctorReport, error)
	Probe(ctx context.Context) (*ProbeReport, error)

	HookCommand(event string) []string
}
