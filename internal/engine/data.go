package engine

import (
	"github.com/kelyonnnn17/flux/internal/data"
)

// DataAdapter implements Engine for data format conversions (CSV, JSON, YAML, TOML).
// It does not use CmdRunner; fromFormat and toFormat are passed via args:
// args[0]=from, args[1]=to (empty means infer from path).
type DataAdapter struct {
	Runner CmdRunner // unused; required for Engine interface consistency
}

func (a *DataAdapter) Convert(src, dst string, args []string) error {
	fromFormat, toFormat := "", ""
	if len(args) >= 2 {
		fromFormat, toFormat = args[0], args[1]
	}
	return data.Convert(src, dst, fromFormat, toFormat)
}
