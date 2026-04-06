package model

import (
	"github.com/charmbracelet/crush/internal/workspace"
)

// selectedLargeModel returns the currently selected large language model from
// the agent coordinator, if one exists.
func (m *UI) selectedLargeModel() *workspace.AgentModel {
	if m.com.Workspace.AgentIsReady() {
		model := m.com.Workspace.AgentModel()
		return &model
	}
	return nil
}

// landingView renders the landing page main area. Context information
// (working directory, model, LSP/MCP status) is shown in the sidebar.
func (m *UI) landingView() string {
	return m.landingSessions.View(m.layout.main.Dx(), m.layout.main.Dy())
}
