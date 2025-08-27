package stream2sentence

import (
	"context"
)

// GenerateSentencesConfig holds configuration for the GenerateSentences function
type GenerateSentencesConfig struct {
	SentenceSplitterConfig
}

// GenerateSentences generates well-formed sentences from a stream of text chunks
// This is the main synchronous API function
func GenerateSentences(generator <-chan string, config GenerateSentencesConfig) <-chan string {
	splitter := NewSentenceSplitter(config.SentenceSplitterConfig)
	resultChan := make(chan string, 10)

	go func() {
		defer close(resultChan)

		// Process input chunks
		for chunk := range generator {
			splitter.Add(chunk)
			for sentence := range splitter.Stream() {
				resultChan <- sentence
			}
		}

		// Flush remaining sentences
		for sentence := range splitter.Flush() {
			resultChan <- sentence
		}
	}()

	return resultChan
}

// GenerateSentencesAsync generates sentences from an async stream with context support
func GenerateSentencesAsync(ctx context.Context, generator <-chan string, config GenerateSentencesConfig) <-chan string {
	splitter := NewSentenceSplitter(config.SentenceSplitterConfig)
	resultChan := make(chan string, 10)

	go func() {
		defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-generator:
				if !ok {
					for sentence := range splitter.Flush() {
						select {
						case <-ctx.Done():
							return
						case resultChan <- sentence:
						}
					}
					return
				}

				splitter.Add(chunk)
				for sentence := range splitter.Stream() {
					select {
					case <-ctx.Done():
						return
					case resultChan <- sentence:
					}
				}
			}
		}
	}()

	return resultChan
}

// GenerateSentencesFromSlice is a convenience function for processing a slice of text chunks
func GenerateSentencesFromSlice(chunks []string, config GenerateSentencesConfig) <-chan string {
	generator := make(chan string, len(chunks))

	go func() {
		defer close(generator)
		for _, chunk := range chunks {
			generator <- chunk
		}
	}()

	return GenerateSentences(generator, config)
}

// GenerateSentencesFromString is a convenience function for processing a single string
func GenerateSentencesFromString(text string, config GenerateSentencesConfig) <-chan string {
	return GenerateSentencesFromSlice([]string{text}, config)
}
