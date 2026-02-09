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
	// Reduced horizontal padding to 0 to minimize sidebar margins as requested
	appStyle = lipgloss.NewStyle().Padding(1, 0, 0, 0)

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

	// Sidebar Styles
	// Added MarginTop(1) to align with Header
	leftSidebarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				MarginTop(1)

	rightSidebarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				MarginTop(1)
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
	ta.Prompt = ""
	ta.Focus() // Focus message input by default

	// Clear default styles to remove "light white-grey focus"
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

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

	// Sidebar Widths
	// Fixed content width of 20
	leftSidebarContentWidth := 20
	rightSidebarContentWidth := 20
	// Rendered width = Content + Border(2) + Padding(0) = 22
	leftSidebarRenderedWidth := leftSidebarContentWidth + 2
	rightSidebarRenderedWidth := rightSidebarContentWidth + 2

	// Calculate Center Column Width
	// Total Width - App Padding (0) - Sidebars
	// appStyle has padding (1,0,0,0) -> 0 horizontal padding
	centerRenderedWidth := m.width - leftSidebarRenderedWidth - rightSidebarRenderedWidth

	// Ensure min width to prevent crashes or weird rendering
	if centerRenderedWidth < 40 {
		centerRenderedWidth = 40
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

	// Header Container Width calculation
	// headerContainerStyle has border(2) + padding(2) = 4 extra width
	// We want the rendered header container to match centerRenderedWidth
	// So Content Width = centerRenderedWidth - 4
	headerContentWidth := centerRenderedWidth - 4

	// Search Bar Width Calculation
	targetTotalWidth := headerContentWidth - leftWidth - rightWidth
	// searchBaseStyle adds: 2 (padding L/R)
	searchContentWidth := targetTotalWidth - 2
	if searchContentWidth < 10 {
		searchContentWidth = 10
	}

	m.textInput.Width = searchContentWidth - 2

	// Dynamic style for search input
	var searchInputView string
	if m.textInput.Focused() {
		pinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		m.textInput.TextStyle = pinkStyle
		m.textInput.PromptStyle = pinkStyle
		searchInputView = searchBaseStyle.Width(searchContentWidth).Render(m.textInput.View())
	} else {
		grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		m.textInput.TextStyle = grayStyle
		m.textInput.PromptStyle = grayStyle
		searchInputView = searchBaseStyle.Width(searchContentWidth).Render(m.textInput.View())
	}

	headerContent := lipgloss.JoinHorizontal(lipgloss.Center, leftSide, searchInputView, rightSide)

	// Set width on container to ensure it fills space
	header := headerContainerStyle.Width(headerContentWidth).Render(headerContent)

	// --- 2. STATUS LINE ---
	// Status line has no border, just content.
	// Matches centerRenderedWidth.
	statusLine := statusLineStyle.
		Width(centerRenderedWidth).
		Render("MESSAGE-BUFFER")

	// --- 4. BOTTOM MESSAGE INPUT ---
	promptColor := lipgloss.Color("240")
	borderColor := lipgloss.Color("240")
	if m.messageInput.Focused() {
		promptColor = lipgloss.Color("212")
		borderColor = lipgloss.Color("212")
	}

	prompt := lipgloss.NewStyle().Foreground(promptColor).Render("> ")
	icons := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" \uee49 \U000F0066")

	// Input Box Width Logic
	// messageBox rendered width = centerRenderedWidth
	// messageBox has border(2) + padding(2) = 4 extra width
	// So Content Available Width = centerRenderedWidth - 4
	messageBoxContentWidth := centerRenderedWidth - 4

	// Textarea Width = Available - Prompt - Icons
	// We subtract an extra 2 for safety to prevent wrap
	inputWidth := messageBoxContentWidth - lipgloss.Width(prompt) - lipgloss.Width(icons) - 2
	if inputWidth < 1 {
		inputWidth = 1
	}
	m.messageInput.SetWidth(inputWidth)

	inputContent := lipgloss.JoinHorizontal(lipgloss.Top,
		prompt,
		m.messageInput.View(),
		icons,
	)

	currentMessageBoxStyle := messageBoxStyle.Copy().BorderForeground(borderColor)
	messageBox := currentMessageBoxStyle.
		Width(messageBoxContentWidth). // Sets content width
		Render(inputContent)

	// --- 3. MAIN CONTENT (Border Box) ---
	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusLine)
	messageH := lipgloss.Height(messageBox)

	// Total Available Height Calculation
	// m.height - appPadding Top(1)
	availableMainHeight := m.height - 1 - headerH - statusH - messageH
	if availableMainHeight < 0 {
		availableMainHeight = 0
	}

	mainContent := mainContentStyle.
		Width(centerRenderedWidth - 4). // Fix: Match Header content width (center - 4)
		Height(availableMainHeight).
		Render("")

	// Compose Center Column
	centerColumn := lipgloss.JoinVertical(lipgloss.Left,
		header,
		statusLine,
		mainContent,
		messageBox,
	)

	// --- SIDEBARS ---
	// Calculate sidebar height
	// Should match center column height OR full available height.
	// Center column total height = headerH + statusH + mainH + messageH
	// Since mainH fills space, Total = m.height - 1.
	sidebarHeight := m.height - 1

	// But sidebars have MarginTop(1).
	// So Content Height = Total - Margin(1) - Border(2) = Total - 3.
	sidebarContentHeight := sidebarHeight - 3
	if sidebarContentHeight < 0 {
		sidebarContentHeight = 0
	}

	leftSidebar := leftSidebarStyle.
		Width(leftSidebarContentWidth).
		Height(sidebarContentHeight).
		Render("")

	rightSidebar := rightSidebarStyle.
		Width(rightSidebarContentWidth).
		Height(sidebarContentHeight).
		Render("")

	// --- COMBINE COLUMNS ---
	finalView := lipgloss.JoinHorizontal(lipgloss.Top,
		leftSidebar,
		centerColumn,
		rightSidebar,
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
