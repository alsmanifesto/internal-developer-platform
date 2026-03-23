// Package tui provides Bubble Tea TUI models for the scaffold CLI.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ravon/scaffold/internal/metadata"
)

// createStep enumerates each step in the create flow.
type createStep int

const (
	stepProjectName createStep = iota
	stepServiceType
	stepWorkload
	stepStack
	stepPipeline
	stepConfirm
	stepDone
)

// CreateModel is the Bubble Tea model for the scaffold create flow.
type CreateModel struct {
	step      createStep
	confirmed bool
	quitting  bool

	// Text input state
	nameInput string
	nameCursor bool

	// List selection state
	cursor int

	// Selections
	name        string
	serviceType string
	workload    string
	stack       string
	pipeline    string
}

// Option lists for each selection step
var (
	serviceTypeOptions = []string{"api", "worker", "job"}
	workloadOptions    = []string{"app", "data", "ml"}
	stackOptions       = []string{"go", "python", "spark", "kafka"}
	pipelineOptions    = []string{"gh-actions", "concourse", "airflow", "mlflow"}
	confirmOptions     = []string{"Yes", "No"}
)

// Lip Gloss styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F5A623")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	summaryKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	summaryValStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")).
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 3).
			MarginTop(1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			MarginTop(1)
)

// NewCreateModel creates a fresh CreateModel ready to start.
func NewCreateModel() CreateModel {
	return CreateModel{
		step: stepProjectName,
	}
}

// Confirmed returns true if the user confirmed creation.
func (m CreateModel) Confirmed() bool {
	return m.confirmed
}

// ServiceConfig returns the ServiceMetadata built from user selections.
func (m CreateModel) ServiceConfig() metadata.ServiceMetadata {
	return metadata.ServiceMetadata{
		Name:        m.name,
		ServiceType: m.serviceType,
		Workload:    m.workload,
		Stack:       m.stack,
		Pipeline:    m.pipeline,
		Path:        m.name,
		CreatedAt:   time.Now(),
	}
}

// Init implements tea.Model.
func (m CreateModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m CreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case stepProjectName:
			return m.updateNameInput(msg)
		case stepServiceType, stepWorkload, stepStack, stepPipeline, stepConfirm:
			return m.updateListSelect(msg)
		}
	}
	return m, nil
}

func (m CreateModel) updateNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if strings.TrimSpace(m.nameInput) != "" {
			m.name = strings.TrimSpace(m.nameInput)
			m.step = stepServiceType
			m.cursor = 0
		}
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.nameInput) > 0 {
			m.nameInput = m.nameInput[:len(m.nameInput)-1]
		}
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	default:
		if msg.Type == tea.KeyRunes {
			m.nameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m CreateModel) updateListSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.currentOptions()

	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(options)-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		selected := options[m.cursor]
		switch m.step {
		case stepServiceType:
			m.serviceType = selected
			m.step = stepWorkload
			m.cursor = 0
		case stepWorkload:
			m.workload = selected
			m.step = stepStack
			m.cursor = 0
		case stepStack:
			m.stack = selected
			m.step = stepPipeline
			m.cursor = 0
		case stepPipeline:
			m.pipeline = selected
			m.step = stepConfirm
			m.cursor = 0
		case stepConfirm:
			if selected == "Yes" {
				m.confirmed = true
			}
			m.step = stepDone
			return m, tea.Quit
		}
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m CreateModel) currentOptions() []string {
	switch m.step {
	case stepServiceType:
		return serviceTypeOptions
	case stepWorkload:
		return workloadOptions
	case stepStack:
		return stackOptions
	case stepPipeline:
		return pipelineOptions
	case stepConfirm:
		return confirmOptions
	}
	return nil
}

// View implements tea.Model.
func (m CreateModel) View() string {
	if m.quitting {
		return ""
	}
	if m.step == stepDone {
		return ""
	}

	var sb strings.Builder

	// Header
	sb.WriteString(headerStyle.Render("⚡ Scaffold — Service Platform"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Let's create a new service."))
	sb.WriteString("\n\n")

	switch m.step {
	case stepProjectName:
		sb.WriteString(m.renderNameInput())
	case stepServiceType:
		sb.WriteString(m.renderList(
			"How does this service run?",
			"",
			serviceTypeOptions,
		))
	case stepWorkload:
		sb.WriteString(m.renderList(
			"What is the primary workload?",
			"",
			workloadOptions,
		))
	case stepStack:
		sb.WriteString(m.renderList(
			"Select a technology stack",
			"",
			stackOptions,
		))
	case stepPipeline:
		sb.WriteString(m.renderList(
			"Select a pipeline tool",
			"",
			pipelineOptions,
		))
	case stepConfirm:
		sb.WriteString(m.renderConfirm())
	}

	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("↑/↓ navigate   enter select   esc quit"))
	sb.WriteString("\n")

	return borderStyle.Render(sb.String())
}

func (m CreateModel) renderNameInput() string {
	var sb strings.Builder
	sb.WriteString(promptStyle.Render("What is the name of your service?"))
	sb.WriteString("\n")
	sb.WriteString(subtitleStyle.Render("Must be unique within the platform"))
	sb.WriteString("\n\n")

	displayText := m.nameInput + "█"
	sb.WriteString(inputStyle.Render(displayText))
	sb.WriteString("\n\n")
	sb.WriteString(hintStyle.Render("type your service name, then press enter"))
	return sb.String()
}

func (m CreateModel) renderList(prompt, subtitle string, options []string) string {
	var sb strings.Builder
	sb.WriteString(promptStyle.Render(prompt))
	sb.WriteString("\n")
	if subtitle != "" {
		sb.WriteString(subtitleStyle.Render(subtitle))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for i, opt := range options {
		if i == m.cursor {
			sb.WriteString(selectedStyle.Render("  ● " + opt))
		} else {
			sb.WriteString(normalStyle.Render("  ○ " + opt))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m CreateModel) renderConfirm() string {
	var sb strings.Builder

	// Summary
	sb.WriteString(promptStyle.Render("Summary"))
	sb.WriteString("\n\n")

	rows := [][]string{
		{"Name", m.name},
		{"Service Type", m.serviceType},
		{"Workload", m.workload},
		{"Stack", m.stack},
		{"Pipeline", m.pipeline},
	}

	for _, row := range rows {
		key := fmt.Sprintf("  %-14s", row[0])
		sb.WriteString(summaryKeyStyle.Render(key))
		sb.WriteString(summaryValStyle.Render(row[1]))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(promptStyle.Render("Create this service?"))
	sb.WriteString("\n\n")

	for i, opt := range confirmOptions {
		if i == m.cursor {
			sb.WriteString(selectedStyle.Render("  ● " + opt))
		} else {
			sb.WriteString(normalStyle.Render("  ○ " + opt))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
