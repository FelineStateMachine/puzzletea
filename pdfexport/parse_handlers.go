package pdfexport

import "strings"

type markdownBodyParser interface {
	ParseMarkdownBody(bodyLines []string, path string, bodyStartLine int) (any, error)
}

type markdownBodyParserFunc func(bodyLines []string, path string, bodyStartLine int) (any, error)

func (f markdownBodyParserFunc) ParseMarkdownBody(bodyLines []string, path string, bodyStartLine int) (any, error) {
	return f(bodyLines, path, bodyStartLine)
}

var markdownBodyParsers = map[string]markdownBodyParser{
	normalizeGameTypeToken("nonogram"): markdownBodyParserFunc(func(bodyLines []string, path string, bodyStartLine int) (any, error) {
		return parseNonogramBody(bodyLines, path, bodyStartLine)
	}),
}

var defaultMarkdownBodyParser markdownBodyParser = markdownBodyParserFunc(func(bodyLines []string, path string, bodyStartLine int) (any, error) {
	return parseGridTableBody(bodyLines, path, bodyStartLine)
})

func lookupMarkdownBodyParser(category string) markdownBodyParser {
	normalized := normalizeGameTypeToken(strings.TrimSpace(category))
	if parser, ok := markdownBodyParsers[normalized]; ok {
		return parser
	}
	return defaultMarkdownBodyParser
}
