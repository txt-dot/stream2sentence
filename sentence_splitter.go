package stream2sentence

import (
	"container/list"
	"strings"
	"unicode"
)

// QuickYieldMode defines different modes for quick yielding
type QuickYieldMode int

const (
	NoQuickYield QuickYieldMode = iota
	QuickYieldFirstFragment
	QuickYieldAllFragments
)

// SentenceSplitter processes text streams and yields well-formed sentences
type SentenceSplitter struct {
	// Core configuration
	contextSize                int
	minimumSentenceLength      int
	minimumFirstFragmentLength int

	// Quick yield mode
	quickYieldMode QuickYieldMode

	// Text cleanup options
	cleanupOptions CleanupFlags

	// Delimiters
	sentenceFragmentDelimiters string
	fullSentenceDelimiters     string

	// Internal state
	inputBuffer           *list.List
	buffer                strings.Builder
	isFirstSentence       bool
	wordCount             int
	lastDelimiterPosition int

	// Character sets for fast lookup
	fragmentDelimiterSet map[rune]bool
	fullDelimiterSet     map[rune]bool
}

// SentenceSplitterConfig holds configuration options for SentenceSplitter
type SentenceSplitterConfig struct {
	ContextSize                int
	MinimumSentenceLength      int
	MinimumFirstFragmentLength int
	QuickYieldMode             QuickYieldMode
	CleanupOptions             CleanupFlags
	SentenceFragmentDelimiters string
	FullSentenceDelimiters     string
}

// DefaultConfig returns a default configuration for SentenceSplitter
func DefaultConfig() SentenceSplitterConfig {
	return SentenceSplitterConfig{
		ContextSize:                12,
		MinimumSentenceLength:      10,
		MinimumFirstFragmentLength: 10,
		QuickYieldMode:             QuickYieldAllFragments,
		CleanupOptions:             CleanupAll, // Default to just strip text
		SentenceFragmentDelimiters: ".?!;:,\n…)]}。-",
		FullSentenceDelimiters:     ".?!\n…。",
	}
}

// NewSentenceSplitter creates a new SentenceSplitter with the given configuration
func NewSentenceSplitter(config SentenceSplitterConfig) *SentenceSplitter {

	splitter := &SentenceSplitter{
		contextSize:                config.ContextSize,
		minimumSentenceLength:      config.MinimumSentenceLength,
		minimumFirstFragmentLength: config.MinimumFirstFragmentLength,
		quickYieldMode:             config.QuickYieldMode,
		cleanupOptions:             config.CleanupOptions,
		sentenceFragmentDelimiters: config.SentenceFragmentDelimiters,
		fullSentenceDelimiters:     config.FullSentenceDelimiters,

		// Initialize internal state
		inputBuffer:           list.New(),
		buffer:                strings.Builder{},
		isFirstSentence:       true,
		wordCount:             0,
		lastDelimiterPosition: -1,

		// Initialize delimiter sets
		fragmentDelimiterSet: make(map[rune]bool),
		fullDelimiterSet:     make(map[rune]bool),
	}

	// Populate delimiter sets for fast lookup
	for _, r := range config.SentenceFragmentDelimiters {
		splitter.fragmentDelimiterSet[r] = true
	}

	for _, r := range config.FullSentenceDelimiters {
		splitter.fullDelimiterSet[r] = true
	}

	return splitter
}

// Add adds a text chunk to the input buffer
func (s *SentenceSplitter) Add(chunk string) {
	s.inputBuffer.PushBack(chunk)
}

