package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// newProgram 统一创建 TUI Program。
// 当前各个页面仍然是独立 Program，必须依赖 AltScreen 才能避免普通终端缓冲里的重复渲染。
// 真正要同时解决“闪屏”和“重复”，需要后续把 goani tui 重构成单一 Program 的多页面路由。
func newProgram(model tea.Model) *tea.Program {
	return tea.NewProgram(model, tea.WithAltScreen())
}
