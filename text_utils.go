package stream2sentence

import (
	"strings"
)

// CleanupFlags represents text cleanup options using bit operations
type CleanupFlags int

// Text cleanup flags using bit operations
const (
	CleanupLinks CleanupFlags = 1 << iota
	CleanupEmojis
	CleanupTable
	StripText

	CleanupAll   = CleanupLinks | CleanupEmojis | CleanupTable | StripText
	CleanupBasic = CleanupLinks | CleanupEmojis | StripText
)

// HasFlag checks if a specific flag is set
func (f CleanupFlags) HasFlag(flag CleanupFlags) bool {
	return f&flag != 0
}

// CleanText cleans the text based on the provided cleanup flags
func CleanText(text string, flags CleanupFlags) string {
	result := text

	// Handle table cleanup
	if flags.HasFlag(CleanupTable) {
		result = stripTableStructures(result)
	}

	// Handle link cleanup
	if flags.HasFlag(CleanupLinks) {
		result = stripHTTPURLs(result)
	}

	// Handle emoji cleanup
	if flags.HasFlag(CleanupEmojis) {
		result = stripEmojis(result)
	}

	// Handle text stripping
	if flags.HasFlag(StripText) {
		result = strings.TrimSpace(result)
	}

	return result
}

// stripHTTPURLs removes simple HTTP/HTTPS URLs using pattern matching
func stripHTTPURLs(text string) string {
	// Simple pattern to match URLs - not using regex to avoid dependency
	lines := strings.Split(text, " ")
	var cleanedWords []string

	for _, word := range lines {
		if !strings.HasPrefix(word, "http://") && !strings.HasPrefix(word, "https://") {
			cleanedWords = append(cleanedWords, word)
		}
	}

	return strings.Join(cleanedWords, " ")
}

// normalizeNewlines replaces multiple consecutive newlines with double newlines
// This replaces the multipleNewlines regex pattern
func normalizeNewlines(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	emptyLineCount := 0

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			emptyLineCount++
			if emptyLineCount <= 1 { // Allow at most 1 consecutive empty line (so max 2 newlines)
				result = append(result, line)
			}
		} else {
			emptyLineCount = 0
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// stripEmojis removes emojis from the input text
// This is a simplified version - for full emoji support, you might want to use a specialized library
func stripEmojis(text string) string {
	var result strings.Builder
	for _, r := range text {
		if !isEmojiRune(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// isEmojiRune checks if a rune is an emoji
func isEmojiRune(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x1F1E6 && r <= 0x1F1FF) || // Regional indicator symbols
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation selectors
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1F018 && r <= 0x1F270) || // Various other emoji ranges
		r == 0x200D
}

// Legacy compatibility functions - these maintain the same API as the original implementation

// CleanMarkdownText is an alias for CleanText optimized for markdown content
func CleanMarkdownText(text string, flags CleanupFlags) string {
	return CleanText(text, flags)
}

// Legacy compatibility: isTableLine function for backward compatibility with tests
// This function provides a simple heuristic for table line detection
func detectTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Count table-like characters
	pipeCount := strings.Count(trimmed, "|")
	tabCount := strings.Count(trimmed, "\t")
	plusCount := strings.Count(trimmed, "+")
	dashCount := strings.Count(trimmed, "-")
	equalsCount := strings.Count(trimmed, "=")

	lineLen := len(trimmed)

	// Pipe-separated tables (at least 2 pipes)
	// or
	// Tab-separated tables (at least 2 tabs)
	if pipeCount >= 2 || tabCount >= 2 {
		return true
	}

	// ASCII art table borders (mostly made of +, -, |, =)
	tableChars := pipeCount + plusCount + dashCount + equalsCount
	if lineLen > 0 && float64(tableChars)/float64(lineLen) > 0.5 {
		return true
	}

	// Table separators (lines with mostly dashes, equals, or underscores)
	// Simple check without regex
	if lineLen > 3 {
		nonTableChars := 0
		for _, r := range trimmed {
			if r != '-' && r != '=' && r != '_' && r != '+' && r != '|' && r != ' ' && r != '\t' {
				nonTableChars++
			}
		}
		if nonTableChars == 0 {
			return true
		}
	}

	// Markdown table format detection (simplified)
	if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
		return true
	}

	return false
}

// stripTableStructures removes table-like structures using simple pattern matching
func stripTableStructures(text string) string {
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		if detectTableLine(line) {
			// Replace table line with empty line, but avoid consecutive empty lines
			if len(cleanedLines) > 0 && cleanedLines[len(cleanedLines)-1] != "" {
				cleanedLines = append(cleanedLines, "")
			}
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// Clean up multiple consecutive empty lines
	return normalizeNewlines(strings.Join(cleanedLines, "\n"))
}
