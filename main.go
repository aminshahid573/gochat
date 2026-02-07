package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI Styles
var (
	// Removed bottom padding to eliminate gap below text field
	appStyle = lipgloss.NewStyle().Padding(1, 2, 0, 2)

	// Header Styles
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			MarginRight(1).
			SetString("\uf489") // 

	channelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			MarginRight(1).
			SetString("#general")

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginRight(1).
			SetString("|")

	topicStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")). // Grey
			MarginRight(1).                    // Reduced margin to fit new divider
			SetString("TOPIC: Discussion")

	searchBaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

	iconBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginLeft(1).
			Align(lipgloss.Center)

	// The big wrapper for everything
	headerContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				MarginTop(1)

	// Status Line Style (Re-added)
	statusLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("212")).
			Padding(0, 1)

	// Main Content Area Style (Empty Border Box)
	mainContentStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1)

	// Message Input Box Style
	messageBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")). // Pink border
			Padding(0, 1).
			MarginTop(0)
)

type model struct {
	width        int
	height       int
	textInput    textinput.Model // Search bar
	messageInput textarea.Model  // Message input
}

func initialModel() model {
	// Search Input
	ti := textinput.New()
	ti.Placeholder = "Search"
	ti.Prompt = "\uf002 " // 
	ti.CharLimit = 156
	ti.Width = 20
	// Style for search input
	color240 := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ti.PromptStyle = color240
	ti.PlaceholderStyle = color240
	ti.TextStyle = color240
	ti.Cursor.Style = color240

	// Message Input (Textarea)
	ta := textarea.New()
	ta.Placeholder = "Type a Message or command (use / for actions)"
	ta.ShowLineNumbers = false
	ta.SetHeight(1)

	// We'll handle the prompt manually in the View
	ta.Prompt = ""
	ta.Focus() // Focus message input by default

	return model{
		textInput:    ti,
		messageInput: ta,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("Bubble Tea TUI"),
		textinput.Blink,
		textarea.Blink,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.messageInput.Focus()
			} else {
				m.messageInput.Blur()
				m.textInput.Focus()
			}
		case "enter":
			if m.messageInput.Focused() {
				// Handle dynamic height expansion
				// If currently 1 line, allow expansion to 2.
				// If 2 lines, submit (or stay at 2 if just typing)
				// The textarea handles inserting the newline in the content.
				// We just need to react to the new content size.
				// However, bubbletea's textarea usually requires explicit height setting.
				// We'll let the default update happen first to insert the char, then check line count.
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update inputs
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	m.messageInput, cmd = m.messageInput.Update(msg)
	cmds = append(cmds, cmd)

	// Post-update Logic for dynamic height
	if m.messageInput.LineCount() > 1 {
		m.messageInput.SetHeight(2)
	} else {
		m.messageInput.SetHeight(1)
	}
	// Limit to 2 lines max visible
	if m.messageInput.Height() > 2 {
		m.messageInput.SetHeight(2)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// --- 1. HEADER ---
	leftSide := lipgloss.JoinHorizontal(lipgloss.Center,
		logoStyle.String(),
		channelStyle.String(),
		dividerStyle.String(),
		topicStyle.String(),
		dividerStyle.String(),
	)
	leftWidth := lipgloss.Width(leftSide)

	bellIcon := iconBoxStyle.Render("\uf0f3") // 
	infoIcon := iconBoxStyle.Render("\uf05a") // 
	rightSide := lipgloss.JoinHorizontal(lipgloss.Center, bellIcon, infoIcon)
	rightWidth := lipgloss.Width(rightSide)

	availableWidth := m.width - 8
	targetTotalWidth := availableWidth - leftWidth - rightWidth
	searchContentWidth := targetTotalWidth - 2
	if searchContentWidth < 10 {
		searchContentWidth = 10
	}

	m.textInput.Width = searchContentWidth - 2

	// Dynamic style for search input
	var searchInputView string
	if m.textInput.Focused() {
		// Pink when focused
		pinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		m.textInput.TextStyle = pinkStyle
		m.textInput.PromptStyle = pinkStyle
		// Use rendered view directly
		searchInputView = searchBaseStyle.Width(searchContentWidth).Render(m.textInput.View())
	} else {
		// Gray when blurred (default)
		grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		m.textInput.TextStyle = grayStyle
		m.textInput.PromptStyle = grayStyle
		searchInputView = searchBaseStyle.Width(searchContentWidth).Render(m.textInput.View())
	}

	headerContent := lipgloss.JoinHorizontal(lipgloss.Center, leftSide, searchInputView, rightSide)
	headerWidth := m.width - 8
	header := headerContainerStyle.Width(headerWidth).Render(headerContent)

	// --- 2. STATUS LINE ---
	statusLine := statusLineStyle.
		Width(m.width - 4). // Match full width
		Render("MESSAGE-BUFFER")

	// --- 4. BOTTOM MESSAGE INPUT ---
	// Render prompt and icons separately
	promptColor := lipgloss.Color("240") // Default gray
	borderColor := lipgloss.Color("240") // Default gray
	if m.messageInput.Focused() {
		promptColor = lipgloss.Color("212") // Pink
		borderColor = lipgloss.Color("212") // Pink
	}

	prompt := lipgloss.NewStyle().Foreground(promptColor).Render("> ")
	icons := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" \uee49 \U000F0066") //  󰁦

	// Calculate width for the textarea proper
	inputWidth := headerWidth - lipgloss.Width(prompt) - lipgloss.Width(icons) - 4
	m.messageInput.SetWidth(inputWidth)

	// Create a horizontal layout for the message box content
	inputContent := lipgloss.JoinHorizontal(lipgloss.Top,
		prompt,
		m.messageInput.View(),
		icons,
	)

	// Apply dynamic border color
	currentMessageBoxStyle := messageBoxStyle.Copy().BorderForeground(borderColor)
	messageBox := currentMessageBoxStyle.
		Width(headerWidth).
		Render(inputContent)

	// --- 3. MAIN CONTENT (Border Box) ---
	// Calculate available height
	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusLine)
	messageH := lipgloss.Height(messageBox)

	// Total height - Components - App Padding (top 1 + bottom 0 = 1) - Extra Margin (1 from header)
	// We adjusted App Padding to (1, 2, 0, 2), so total vertical padding is 1.
	// Header margin top is 1.
	availableHeight := m.height - headerH - statusH - messageH - 1 - 1
	if availableHeight < 0 {
		availableHeight = 0
	}

	mainContent := mainContentStyle.
		Width(headerWidth).
		Height(availableHeight).
		Render("")

	// --- COMBINE ALL ---
	finalView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		statusLine,
		mainContent,
		messageBox,
	)

	return appStyle.Render(finalView)
}

func main() {
	m := initialModel()
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
