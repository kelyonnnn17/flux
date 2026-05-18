package engine

import (
	"fmt"
)

type FFmpegAdapter struct {
	Runner CmdRunner
}

func (a *FFmpegAdapter) Convert(src, dst string, args []string) error {
	// Build command: ffmpeg -i src [args...] dst
	cmdArgs := []string{"-i", src}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, dst)
	ctx, cancel := engineContext()
	defer cancel()
	cmd := a.Runner.CommandContext(ctx, "ffmpeg", cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg error: %w, output: %s", err, string(out))
	}
	return nil
}
