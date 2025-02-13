package common

import (
	"fmt"
	"strings"
	"unicode/utf8"
	
	"github.com/rivo/uniseg"
)


// chunkify takes a string s and a maximum allowed length, and returns a slice
// of chunks that together cover s. Each chunk will be as long as possible (while
// remaining under or equal to max), so that we use as few chunks as possible.
// The function tries several tokenization strategies and, for each, attempts to
// recombine the tokens into chunks.
func chunkify(s string, max int) (chunks []string, err error) {
	// If the entire string fits, return early.
	if max > 0 && utf8.RuneCountInString(s) <= max {
		return []string{s}, nil
	}

	// Try several splitting methods, in order of decreasing “natural” chunk sizes.
	// For each, we check that the tokens are individually under the limit (they
	// should be, but if not, we have to move on) and then try to combine them.
	splitMethods := []struct {
		splitFn func(string) []string
		joiner  string // how to join tokens to re-create a chunk
	}{
		{splitFn: splitSpace, joiner: " "},
		{splitFn: splitSentences, joiner: " "},
		//{splitFn: splitWords, joiner: " "}, // too problematic with writing systems that don't use spaces
		//{splitFn: splitGraphemes, joiner: ""}, // risk of truncating words
	}

	for _, method := range splitMethods {
		tokens := method.splitFn(s)
		// Check that each token is individually not too long.
		if !tokensAreWithinLimit(tokens, max) {
			continue
		}
		combined := combineTokens(tokens, method.joiner, max)
		if combined != nil {
			return combined, nil
		}
	}

	return nil, fmt.Errorf("couldn't decompose string into smaller parts: →%s←", s)
}

// tokensAreWithinLimit checks that each token in the slice is within the allowed length.
func tokensAreWithinLimit(tokens []string, max int) bool {
	for _, token := range tokens {
		if max > 0 && utf8.RuneCountInString(token) > max {
			return false
		}
	}
	return true
}

// combineTokens takes a slice of tokens and a joiner string, and returns a slice
// of chunks where tokens are combined greedily into the largest possible chunks
// that do not exceed the maximum allowed length.
// If for any reason a valid combination cannot be formed, it returns nil.
func combineTokens(tokens []string, joiner string, max int) []string {
	var result []string
	var current string

	for i, token := range tokens {
		if current == "" {
			// Start a new chunk.
			current = token
			continue
		}
		// Prepare a candidate by appending the joiner and the token.
		candidate := current + joiner + token
		if utf8.RuneCountInString(candidate) <= max {
			// If the candidate is valid, update the current chunk.
			current = candidate
		} else {
			// Otherwise, push the current chunk onto result and start a new chunk.
			result = append(result, current)
			current = token
		}

		// If this is the last token, then add the current chunk to the result.
		if i == len(tokens)-1 {
			result = append(result, current)
		}
	}
	// In the case where there were no tokens, current might be non-empty.
	if current != "" && (len(result) == 0 || result[len(result)-1] != current) {
		result = append(result, current)
	}
	// Final check: all combined chunks should be within the limit.
	for _, chunk := range result {
		if utf8.RuneCountInString(chunk) > max {
			return nil
		}
	}
	return result
}

// splitSpace splits the input string into tokens that include both words and spaces.
// Every space character is treated as its own token.
func splitSpace(str string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range str {
		if r == ' ' {
			// If we've accumulated non-space characters, add them as a token.
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			// Add the space itself as a token.
			tokens = append(tokens, string(r))
		} else {
			// Accumulate non-space runes.
			current.WriteRune(r)
		}
	}

	// If there's any remaining non-space text, add it as a token.
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}


// splitSentences uses uniseg to split the text into sentences.
func splitSentences(text string) (splitted []string) {
	if len(text) == 0 {
		return nil
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		// uniseg.FirstSentenceInString returns the sentence, the remaining text,
		// and a new state.
		sentence, rest, newState := uniseg.FirstSentenceInString(remaining, state)
		if sentence != "" {
			splitted = append(splitted, strings.TrimSpace(sentence))
		}
		remaining = rest
		state = newState
	}

	return splitted
}

// splitWords uses uniseg to split the text into words.
func splitWords(text string) (splitted []string) {
	if len(text) == 0 {
		return nil
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		// uniseg.FirstWordInString returns the word, the remaining text,
		// and a new state.
		word, rest, newState := uniseg.FirstWordInString(remaining, state)
		if word != "" {
			splitted = append(splitted, strings.TrimSpace(word))
		}
		remaining = rest
		state = newState
	}

	return splitted
}

// splitGraphemes uses uniseg to split the text into grapheme clusters.
func splitGraphemes(text string) (splitted []string) {
	if len(text) == 0 {
		return nil
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		// uniseg.FirstGraphemeClusterInString returns the grapheme, the remaining text,
		// and a new state.
		grapheme, rest, _, newState := uniseg.FirstGraphemeClusterInString(remaining, state)
		if grapheme != "" {
			splitted = append(splitted, grapheme)
		}
		remaining = rest
		state = newState
	}

	return splitted
}
