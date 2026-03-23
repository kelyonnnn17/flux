package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Meta struct {
	Key   string
	Value string
}

type Line struct {
	Kind string
	Text string
}

type Result struct {
	Meta      []Meta
	Lines     []Line
	Success   string
	ErrorHint string
}

type Spec struct {
	Command string
	Running string
	Run     func() (Result, error)
}

var (
	enabled    = true
	disableClr = false
)

func Configure(isEnabled bool, noColor bool) {
	enabled = isEnabled
	disableClr = noColor || os.Getenv("NO_COLOR") != ""
}

func Enabled() bool {
	return enabled
}

func ShouldRender() bool {
	if !enabled {
		return false
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	return true
}

type doneMsg struct {
	result Result
	err    error
}

type model struct {
	spinner spinner.Model
	spec    Spec
	done    bool
	result  Result
	err     error

	promptStyle  lipgloss.Style
	cmdStyle     lipgloss.Style
	dimStyle     lipgloss.Style
	cyanStyle    lipgloss.Style
	errStyle     lipgloss.Style
	okStyle      lipgloss.Style
	metaStyle    lipgloss.Style
	metaKeyStyle lipgloss.Style
	metaValStyle lipgloss.Style
}

func newModel(spec Spec) model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"-", "\\", "|", "/"},
		FPS:    110000000,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4"))

	m := model{
		spinner: s,
		spec:    spec,

		promptStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4")),
		cmdStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0")),
		dimStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		cyanStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4")),
		errStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#C0392B")),
		okStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4")).Bold(true),
		metaStyle:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#1A1A1A")).Padding(0, 1),
		metaKeyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Width(8),
		metaValStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0")),
	}

	if disableClr {
		base := lipgloss.NewStyle()
		m.promptStyle = base
		m.cmdStyle = base
		m.dimStyle = base
		m.cyanStyle = base
		m.errStyle = base
		m.okStyle = base
		m.metaStyle = base
		m.metaKeyStyle = base.Width(8)
		m.metaValStyle = base
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runTask(m.spec))
}

func runTask(spec Spec) tea.Cmd {
	return func() tea.Msg {
		res, err := spec.Run()
		return doneMsg{result: res, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case doneMsg:
		m.done = true
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	prompt := m.promptStyle.Render(">")
	cmdLine := m.cmdStyle.Render(m.spec.Command)

	if !m.done {
		running := m.dimStyle.Render(m.spec.Running)
		return fmt.Sprintf("\n  %s %s\n\n  %s %s\n", prompt, cmdLine, m.spinner.View(), running)
	}

	if m.err != nil {
		hint := m.result.ErrorHint
		if hint == "" {
			hint = "run --help for usage"
		}
		return fmt.Sprintf("\n  %s %s\n\n  X %s\n    %s\n", prompt, cmdLine, m.errStyle.Render(m.err.Error()), m.dimStyle.Render(hint))
	}

	metaRows := make([]string, 0, len(m.result.Meta))
	for _, row := range m.result.Meta {
		metaRows = append(metaRows, m.metaKeyStyle.Render(row.Key)+m.metaValStyle.Render(row.Value))
	}
	metaBlock := ""
	if len(metaRows) > 0 {
		metaBlock = m.metaStyle.Render(strings.Join(metaRows, "\n")) + "\n\n"
	}

	lineRows := make([]string, 0, len(m.result.Lines))
	for _, ln := range m.result.Lines {
		switch ln.Kind {
		case "ok":
			lineRows = append(lineRows, "  "+m.okStyle.Render("OK")+"  "+m.cmdStyle.Render(ln.Text))
		case "warn":
			lineRows = append(lineRows, "  !   "+m.dimStyle.Render(ln.Text))
		case "error":
			lineRows = append(lineRows, "  X   "+m.errStyle.Render(ln.Text))
		default:
			lineRows = append(lineRows, "  i   "+m.dimStyle.Render(ln.Text))
		}
	}
	linesBlock := ""
	if len(lineRows) > 0 {
		linesBlock = strings.Join(lineRows, "\n") + "\n\n"
	}

	success := m.result.Success
	if success == "" {
		success = "done"
	}

	return fmt.Sprintf("\n  %s %s\n\n%s%s  %s\n", prompt, cmdLine, metaBlock, linesBlock, m.cyanStyle.Render(success))
}

func Run(spec Spec) error {
	p := tea.NewProgram(newModel(spec))
	out, err := p.Run()
	if err != nil {
		return err
	}
	if m, ok := out.(model); ok {
		if m.err != nil {
			return m.err
		}
	}
	return nil
}
