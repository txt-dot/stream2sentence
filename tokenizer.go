package stream2sentence

import (
	"slices"
	"strings"
	"unicode"
)

var (
	commonAbbreviations = map[string]bool{
		"Mr": true, "Mrs": true, "Ms": true, "Dr": true,
		"Prof": true, "Rev": true, "St": true,
		"Jr": true, "Sr": true, "Inc": true, "Ltd": true,
		"Corp": true, "Co": true, "Ave": true, "Blvd": true,
		"vs": true, "etc": true, "ie": true, "eg": true,
		"am": true, "pm": true,
	}
)

// TokenizeSentencesWithDelimiters splits text into sentences using provided delimiters
func TokenizeSentencesWithDelimiters(text string, delimiters string) []string {
	if text == "" {
		return []string{}
	}

	var (
		delimiterRunes  = []rune(delimiters)
		sentences       []string
		currentSentence = strings.Builder{}
	)

	runes := []rune(text)
	for i, r := range runes {
		currentSentence.WriteRune(r)

		if !slices.Contains(delimiterRunes, r) {
			continue
		}

		if isEndOfSentence(runes, i) {
			if sentence := strings.TrimSpace(currentSentence.String()); sentence != "" {
				sentences = append(sentences, sentence)
			}
			currentSentence.Reset()
		}
	}

	if sentence := strings.TrimSpace(currentSentence.String()); sentence != "" {
		sentences = append(sentences, sentence)
	}

	return sentences
}

// isEndOfSentence determines if a delimiter actually ends a sentence
func isEndOfSentence(runes []rune, pos int) bool {
	current := runes[pos]

	// Check if we're inside a table cell - if so, avoid sentence breaks
	if isInsideTableCell(runes, pos) {
		return false
	}

	// Periods need special handling: they might be part of abbreviations,
	// so we need to check context to determine if it's a real sentence end
	if current == '.' {
		// Check if this is a decimal number (digit before and after the period)
		if pos > 0 && pos < len(runes)-1 &&
			unicode.IsDigit(runes[pos-1]) && unicode.IsDigit(runes[pos+1]) {
			return false
		}

		if pos > 0 && isAbbreviation(runes, pos) {
			return false
		}

		nextNonSpace := findNextNonSpace(runes, pos+1)
		if nextNonSpace == -1 {
			return true
		}

		if unicode.IsUpper(runes[nextNonSpace]) {
			return true
		}

		if unicode.IsLower(runes[nextNonSpace]) {
			return false
		}
	}

	// Commas need special handling: they might be thousand separators in numbers
	if current == ',' {
		// Check if this is a thousand separator (digit before and after the comma)
		if pos > 0 && pos < len(runes)-1 &&
			unicode.IsDigit(runes[pos-1]) && unicode.IsDigit(runes[pos+1]) {
			return false
		}
	}

	return true
}

// isAbbreviation checks if the period is part of a common abbreviation
func isAbbreviation(runes []rune, dotPos int) bool {
	if dotPos < 2 {
		return false
	}

	// Look for common abbreviation patterns
	// This is a simplified check - in a full implementation you'd want
	// a comprehensive list of abbreviations
	start := dotPos - 1
	for start >= 0 && unicode.IsLetter(runes[start]) {
		start--
	}
	start++

	if start >= dotPos {
		return false
	}

	abbrev := string(runes[start:dotPos])
	return commonAbbreviations[abbrev] || commonAbbreviations[strings.ToLower(abbrev)]
}

// findNextNonSpace finds the next non-whitespace character
func findNextNonSpace(runes []rune, start int) int {
	for i := start; i < len(runes); i++ {
		if !unicode.IsSpace(runes[i]) {
			return i
		}
	}
	return -1
}

// isInsideTableCell checks if the current position is inside a table cell
// This function looks for pipe characters (|) that indicate table boundaries
func isInsideTableCell(runes []rune, pos int) bool {
	// Look backwards for the start of the line
	lineStart := pos
	for lineStart > 0 && runes[lineStart-1] != '\n' {
		lineStart--
	}

	// Look forwards for the end of the line
	lineEnd := pos
	for lineEnd < len(runes)-1 && runes[lineEnd+1] != '\n' {
		lineEnd++
	}

	// Extract the current line
	line := string(runes[lineStart : lineEnd+1])

	// Use the existing table detection logic from text_utils.go
	if detectTableLine(line) {
		// Find position within the line
		posInLine := pos - lineStart

		// Count pipes before and after the current position to see if we're between them
		pipesBefore := 0
		pipesAfter := 0

		for i := 0; i < posInLine; i++ {
			if runes[lineStart+i] == '|' {
				pipesBefore++
			}
		}

		for i := posInLine + 1; i <= lineEnd-lineStart; i++ {
			if runes[lineStart+i] == '|' {
				pipesAfter++
			}
		}

		// If we have pipes both before and after, we're likely inside a table cell
		return pipesBefore > 0 && pipesAfter > 0
	}

	return false
}
