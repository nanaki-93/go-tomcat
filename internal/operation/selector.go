package operation

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nanaki-93/go-tomcat/internal/model"
)

type acquirerChoiceModel struct {
	choices  []string
	cursor   int
	selected string
	quit     bool
	err      error
}

func newAcquirerChoiceModel(acquirerMap map[string]model.Acquirer) acquirerChoiceModel {
	return acquirerChoiceModel{
		choices: GetOrderedKeys(acquirerMap),
		cursor:  0,
	}
}

func (m acquirerChoiceModel) Init() tea.Cmd {
	return nil
}

func (m acquirerChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quit = true
			m.selected = ""
			return m, tea.Quit
		case "enter":
			if len(m.choices) > 0 {
				m.selected = m.choices[m.cursor]
			}
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m acquirerChoiceModel) View() string {
	if len(m.choices) == 0 {
		return "No acquirers found.\n"
	}

	var b strings.Builder
	b.WriteString("Select acquirer (↑/↓, enter to confirm, q to quit):\n\n")
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor
		}
		fmt.Fprintf(&b, "%s %s\n", cursor, choice)
	}
	return b.String()
}

func selectAcquirerWithBubbleTea(acquirerMap map[string]model.Acquirer) (string, error) {
	if len(acquirerMap) == 0 {
		return "", fmt.Errorf("no acquirers available to select")
	}

	p := tea.NewProgram(newAcquirerChoiceModel(acquirerMap))
	modelChoose, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("bubble Tea program failed: %w", err)
	}

	m, ok := modelChoose.(acquirerChoiceModel)
	if !ok {
		return "", fmt.Errorf("unexpected Bubble Tea model type")
	}
	if m.selected == "" {
		return "", fmt.Errorf("acquirer selection canceled")
	}
	return m.selected, nil
}
