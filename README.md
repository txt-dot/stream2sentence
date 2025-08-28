# Stream2Sentence Go Implementation

Real-time processing and delivery of sentences from a continuous stream of characters or text chunks, implemented in Go.

This is a Go port of the Python [stream2sentence](https://github.com/KoljaB/stream2sentence) library.

## Features

- Real-time sentence detection from character streams
- Multiple tokenization strategies 
- Time-based sentence generation for TTS
- Configurable text cleanup
- Async processing with context support
- Concurrent-safe design using Go channels

## Installation

```bash
go get https://github.com/KoljaB/stream2sentence
```

## Quick Start

### Non-streaming Input (Simple Text Processing)

```go
package main

import (
    "fmt"
    "stream2sentence"
)

func main() {
    text := "This is a sample sentence. Here's another one! And yet another sentence? Finally, the last sentence."
    
    config := stream2sentence.GenerateSentencesConfig{
        SentenceSplitterConfig: stream2sentence.DefaultConfig(),
    }
    
    for sentence := range stream2sentence.GenerateSentencesFromString(text, config) {
        fmt.Printf("Generated: %q\n", sentence)
    }
}
```

### Streaming Input (Real-time Processing)

```go
package main

import (
    "fmt"
    "time"
    "stream2sentence"
)

func main() {
    // Create a channel to simulate streaming input
    textStream := make(chan string)
    
    config := stream2sentence.GenerateSentencesConfig{
        SentenceSplitterConfig: stream2sentence.DefaultConfig(),
    }
    
    // Start processing the stream
    sentences := stream2sentence.GenerateSentences(textStream, config)
    
    // Simulate streaming text chunks
    go func() {
        defer close(textStream)
        chunks := []string{
            "This is ", "a streaming ", "sentence. ",
            "Here comes ", "another one! ",
            "And yet ", "another sentence? ",
            "Finally, ", "the last ", "sentence.",
        }
        
        for _, chunk := range chunks {
            textStream <- chunk
            time.Sleep(100 * time.Millisecond) // Simulate delay
        }
    }()
    
    // Process generated sentences
    for sentence := range sentences {
        fmt.Printf("Generated: %q\n", sentence)
    }
}
```