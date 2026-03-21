package engine

import (
    "context"
    "os/exec"
)

type CmdRunner interface {
    CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd
}

type defaultRunner struct{}

func (d *defaultRunner) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
    return exec.CommandContext(ctx, name, arg...)
}

func NewDefaultRunner() CmdRunner {
    return &defaultRunner{}
}
