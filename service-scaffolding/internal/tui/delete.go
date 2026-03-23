package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ravon/scaffold/internal/metadata"
)

// deleteStep enumerates the steps in the delete flow.
type deleteStep int

const (
	deleteStepList deleteStep = iota
	deleteStepConfirm
	deleteStepDone
)

// DeleteModel is the Bubble Tea model for the scaffold delete flow.
type DeleteModel struct {
	step            deleteStep
	services        []metadata.ServiceMetadata
	cursor          int // confirm step cursor (0=Yes, 1=No)
	listCursor      int // service list cursor
	confirmed       bool
	quitting        bool
}

// Delete-specific styles
var (
	deleteHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6B6B")).
				MarginBottom(1)

	deleteSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true)

	deleteNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC"))

	deleteDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	deleteBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF6B6B")).
				Padding(1, 3).
				MarginTop(1)
)

// NewDeleteModel creates a DeleteModel pre-populated with the given services.
func NewDeleteModel(services []metadata.ServiceMetadata) DeleteModel {
	return DeleteModel{
		step:     deleteStepList,
		services: services,
	}
}

// Confirmed returns true if the user confirmed deletion.
func (m DeleteModel) Confirmed() bool {
	return m.confirmed
}

// SelectedService returns the name of the service selected for deletion.
func (m DeleteModel) SelectedService() string {
	if len(m.services) == 0 {
		return ""
	}
	return m.services[m.listCursor].Name
}

// Init implements tea.Model.
func (m DeleteModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m DeleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case deleteStepList:
			return m.updateList(msg)
		case deleteStepConfirm:
			return m.updateConfirm(msg)
		}
	}
	return m, nil
}

func (m DeleteModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.listCursor > 0 {
			m.listCursor--
		}
	case tea.KeyDown:
		if m.listCursor < len(m.services)-1 {
			m.listCursor++
		}
	case tea.KeyEnter:
		m.step = deleteStepConfirm
		m.cursor = 1 // Default to "No" for safety
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m DeleteModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < 1 {
			m.cursor++
		}
	case tea.KeyEnter:
		if m.cursor == 0 {
			m.confirmed = true
		}
		m.step = deleteStepDone
		return m, tea.Quit
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// View implements tea.Model.
func (m DeleteModel) View() string {
	if m.quitting || m.step == deleteStepDone {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(deleteHeaderStyle.Render("⚡ Scaffold — Service Platform"))
	sb.WriteString("\n")
	sb.WriteString(deleteDimStyle.Render("Delete an existing service."))
	sb.WriteString("\n\n")

	switch m.step {
	case deleteStepList:
		sb.WriteString(m.renderServiceList())
	case deleteStepConfirm:
		sb.WriteString(m.renderDeleteConfirm())
	}

	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("↑/↓ navigate   enter select   esc quit"))
	sb.WriteString("\n")

	return deleteBorderStyle.Render(sb.String())
}

func (m DeleteModel) renderServiceList() string {
	var sb strings.Builder

	sb.WriteString(promptStyle.Render("Select a service to delete"))
	sb.WriteString("\n\n")

	// Column header
	header := fmt.Sprintf("  %-30s %-12s %-10s", "NAME", "TYPE", "STACK")
	sb.WriteString(deleteDimStyle.Render(header))
	sb.WriteString("\n")
	sb.WriteString(deleteDimStyle.Render("  " + strings.Repeat("─", 52)))
	sb.WriteString("\n")

	for i, svc := range m.services {
		row := fmt.Sprintf("%-30s %-12s %-10s", svc.Name, svc.ServiceType, svc.Stack)
		if i == m.listCursor {
			sb.WriteString(deleteSelectedStyle.Render("● " + row))
		} else {
			sb.WriteString(deleteNormalStyle.Render("○ " + row))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m DeleteModel) renderDeleteConfirm() string {
	var sb strings.Builder

	var svcName string
	if len(m.services) > 0 {
		svcName = m.services[m.listCursor].Name
	}

	sb.WriteString(promptStyle.Render(fmt.Sprintf("Are you sure you want to delete %q?", svcName)))
	sb.WriteString("\n")
	sb.WriteString(deleteDimStyle.Render("This action cannot be undone."))
	sb.WriteString("\n\n")

	deleteConfirmOptions := []string{"Yes", "No"}
	for i, opt := range deleteConfirmOptions {
		if i == m.cursor {
			sb.WriteString(deleteSelectedStyle.Render("  ● " + opt))
		} else {
			sb.WriteString(deleteNormalStyle.Render("  ○ " + opt))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
