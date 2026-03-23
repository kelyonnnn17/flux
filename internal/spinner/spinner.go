package spinner

import (
	"github.com/briandowns/spinner"
	"time"
)

type Spinner struct {
	s *spinner.Spinner
}

func New(message string) *Spinner {
	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Suffix = " " + message
	return &Spinner{s: sp}
}

func (sp *Spinner) Start() { sp.s.Start() }
func (sp *Spinner) Stop()  { sp.s.Stop() }
