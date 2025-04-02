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
	Name    string   // Name for logging and debugging
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
		{Name: "SplitSpace", SplitFn: c.SplitSpace, Joiner: " "},
		{Name: "SplitSentences", SplitFn: c.SplitSentences, Joiner: " "},
		{Name: "SplitOnSplitter", SplitFn: c.SplitOnSplitter, Joiner: ""},
		// too problematic with writing systems that don't use spaces =
		// if there is no word delimitations found it will behave like splitGraphemes
		//{Name: "SplitWords", SplitFn: c.SplitWords, Joiner: ""},
		
		// risk of truncating words
		// {Name: "SplitGraphemes", SplitFn: c.SplitGraphemes, Joiner: ""},
	}
	return c
}

// Chunkify takes the given string s and a max length. The function tries different 
// approaches to split the text into chunks that are all within the maximum length.
func (c *Chunkifier) Chunkify(s string) ([]string, error) {
	Log.Trace().
		Int("MaxLength", c.MaxLength).
		Msgf("Chunkify: starting with input string of length %d", utf8.RuneCountInString(s))
	
	// If a negative max was passed or if the entire string already fits
	if c.MaxLength <= 0 || utf8.RuneCountInString(s) <= c.MaxLength {
		Log.Trace().Msg("Chunkify: string fits within max length, returning original string")
		return []string{s}, nil
	}

	// First try the standard method-by-method approach
	for _, method := range c.SplitMethods {
		Log.Trace().Msgf("Chunkify: trying split method %s with joiner %q", method.Name, method.Joiner)
		chunks, success, err := c.tryStandardSplit(s, method)
		if err != nil {
			return nil, err
		}
		if success {
			return chunks, nil
		}
	}
	
	// If standard splitting fails, try the recursive approach
	Log.Trace().Msg("Chunkify: standard splitting failed, attempting recursive approach")
	chunks, err := c.tryRecursiveSplit(s)
	if err != nil {
		// Try a more aggressive hybrid approach as a last resort
		Log.Trace().Msg("Chunkify: recursive splitting failed, attempting hybrid approach")
		chunks, err = c.tryHybridSplit(s)
		if err != nil {
			errMsg := fmt.Sprintf("could not decompose string into smaller parts: %q", s)
			Log.Trace().Msg(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
	}
	
	return chunks, nil
}

// tryStandardSplit attempts to split the string using a single method
// and checks if all tokens are within the length limit
func (c *Chunkifier) tryStandardSplit(s string, method SplitMethod) ([]string, bool, error) {
	tokens := method.SplitFn(s)
	Log.Trace().Msgf("Chunkify: obtained %d tokens", len(tokens))
	
	// Check if any tokens are too large
	allWithinLimit := true
	for _, token := range tokens {
		if count := utf8.RuneCountInString(token); count > c.MaxLength {
			Log.Trace().Msgf("Chunkify: oversized token (len=%d): %s", count, token)
			allWithinLimit = false
		}
	}
	
	if !allWithinLimit {
		Log.Trace().Msg("Chunkify: tokens exceed max length, skipping this split method")
		return nil, false, nil
	}
	
	// All tokens are within limit, combine them
	combined := combineTokens(tokens, method.Joiner, c.MaxLength)
	if combined == nil {
		return nil, false, nil
	}
	
	Log.Trace().Msgf("Chunkify: successfully combined tokens into %d chunks", len(combined))
	return combined, true, nil
}

// tryRecursiveSplit attempts to split the string recursively
// by applying different methods to problematic tokens
func (c *Chunkifier) tryRecursiveSplit(s string) ([]string, error) {
	Log.Trace().Msg("Chunkify: attempting recursive splitting")
	
	// Try the first method
	if len(c.SplitMethods) == 0 {
		return nil, fmt.Errorf("no split methods defined")
	}
	
	return c.splitRecursively(s, 0)
}

// splitRecursively splits a string using the method at the given index
// and recursively applies the next methods to any tokens that are too large
func (c *Chunkifier) splitRecursively(s string, methodIndex int) ([]string, error) {
	// If we've run out of methods, we can't split further
	if methodIndex >= len(c.SplitMethods) {
		return nil, fmt.Errorf("all splitting methods exhausted, unable to split string within max length")
	}
	
	method := c.SplitMethods[methodIndex]
	Log.Trace().Msgf("Chunkify: recursive split using method %d (%s) with joiner %q", 
		methodIndex, method.Name, method.Joiner)
	
	tokens := method.SplitFn(s)
	if len(tokens) <= 1 {
		// This method didn't help split the string, try the next one
		Log.Trace().Msgf("Chunkify: method %s produced only %d tokens, trying next method", 
			method.Name, len(tokens))
		return c.splitRecursively(s, methodIndex+1)
	}
	
	// Process each token: either keep it if small enough, or recursively split it
	var processedTokens []string
	for i, token := range tokens {
		tokenLen := utf8.RuneCountInString(token)
		if tokenLen <= c.MaxLength {
			// Token is small enough, keep it
			processedTokens = append(processedTokens, token)
		} else {
			// Token is too large, try to split it with the next method
			Log.Trace().Msgf("Chunkify: token %d from method %s is too large (len=%d), splitting recursively", 
				i, method.Name, tokenLen)
			
			var subTokens []string
			var err error
			
			// Try all remaining methods on this token
			for nextMethodIndex := 0; nextMethodIndex < len(c.SplitMethods); nextMethodIndex++ {
				// Skip current method to avoid infinite recursion
				if nextMethodIndex == methodIndex {
					continue
				}
				
				Log.Trace().Msgf("Chunkify: trying alternative method %s on oversized token", 
					c.SplitMethods[nextMethodIndex].Name)
					
				tempTokens := c.SplitMethods[nextMethodIndex].SplitFn(token)
				if len(tempTokens) > 1 {
					// This method helped split the token
					allSmall := true
					for _, t := range tempTokens {
						if utf8.RuneCountInString(t) > c.MaxLength {
							allSmall = false
							break
						}
					}
					
					if allSmall {
						// All sub-tokens are within limit
						subTokens = tempTokens
				        Log.Trace().Strs("subTokens", subTokens).Msgf("Chunkify: after using alternative method %s on oversized token: SUCCESS: allSmall is true", 
					    c.SplitMethods[nextMethodIndex].Name)
						break
					}
				}
			}
			
			// If we couldn't split with any other method, try recursive as last resort
			if len(subTokens) == 0 {
				subTokens, err = c.splitRecursively(token, methodIndex+1)
				if err != nil {
					// If we can't split this token further, log the problem and propagate the error
					Log.Trace().Msgf("Chunkify: failed to recursively split token: %s", err)
					return nil, err
				}
			}
			
			processedTokens = append(processedTokens, subTokens...)
		}
	}
	
	// Combine processed tokens
	combined := combineTokens(processedTokens, method.Joiner, c.MaxLength)
	if combined == nil {
		return nil, fmt.Errorf("failed to combine processed tokens within max length")
	}
	
	return combined, nil
}

// tryHybridSplit attempts a hybrid approach that tries each method on
// each problematic token independently
func (c *Chunkifier) tryHybridSplit(s string) ([]string, error) {
	Log.Trace().Msg("Chunkify: attempting hybrid splitting")
	
	// Start with the whole string as a single token
	tokens := []string{s}
	
	// Keep splitting until all tokens are within limit or we can't split further
	var progress bool
	maxIterations := 100 // Safety to prevent infinite loops
	iteration := 0
	
	for iteration < maxIterations {
		iteration++
		progress = false
		
		// Find tokens that are still too large
		var newTokens []string
		var hasLargeTokens bool
		
		for _, token := range tokens {
			if utf8.RuneCountInString(token) <= c.MaxLength {
				// This token is fine, keep it
				newTokens = append(newTokens, token)
			} else {
				// This token is too large, try to split it with any method
				hasLargeTokens = true
				split := false
				
				// Try each method in turn
				for _, method := range c.SplitMethods {
					splitTokens := method.SplitFn(token)
					if len(splitTokens) > 1 {
						// This method split the token into smaller parts
						// Check if any of the resulting tokens are now small enough
						smallerTokensFound := false
						for _, st := range splitTokens {
							if utf8.RuneCountInString(st) < utf8.RuneCountInString(token) {
								smallerTokensFound = true
								break
							}
						}
						
						if smallerTokensFound {
							newTokens = append(newTokens, splitTokens...)
							split = true
							progress = true
							break
						}
					}
				}
				
				// If we couldn't split this token with any method, keep it as is
				// (we'll eventually return an error for it)
				if !split {
					newTokens = append(newTokens, token)
				}
			}
		}
		
		// Update our token list
		tokens = newTokens
		
		// If we made no progress or all tokens are now within limit, we're done
		if !progress || !hasLargeTokens {
			break
		}
	}
	
	// Final check: are all tokens within limit?
	hasLargeTokens := false
	for _, token := range tokens {
		if utf8.RuneCountInString(token) > c.MaxLength {
			Log.Trace().Msgf("Chunkify: still have oversized token after hybrid split (len=%d): %s", 
				utf8.RuneCountInString(token), token)
			hasLargeTokens = true
		}
	}
	
	if hasLargeTokens {
		return nil, fmt.Errorf("failed to split all tokens within max length after hybrid approach")
	}
	
	// If we get here, all tokens are within limit
	// Combine them optimally
	result := combineTokens(tokens, "", c.MaxLength)
	if result == nil {
		return nil, fmt.Errorf("failed to combine tokens after hybrid splitting")
	}
	
	return result, nil
}

// tokensAreWithinLimit checks that each token is within the allowed length.
func tokensAreWithinLimit(tokens []string, max int) bool {
	if max <= 0 {
		return true
	}
	for i, token := range tokens {
		if count := utf8.RuneCountInString(token); count > max {
			Log.Trace().Msgf("Chunkify: ERROR: token %d len=%d: %s", i, count, token)
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
	// Verify the final result doesn't exceed max in any chunk
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