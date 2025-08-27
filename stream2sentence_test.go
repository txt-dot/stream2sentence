package stream2sentence

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function to create character-by-character generator
func createCharacterGenerator(text string) <-chan string {
	generator := make(chan string, len(text))
	go func() {
		defer close(generator)
		for _, char := range text {
			generator <- string(char)
		}
	}()
	return generator
}

// Helper function to create word-by-word generator
func createWordGenerator(text string) <-chan string {
	words := strings.Fields(text)
	generator := make(chan string, len(words))
	go func() {
		defer close(generator)
		for i, word := range words {
			if i == len(words)-1 {
				generator <- word
			} else {
				generator <- word + " "
			}
		}
	}()
	return generator
}

// Helper function to collect all sentences from a channel
func collectSentences(sentenceChan <-chan string) []string {
	var sentences []string
	for sentence := range sentenceChan {
		if trimmed := strings.TrimSpace(sentence); trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}
	return sentences
}

// === Basic Functionality Tests ===

func TestBasicSentenceGeneration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Simple sentences",
			input: "This is a sentence. And here's another! Yet, there's more. This ends now.",
			expected: []string{
				"This is a sentence.",
				"And here's another!",
				"Yet, there's more.",
				"This ends now.",
			},
		},
		{
			name:  "Mixed punctuation",
			input: "Question? Exclamation! Statement. Another question?",
			expected: []string{
				"Question? Exclamation!",
				"Statement.",
				"Another question?",
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: nil,
		},
		{
			name:  "Single sentence",
			input: "Only one sentence here.",
			expected: []string{
				"Only one sentence here.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GenerateSentencesConfig{
				SentenceSplitterConfig: DefaultConfig(),
			}
			generator := GenerateSentencesFromString(tt.input, config)
			sentences := collectSentences(generator)
			assert.Equal(t, tt.expected, sentences)
		})
	}
}

func TestCharacterStreamingInput(t *testing.T) {
	text := "Character by character. This should work perfectly!"

	generator := createCharacterGenerator(text)
	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := collectSentences(GenerateSentences(generator, config))

	expected := []string{
		"Character by character.",
		"This should work perfectly!",
	}

	assert.Equal(t, expected, sentences)
}

func TestWordStreamingInput(t *testing.T) {
	text := "Word by word streaming. Each word sent separately."

	generator := createWordGenerator(text)
	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := collectSentences(GenerateSentences(generator, config))

	expected := []string{
		"Word by word streaming.",
		"Each word sent separately.",
	}

	assert.Equal(t, expected, sentences)
}

// === Configuration Tests ===

func TestMinimumSentenceLength(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		minLength      int
		expectedMinLen int
	}{
		{
			name:           "Short sentences combined",
			input:          "Hi. How are you? I am fine. Thanks for asking.",
			minLength:      15,
			expectedMinLen: 15,
		},
		{
			name:           "Very short minimum",
			input:          "A. B. C. D.",
			minLength:      5,
			expectedMinLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.MinimumSentenceLength = tt.minLength

			genConfig := GenerateSentencesConfig{
				SentenceSplitterConfig: config,
			}

			generator := GenerateSentencesFromString(tt.input, genConfig)
			sentences := collectSentences(generator)

			// All sentences except possibly the last should meet minimum length
			for i, sentence := range sentences {
				if i < len(sentences)-1 { // Not the last sentence
					assert.GreaterOrEqual(t, len(sentence), tt.expectedMinLen,
						"Sentence %d too short: %s", i, sentence)
				}
			}
		})
	}
}

func TestQuickYieldModes(t *testing.T) {
	text := "This is a quick sentence fragment, and it should be yielded quickly."

	tests := []struct {
		name             string
		mode             QuickYieldMode
		minFirstLength   int
		expectQuickYield bool
	}{
		{
			name:             "No quick yield",
			mode:             NoQuickYield,
			minFirstLength:   10,
			expectQuickYield: false,
		},
		{
			name:             "Quick yield first fragment",
			mode:             QuickYieldFirstFragment,
			minFirstLength:   10,
			expectQuickYield: true,
		},
		{
			name:             "Quick yield all fragments",
			mode:             QuickYieldAllFragments,
			minFirstLength:   10,
			expectQuickYield: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.QuickYieldMode = tt.mode
			config.MinimumFirstFragmentLength = tt.minFirstLength

			genConfig := GenerateSentencesConfig{
				SentenceSplitterConfig: config,
			}

			generator := createCharacterGenerator(text)
			sentences := collectSentences(GenerateSentences(generator, genConfig))

			assert.True(t, len(sentences) > 0)
			if tt.expectQuickYield {
				assert.True(t, len(sentences[0]) >= tt.minFirstLength)
			}
		})
	}
}

