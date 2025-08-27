package stream2sentence

import "strings"

// Conjunctions
var conjunctions = []string{"and", "or", "but", "so", "for", "nor", "yet"}

// Prepositions
var prepositions = []string{
	"in", "on", "at", "by", "with", "about", "of", "to", "for",
	"from", "as", "over", "under", "through", "between", "during", "there",
}

// Articles
var articles = []string{"a", "an", "the"}

// Possessives and Demonstratives
var possessivesAndDemonstratives = []string{
	"my", "your", "his", "her", "its", "our", "their",
	"this", "that", "these", "those",
}

// Auxiliary/Helping Verbs
var auxiliaryVerbs = []string{
	"is", "are", "was", "were", "am", "be", "been", "being",
	"do", "does", "did", "have", "has", "had",
	"can", "could", "shall", "should",
	"will", "would", "may", "might", "must",
}

// Pronouns
var pronouns = []string{
	"I", "we", "you", "he", "she", "it", "they",
	"who", "whom", "whose", "which", "that",
}

// Quantifiers
var quantifiers = []string{"some", "many", "few", "all", "any", "most", "much", "none", "several"}

// Adverbs that Modify Flow
var adverbs = []string{"very", "too", "just", "quite", "almost", "nearly", "only"}

// Interrogatives
var interrogatives = []string{"what", "where", "when", "why", "how"}

// Relative Pronouns
var relativePronouns = []string{"who", "which", "that"}

// Subordinating Conjunctions
var subordinatingConjunctions = []string{
	"although", "because", "if", "since",
	"though", "while", "until", "unless",
}

// AvoidPauseWords contains all words that should be avoided for pausing
var AvoidPauseWords map[string]bool

// initAvoidPauseWords initializes the AvoidPauseWords map
func initAvoidPauseWords() {
	if AvoidPauseWords != nil {
		return
	}

	allWords := []string{}
	allWords = append(allWords, conjunctions...)
	allWords = append(allWords, prepositions...)
	allWords = append(allWords, articles...)
	allWords = append(allWords, possessivesAndDemonstratives...)
	allWords = append(allWords, auxiliaryVerbs...)
	allWords = append(allWords, pronouns...)
	allWords = append(allWords, quantifiers...)
	allWords = append(allWords, adverbs...)
	allWords = append(allWords, interrogatives...)
	allWords = append(allWords, subordinatingConjunctions...)

	AvoidPauseWords = make(map[string]bool)
	for _, word := range allWords {
		AvoidPauseWords[word] = true
		AvoidPauseWords[strings.ToLower(word)] = true
	}
}

// IsAvoidPauseWord checks if a word should be avoided for pausing
func IsAvoidPauseWord(word string) bool {
	initAvoidPauseWords()
	return AvoidPauseWords[word] || AvoidPauseWords[strings.ToLower(word)]
}