// Stream processes the input buffer and yields sentences
func (s *SentenceSplitter) Stream() <-chan string {
	resultChan := make(chan string, 10) // Buffered channel

	go func() {
		defer close(resultChan)

		for s.inputBuffer.Len() > 0 {
			element := s.inputBuffer.Front()
			s.inputBuffer.Remove(element)
			chunk := element.Value.(string)

			for _, char := range chunk {
				if char == 0 {
					continue
				}

				// Add character to buffer and trim left whitespace
				s.buffer.WriteRune(char)
				bufferStr := strings.TrimLeft(s.buffer.String(), " \t\n\r")
				s.buffer.Reset()
				s.buffer.WriteString(bufferStr)

				// Update word count on encountering space or sentence fragment delimiter
				if unicode.IsSpace(char) || s.fragmentDelimiterSet[char] {
					s.wordCount++
				}

				// Check conditions to yield first sentence fragment quickly
				if s.isFirstSentence &&
					s.buffer.Len() > s.minimumFirstFragmentLength &&
					(s.quickYieldMode == QuickYieldFirstFragment || s.quickYieldMode == QuickYieldAllFragments) {

					shouldYield := false

					if s.fragmentDelimiterSet[char] {
						shouldYield = true
					}

					if shouldYield {
						yieldText := CleanText(s.buffer.String(), s.cleanupOptions)
						resultChan <- yieldText
						s.buffer.Reset()
						s.wordCount = 0
						s.isFirstSentence = false
						continue
					}
				}

				// Continue accumulating characters if buffer is under minimum sentence length
				if s.buffer.Len() <= s.minimumSentenceLength+s.contextSize {
					continue
				}

				// Update last delimiter position if a new delimiter is found
				if s.fullDelimiterSet[char] {
					s.lastDelimiterPosition = s.buffer.Len() - 1
				}

				// Define context window for checking potential sentence boundaries
				contextWindowEndPos := s.buffer.Len() - s.contextSize - 1
				contextWindowStartPos := contextWindowEndPos - s.contextSize
				if contextWindowStartPos < 0 {
					contextWindowStartPos = 0
				}

				// Tokenize sentences from buffer using fragment delimiters
				sentences := TokenizeSentencesWithDelimiters(s.buffer.String(), s.sentenceFragmentDelimiters)

				// Combine sentences below minimum_sentence_length with the next sentence(s)
				combinedSentences := s.combineSentences(sentences)
				sentences = combinedSentences

				// Process and yield sentences based on conditions
				shouldProcess := len(sentences) > 2 ||
					(s.lastDelimiterPosition >= 0 &&
						contextWindowStartPos <= s.lastDelimiterPosition &&
						s.lastDelimiterPosition <= contextWindowEndPos)

				if shouldProcess && len(sentences) > 1 {
					totalLengthExceptLast := 0
					for i := 0; i < len(sentences)-1; i++ {
						totalLengthExceptLast += len(sentences[i])
					}

					if totalLengthExceptLast >= s.minimumSentenceLength {
						for i := 0; i < len(sentences)-1; i++ {
							yieldText := CleanText(sentences[i], s.cleanupOptions)
							resultChan <- yieldText
							s.wordCount = 0
						}

						if s.quickYieldMode == QuickYieldAllFragments {
							s.isFirstSentence = true
						}

						// Handle buffer ending with space
						endsWithSpace := strings.HasSuffix(s.buffer.String(), " ")
						s.buffer.Reset()
						s.buffer.WriteString(sentences[len(sentences)-1])
						if endsWithSpace {
							s.buffer.WriteString(" ")
						}

						// Reset the last delimiter position after yielding
						s.lastDelimiterPosition = -1
					}
				}
			}
		}
	}()

	return resultChan
}

// Flush yields remaining buffer as final sentence(s)
func (s *SentenceSplitter) Flush() <-chan string {
	resultChan := make(chan string, 10)

	go func() {
		defer close(resultChan)

		if s.buffer.Len() > 0 {
			sentences := TokenizeSentencesWithDelimiters(s.buffer.String(), s.sentenceFragmentDelimiters)
			sentenceBuffer := ""

			for _, sentence := range sentences {
				sentenceBuffer += sentence
				if len(sentenceBuffer) < s.minimumSentenceLength {
					sentenceBuffer += " "
					continue
				}

				yieldText := CleanText(sentenceBuffer, s.cleanupOptions)
				resultChan <- yieldText
				sentenceBuffer = ""
			}

			if sentenceBuffer != "" {
				yieldText := CleanText(sentenceBuffer, s.cleanupOptions)
				resultChan <- yieldText
			}
		}
	}()

	return resultChan
}

// combineSentences combines short sentences with following ones
func (s *SentenceSplitter) combineSentences(sentences []string) []string {
	if len(sentences) <= 1 {
		return sentences
	}

	var (
		combined     []string
		tempSentence string
	)

	for _, sentence := range sentences {
		if len(sentence) < s.minimumSentenceLength {
			tempSentence += sentence + " "
		} else {
			if tempSentence != "" {
				tempSentence += sentence
				combined = append(combined, strings.TrimSpace(tempSentence))
				tempSentence = ""
			} else {
				combined = append(combined, strings.TrimSpace(sentence))
			}
		}
	}

	// If there's a leftover temp_sentence that hasn't been appended
	if tempSentence != "" {
		combined = append(combined, strings.TrimSpace(tempSentence))
	}

	return combined
}