func TestContextSize(t *testing.T) {
	text := "Test context. Window sizing. Should work. Properly now."

	tests := []struct {
		name        string
		contextSize int
	}{
		{"Small context", 5},
		{"Medium context", 12},
		{"Large context", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.ContextSize = tt.contextSize

			genConfig := GenerateSentencesConfig{
				SentenceSplitterConfig: config,
			}

			generator := GenerateSentencesFromString(text, genConfig)
			sentences := collectSentences(generator)

			// Should produce some sentences regardless of context size
			assert.True(t, len(sentences) > 0)
		})
	}
}

// === Text Cleanup Tests ===

func TestCleanupLinks(t *testing.T) {
	text := "Visit https://example.com for more info. Also check http://test.org!"

	config := DefaultConfig()
	config.CleanupOptions = CleanupLinks

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	result := strings.Join(sentences, " ")
	assert.NotContains(t, result, "https://example.com")
	assert.NotContains(t, result, "http://test.org")
	assert.Contains(t, result, "Visit")
	assert.Contains(t, result, "for more info")
}

func TestCleanupEmojis(t *testing.T) {
	text := "Hello 😀 world! This is great 🎉 stuff."

	config := DefaultConfig()
	config.CleanupOptions = CleanupEmojis

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	result := strings.Join(sentences, " ")
	assert.NotContains(t, result, "😀")
	assert.NotContains(t, result, "🎉")
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "world")
}

func TestCleanupTable(t *testing.T) {
	text := `Introduction text here.

| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |

Conclusion text here.`

	config := DefaultConfig()
	config.CleanupOptions = CleanupTable

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	result := strings.Join(sentences, " ")
	assert.Contains(t, result, "Introduction text here")
	assert.Contains(t, result, "Conclusion text here")
	assert.NotContains(t, result, "| Name |")
	assert.NotContains(t, result, "|------|")
}

func TestMultipleCleanupOptions(t *testing.T) {
	text := "Check https://example.com! 😀 Also see this table:\n| A | B |\n|---|---|\n| 1 | 2 |"

	config := DefaultConfig()
	config.CleanupOptions = CleanupLinks | CleanupEmojis | CleanupTable

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	result := strings.Join(sentences, " ")
	assert.NotContains(t, result, "https://example.com")
	assert.NotContains(t, result, "😀")
	assert.NotContains(t, result, "| A |")
	assert.Contains(t, result, "Check")
	assert.Contains(t, result, "Also see this table")
}

// === Delimiter Tests ===

func TestCustomDelimiters(t *testing.T) {
	text := "First part; Second part: Third part, Fourth part."

	config := DefaultConfig()
	config.SentenceFragmentDelimiters = ";:,"
	config.FullSentenceDelimiters = "."

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	// Should split on custom delimiters
	assert.True(t, len(sentences) > 1)
}

func TestChineseText(t *testing.T) {
	text := "这是第一句话。这是第二句话！还有第三句话？最后一句话。"

	config := DefaultConfig()

	genConfig := GenerateSentencesConfig{
		SentenceSplitterConfig: config,
	}

	generator := GenerateSentencesFromString(text, genConfig)
	sentences := collectSentences(generator)

	expected := []string{
		"这是第一句话。",
		"这是第二句话！",
		"还有第三句话？",
		"最后一句话。",
	}

	assert.Equal(t, expected, sentences)
}

// === Async and Context Tests ===

func TestAsyncGeneration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	generator := make(chan string, 3)
	go func() {
		defer close(generator)
		generator <- "First sentence. "
		time.Sleep(50 * time.Millisecond)
		generator <- "Second sentence. "
		time.Sleep(50 * time.Millisecond)
		generator <- "Third sentence."
	}()

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := collectSentences(GenerateSentencesAsync(ctx, generator, config))
	assert.True(t, len(sentences) >= 2)
}

func TestAsyncCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	generator := make(chan string)
	go func() {
		defer close(generator)
		for i := 0; i < 100; i++ {
			generator <- fmt.Sprintf("Sentence %d. ", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := GenerateSentencesAsync(ctx, generator, config)

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	results := collectSentences(sentences)

	// Should have been cancelled before processing all 100 sentences
	assert.True(t, len(results) < 50)
}

// === Tokenizer Tests ===

func TestTokenizerWithAbbreviations(t *testing.T) {
	text := "Dr. Smith went to U.S.A. He met Mr. Johnson."

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := collectSentences(GenerateSentencesFromString(text, config))

	// Should not split on abbreviations
	assert.True(t, len(sentences) > 0)
	// The first sentence should contain the abbreviations intact
	firstSentence := sentences[0]
	assert.Contains(t, firstSentence, "Dr. Smith")
	assert.Contains(t, firstSentence, "U.S.A.")
}

func TestTokenizerWithNumbers(t *testing.T) {
	text := "The price is $12.50 per item. Total came to 1,234.56 dollars."

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	sentences := collectSentences(GenerateSentencesFromString(text, config))

	expected := []string{
		"The price is $12.50 per item.",
		"Total came to 1,234.56 dollars.",
	}

	assert.Equal(t, expected, sentences)
}

func TestTokenizerEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Multiple spaces",
			input:    "First   sentence.    Second     sentence.",
			expected: []string{"First   sentence.", "Second     sentence."},
		},
		{
			name:     "Leading/trailing whitespace",
			input:    "   First sentence.   Second sentence.   ",
			expected: []string{"First sentence.", "Second sentence."},
		},
		{
			name:     "Mixed line endings",
			input:    "First sentence.\nSecond sentence.\r\nThird sentence.",
			expected: []string{"First sentence.", "Second sentence.", "Third sentence."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GenerateSentencesConfig{
				SentenceSplitterConfig: DefaultConfig(),
			}

			sentences := collectSentences(GenerateSentencesFromString(tt.input, config))
			assert.Equal(t, tt.expected, sentences)
		})
	}
}

// === Sentence Splitter Reuse Tests ===

func TestSentenceSplitterReuse(t *testing.T) {
	config := DefaultConfig()
	splitter := NewSentenceSplitter(config)

	// First use
	splitter.Add("First test. ")
	splitter.Add("Second sentence.")

	var firstResults []string
	for sentence := range splitter.Stream() {
		firstResults = append(firstResults, sentence)
	}
	for sentence := range splitter.Flush() {
		firstResults = append(firstResults, sentence)
	}

	assert.True(t, len(firstResults) > 0)

	// Second use - should work with a fresh splitter
	splitter2 := NewSentenceSplitter(config)
	splitter2.Add("Third test. ")
	splitter2.Add("Fourth sentence.")

	var secondResults []string
	for sentence := range splitter2.Stream() {
		secondResults = append(secondResults, sentence)
	}
	for sentence := range splitter2.Flush() {
		secondResults = append(secondResults, sentence)
	}

	assert.True(t, len(secondResults) > 0)
}

// === Utility Function Tests ===

func TestAvoidPauseWords(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"and", true},
		{"the", true},
		{"is", true},
		{"elephant", false},
		{"unusual", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := IsAvoidPauseWord(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDelimiterIgnorePrefixes(t *testing.T) {
	tests := []struct {
		prefix   string
		expected bool
	}{
		{"Dr.", true},
		{"Mr.", true},
		{"U.S.", true},
		{"elephant.", false},
		{"random.", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			result := IsDelimiterIgnorePrefix(tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// === Stress and Performance Tests ===

func TestLargeTextProcessing(t *testing.T) {
	// Create a large text with many sentences
	var textBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		textBuilder.WriteString(fmt.Sprintf("This is sentence number %d. ", i))
	}
	text := textBuilder.String()

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	start := time.Now()
	sentences := collectSentences(GenerateSentencesFromString(text, config))
	duration := time.Since(start)

	assert.Equal(t, 1000, len(sentences))
	assert.True(t, duration < 5*time.Second, "Processing took too long: %v", duration)
}

func TestConcurrentProcessing(t *testing.T) {
	const numGoroutines = 10
	const textPerGoroutine = "Concurrent test. Multiple goroutines. Should work fine."

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	results := make(chan []string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			sentences := collectSentences(GenerateSentencesFromString(textPerGoroutine, config))
			results <- sentences
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		sentences := <-results
		assert.Equal(t, 3, len(sentences))
		assert.Equal(t, "Concurrent test.", sentences[0])
		assert.Equal(t, "Multiple goroutines.", sentences[1])
		assert.Equal(t, "Should work fine.", sentences[2])
	}
}

// === Benchmark Tests ===

func BenchmarkSentenceGeneration(b *testing.B) {
	text := strings.Repeat("This is a benchmark sentence. ", 100)
	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator := GenerateSentencesFromString(text, config)
		for range generator {
			// Consume all sentences
		}
	}
}

func BenchmarkCharacterStreaming(b *testing.B) {
	text := "This is a benchmark for character streaming. It should be reasonably fast."
	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator := createCharacterGenerator(text)
		sentences := GenerateSentences(generator, config)
		for range sentences {
			// Consume all sentences
		}
	}
}

func BenchmarkLargeTextProcessing(b *testing.B) {
	var textBuilder strings.Builder
	for i := 0; i < 10000; i++ {
		textBuilder.WriteString("This is a sentence for benchmarking purposes. ")
	}
	text := textBuilder.String()

	config := GenerateSentencesConfig{
		SentenceSplitterConfig: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator := GenerateSentencesFromString(text, config)
		for range generator {
			// Consume all sentences
		}
	}
}
