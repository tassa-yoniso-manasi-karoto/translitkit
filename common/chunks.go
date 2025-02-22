package common

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// Default splitter used by NewChunkifier
var DefaultSplitter = "𓃰"

// SplitFunc defines the signature of a method used to split a string into tokens.
type SplitFunc func(string) []string

// SplitMethod represents a single splitting strategy (SplitFn)
// combined with a string joiner used when recombining tokens.
type SplitMethod struct {
	SplitFn SplitFunc
	Joiner  string
}

// Chunkifier is the main type that orchestrates chunk splitting so that the string
// passed to the initial provider is guaranteed to be within this provider's input limits.
type Chunkifier struct {
	// SplitMethods is the sequence of splitting strategies this chunkifier applies in order.
	SplitMethods []SplitMethod

	// Splitter is used specifically by SplitOnSplitter. You can set it to any marker 
	// string that you want to preserve in your tokens.
	Splitter string

	// MaxLength is a default maximum chunk size.
	MaxLength int
}

// NewChunkifier creates a chunkifier initialized with default fields:
// some default splitting methods, a default splitter, and zero for MaxLength (unbounded).
func NewChunkifier(max int) *Chunkifier {
	c := &Chunkifier{
		Splitter: DefaultSplitter,
		MaxLength: max,
	}
	// Build a default set of split methods:
	c.SplitMethods = []SplitMethod{
		{SplitFn: c.SplitSpace, Joiner: " "},
		{SplitFn: c.SplitSentences, Joiner: " "},
		{SplitFn: c.SplitOnSplitter, Joiner: ""},
		// too problematic with writing systems that don't use spaces =
		// if there is no word delimitations found it will behave like splitGraphemes
		//{SplitFn: c.SplitWords, Joiner: ""},
		
		// risk of truncating words
		// {SplitFn: c.SplitGraphemes, Joiner: ""},
	}
	return c
}

// Chunkify takes the given string s and a max length. The function tries each
// of c.SplitMethods in turn, splitting the string and then greedily combining tokens into chunks.
func (c *Chunkifier) Chunkify(s string) ([]string, error) {
	// If a negative max was passed or if the entire string already fits
	if c.MaxLength <= 0 || utf8.RuneCountInString(s) <= c.MaxLength {
		return []string{s}, nil
	}

	for _, method := range c.SplitMethods {
		tokens := method.SplitFn(s)
		// Debug / inspection print
		// fmt.Printf("DEBUG: tokens from method %v: %#v\n", method, tokens)

		if !tokensAreWithinLimit(tokens, c.MaxLength) {
			// If this split creates tokens bigger than max, skip it
			continue
		}
		combined := combineTokens(tokens, method.Joiner, c.MaxLength)
		if combined != nil {
			return combined, nil
		}
	}

	return nil, fmt.Errorf("could not decompose string into smaller parts: %q", s)
}

// --- Utility functions ---

// tokensAreWithinLimit checks that each token is within the allowed length.
func tokensAreWithinLimit(tokens []string, max int) bool {
	if max <= 0 {
		return true
	}
	for _, token := range tokens {
		if utf8.RuneCountInString(token) > max {
			return false
		}
	}
	return true
}

// combineTokens greedily merges tokens with the specified joiner
// without exceeding the max length (if max > 0).
func combineTokens(tokens []string, joiner string, max int) []string {
	var result []string
	var current string

	for i, token := range tokens {
		if current == "" {
			current = token
			continue
		}
		candidate := current + joiner + token
		if max <= 0 || utf8.RuneCountInString(candidate) <= max {
			current = candidate
		} else {
			result = append(result, current)
			current = token
		}
		if i == len(tokens)-1 {
			result = append(result, current)
		}
	}
	if current != "" && (len(result) == 0 || result[len(result)-1] != current) {
		result = append(result, current)
	}
	// Verify the final result doesn’t exceed max in any chunk
	if max > 0 {
		for _, chunk := range result {
			if utf8.RuneCountInString(chunk) > max {
				return nil
			}
		}
	}
	return result
}

// --- The splitting methods, now as *Chunkifier methods ---

// SplitSpace splits the input string into tokens that include both words and spaces. 
// Every space character is treated as its own token.
func (c *Chunkifier) SplitSpace(str string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range str {
		if r == ' ' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// SplitSentences uses uniseg to split the text into sentences.
func (c *Chunkifier) SplitSentences(text string) (splitted []string) {
	if len(text) == 0 {
		return nil
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		sentence, rest, newState := uniseg.FirstSentenceInString(remaining, state)
		// Debug
		// fmt.Printf("DEBUG: uniseg.FirstSentenceInString => sentence=%q, rest=%q\n", sentence, rest)
		if sentence != "" {
			splitted = append(splitted, sentence)
		}
		remaining = rest
		state = newState
	}
	return splitted
}

// SplitOnSplitter splits the text using c.Splitter. The splitter substring is
// preserved in the token that ends with it.
func (c *Chunkifier) SplitOnSplitter(text string) []string {
	if len(text) == 0 {
		return nil
	}
	if c.Splitter == "" {
		// if no splitter is defined, return the entire text as a single token
		return []string{text}
	}

	start := 0
	var out []string
	for {
		idx := strings.Index(text[start:], c.Splitter)
		if idx == -1 {
			break
		}
		end := start + idx + len(c.Splitter)
		out = append(out, text[start:end])
		start = end
	}
	if start < len(text) {
		out = append(out, text[start:])
	}
	return out
}

// SplitWords uses uniseg to split the text into words.
// CAVEAT: without spaces in the string it will behave like SplitGraphemes
func (c *Chunkifier) SplitWords(text string) []string {
	if len(text) == 0 {
		return nil
	}

	remaining := text
	state := -1
	var splitted []string
	for len(remaining) > 0 {
		word, rest, newState := uniseg.FirstWordInString(remaining, state)
		// Debug
		// fmt.Printf("DEBUG: uniseg.FirstWordInString => word=%q, rest=%q\n", word, rest)
		if word != "" {
			splitted = append(splitted, word)
		}
		remaining = rest
		state = newState
	}
	return splitted
}

// SplitGraphemes uses uniseg to split text into individual grapheme clusters. 
// This can be used for scripts that do not have clear spaces or word breaks, 
// but often leads to very short tokens.
func (c *Chunkifier) SplitGraphemes(text string) []string {
	if len(text) == 0 {
		return nil
	}

	var splitted []string
	remaining := text
	state := -1
	for len(remaining) > 0 {
		grapheme, rest, _, newState := uniseg.FirstGraphemeClusterInString(remaining, state)
		if grapheme != "" {
			splitted = append(splitted, grapheme)
		}
		remaining = rest
		state = newState
	}
	return splitted
}
