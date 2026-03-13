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
	if m.notice.message == "" {
		return content
	}

	style := ui.ErrorNoticeStyle()
	block := style.
		MaxWidth(max(m.width-8, 24)).
		Render(m.notice.message)
	if content == "" {
		return block
	}
	return lipgloss.JoinVertical(lipgloss.Left, content, "", block)
}

func (m model) renderPanel(title, content, footer string) string {
	panel := ui.Panel(title, m.appendNotice(content), footer)
	return ui.CenterView(m.width, m.height, panel)
}
