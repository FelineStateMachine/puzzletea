package app

func (m model) persistCompletionIfSolved() model {
	newSessionController(&m).persistCompletionIfSolved()
	return m
}
