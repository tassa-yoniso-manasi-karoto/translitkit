package common

import (
	"fmt"
	"strings"
	"unicode/utf8"
	
	"github.com/rivo/uniseg"
)

func chunkify(s string, max int) (QuerySliced []string, err error) {
	var chunks = []string{s}
	if notTooBig(chunks, max) {
		return chunks, nil
	}
	// SplitSpace should do in most cases, the rest is just in case
	chunks = splitSpace(s)
	if notTooBig(chunks, max) {
		return chunks, nil
	}
	chunks = splitSentences(s)
	if notTooBig(chunks, max) {
		return chunks, nil
	}
	chunks = splitWords(s)
	if notTooBig(chunks, max) {
		return chunks, nil
	}
	chunks = splitGraphemes(s)
	if notTooBig(chunks, max) {
		return chunks, nil
	}
	return nil, fmt.Errorf("couldn't decompose string into smaller parts: →%s←" +
		"SplitGraphemes did at most: %#v", s, chunks)
}

func notTooBig(arr []string, max int) bool {
	for _, str := range arr {
		if max > 0 && utf8.RuneCountInString(str) > max {
			return false
		}
	}
	return true
}

func splitSpace(str string) []string {
	return strings.Split(str, " ")
}

func splitSentences(text string) (splitted []string) {
	if len(text) == 0 {
		return
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		sentence, rest, newState := uniseg.FirstSentenceInString(remaining, state)
		if sentence != "" {
			splitted = append(splitted, strings.TrimSpace(sentence))
		}
		remaining = rest
		state = newState
	}

	return
}

func splitWords(text string) (splitted []string) {
	if len(text) == 0 {
		return
	}

	remaining := text
	state := -1
	for len(remaining) > 0 {
		word, rest, newState := uniseg.FirstWordInString(remaining, state)
		if word != "" {
			splitted = append(splitted, strings.TrimSpace(word))
		}
		remaining = rest
		state = newState
	}

	return
}

func splitGraphemes(text string) (splitted []string) {
	if len(text) == 0 {
		return
	}

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

	return
}
