package engine

import (
	"fmt"
	"os/exec"
)

type ImageMagickAdapter struct {
	Runner CmdRunner
}

func (a *ImageMagickAdapter) Convert(src, dst string, args []string) error {
	// Choose binary: prefer "magick" if available, otherwise "convert"
	binary := "magick"
	if _, err := exec.LookPath("magick"); err != nil {
		binary = "convert"
	}
	// Build args: src [args...] dst
	cmdArgs := []string{src}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, dst)
	ctx, cancel := engineContext()
	defer cancel()
	cmd := a.Runner.CommandContext(ctx, binary, cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("imagemagick error: %w, output: %s", err, string(out))
	}
	return nil
}
