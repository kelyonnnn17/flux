package format

import (
	"os"
	"sync"

	"github.com/fatih/color"
)

var (
	noColor     bool
	noColorOnce sync.Once
)

// Init must be called before using formatters. Pass true to disable colors.
func Init(disable bool) {
	noColorOnce.Do(func() {
		noColor = disable || os.Getenv("NO_COLOR") != ""
		if noColor {
			color.NoColor = true
		}
	})
}

// Success prints success message (bright white, OK prefix).
func Success(msg string, args ...interface{}) {
	c := color.New(color.FgHiWhite)
	c.Printf("OK "+msg+"\n", args...)
}

// Error prints error message (red, X prefix).
func Error(msg string, args ...interface{}) {
	c := color.New(color.FgRed)
	c.Printf("X "+msg+"\n", args...)
}

// Info prints secondary/info message (dim grey).
func Info(msg string, args ...interface{}) {
	c := color.New(color.FgWhite)
	c.Add(color.Faint)
	c.Printf(msg+"\n", args...)
}

// Primary prints primary text (muted cyan).
func Primary(msg string, args ...interface{}) {
	c := color.New(color.FgCyan)
	c.Printf(msg+"\n", args...)
}
