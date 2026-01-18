package tokenizer

import (
	"regexp"
	"strings"
	"unicode"
)

// Tokenizer interface for text tokenization
type Tokenizer interface {
	Tokenize(text string) []string
}

// SimpleTokenizer implements basic word-level tokenization
type SimpleTokenizer struct {
	lowercase   bool
	removePunct bool
}

// NewSimpleTokenizer creates a new simple tokenizer
func NewSimpleTokenizer(lowercase, removePunct bool) *SimpleTokenizer {
	return &SimpleTokenizer{
		lowercase:   lowercase,
		removePunct: removePunct,
	}
}

// Tokenize splits text into tokens
func (t *SimpleTokenizer) Tokenize(text string) []string {
	// Convert to lowercase if enabled
	if t.lowercase {
		text = strings.ToLower(text)
	}

	// Remove punctuation if enabled
	if t.removePunct {
		text = removePunctuation(text)
	}

	// Split on whitespace
	tokens := strings.Fields(text)

	return tokens
}

// removePunctuation removes punctuation from text
func removePunctuation(text string) string {
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return result.String()
}

// RegexTokenizer uses regular expressions for tokenization
type RegexTokenizer struct {
	pattern *regexp.Regexp
}

// NewRegexTokenizer creates a new regex-based tokenizer
func NewRegexTokenizer(pattern string) (*RegexTokenizer, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexTokenizer{pattern: re}, nil
}

// DefaultRegexTokenizer creates a tokenizer with default pattern
func DefaultRegexTokenizer() *RegexTokenizer {
	// Default pattern matches words (including Unicode)
	re := regexp.MustCompile(`\w+`)
	return &RegexTokenizer{pattern: re}
}

// Tokenize splits text using regex pattern
func (t *RegexTokenizer) Tokenize(text string) []string {
	matches := t.pattern.FindAllString(text, -1)
	return matches
}

// NgramTokenizer creates n-gram tokens
type NgramTokenizer struct {
	n        int
	delegate Tokenizer
}

// NewNgramTokenizer creates a new n-gram tokenizer
func NewNgramTokenizer(n int, delegate Tokenizer) *NgramTokenizer {
	return &NgramTokenizer{
		n:        n,
		delegate: delegate,
	}
}

// Tokenize splits text into n-grams
func (t *NgramTokenizer) Tokenize(text string) []string {
	tokens := t.delegate.Tokenize(text)

	if t.n <= 1 || len(tokens) < t.n {
		return tokens
	}

	ngrams := make([]string, 0, len(tokens)-t.n+1)
	for i := 0; i <= len(tokens)-t.n; i++ {
		ngram := strings.Join(tokens[i:i+t.n], " ")
		ngrams = append(ngrams, ngram)
	}

	return ngrams
}

// ChineseTokenizer handles Chinese text tokenization
type ChineseTokenizer struct {
	simple *SimpleTokenizer
}

// NewChineseTokenizer creates a tokenizer for Chinese text
func NewChineseTokenizer() *ChineseTokenizer {
	return &ChineseTokenizer{
		simple: NewSimpleTokenizer(true, true),
	}
}

// Tokenize splits Chinese text into characters and words
func (t *ChineseTokenizer) Tokenize(text string) []string {
	tokens := make([]string, 0)

	// For Chinese, tokenize by characters and spaces
	currentWord := ""
	for _, r := range text {
		if isChineseChar(r) {
			// Add current word if exists
			if currentWord != "" {
				tokens = append(tokens, currentWord)
				currentWord = ""
			}
			// Add individual Chinese characters as tokens
			tokens = append(tokens, string(r))
		} else if unicode.IsSpace(r) {
			if currentWord != "" {
				tokens = append(tokens, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(r)
		}
	}

	if currentWord != "" {
		tokens = append(tokens, currentWord)
	}

	return tokens
}

// isChineseChar checks if a rune is a Chinese character
func isChineseChar(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
