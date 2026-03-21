package engine

import (
    "context"
    "fmt"
    "time"
)

type PandocAdapter struct {
    Runner CmdRunner
}

func (a *PandocAdapter) Convert(src, dst string, args []string) error {
    // pandoc -o dst src [args...]
    cmdArgs := []string{"-o", dst, src}
    cmdArgs = append(cmdArgs, args...)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    cmd := a.Runner.CommandContext(ctx, "pandoc", cmdArgs...)
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("pandoc error: %w, output: %s", err, string(out))
    }
    return nil
}
