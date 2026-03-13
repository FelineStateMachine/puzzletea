package app

import (
	"fmt"
	"log"

	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

type noticeLevel string

const noticeLevelError noticeLevel = "error"

type noticeState struct {
	level   noticeLevel
	message string
}

func (m model) clearNotice() model {
	m.notice = noticeState{}
	return m
}

func (m model) setErrorf(format string, args ...any) model {
	message := fmt.Sprintf(format, args...)
	log.Printf("%s", message)
	m.notice = noticeState{
		level:   noticeLevelError,
		message: message,
	}
	return m
}

func (m model) appendNotice(content string) string {
	return appendNoticeContent(m.width, m.notice, content)
}

func appendNoticeContent(width int, notice noticeState, content string) string {
	if notice.message == "" {
		return content
	}

	style := ui.ErrorNoticeStyle()
	block := style.
		MaxWidth(max(width-8, 24)).
		Render(notice.message)
	if content == "" {
		return block
	}
	return lipgloss.JoinVertical(lipgloss.Left, content, "", block)
}

func (m model) renderPanel(title, content, footer string) string {
	return renderPanelView(m.width, m.height, m.notice, title, content, footer)
}

func centerContentWithNotice(width, height int, notice noticeState, content string) string {
	return ui.CenterView(width, height, appendNoticeContent(width, notice, content))
}

func renderPanelView(width, height int, notice noticeState, title, content, footer string) string {
	panel := ui.Panel(title, appendNoticeContent(width, notice, content), footer)
	return ui.CenterView(width, height, panel)
}
