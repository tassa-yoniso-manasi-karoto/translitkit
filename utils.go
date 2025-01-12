package translitkit

import (
	"github.com/rivo/uniseg"
	"strings"
	"unicode/utf8"
)

// utils related to GenericQuerySplitter, WIP

type stringer interface {
	~string
}

func toString[T stringer](s T) string {
	return string(s)
}

func toStringSlice[T stringer](s []string) []T {
	result := make([]T, len(s))
	for i, v := range s {
		result[i] = T(v)
	}
	return result
}

func notTooBig[T stringer](arr []T, max int) bool {
	for _, str := range arr {
		if max > 0 && utf8.RuneCountInString(toString(str)) > max {
			return false
		}
	}
	return true
}

func splitSpace[T stringer](str T) []T {
	splits := strings.Split(toString(str), " ")
	return toStringSlice[T](splits)
}

func splitSentences[T stringer](text T) []T {
	sentences := make([]string, 0)

	if len(text) == 0 {
		return make([]T, 0)
	}

	remaining := toString(text)
	state := -1
	for len(remaining) > 0 {
		sentence, rest, newState := uniseg.FirstSentenceInString(remaining, state)
		if sentence != "" {
			sentences = append(sentences, strings.TrimSpace(sentence))
		}
		remaining = rest
		state = newState
	}

	return toStringSlice[T](sentences)
}

func splitWords[T stringer](text T) []T {
	words := make([]string, 0)

	if len(text) == 0 {
		return make([]T, 0)
	}

	remaining := toString(text)
	state := -1
	for len(remaining) > 0 {
		word, rest, newState := uniseg.FirstWordInString(remaining, state)
		if word != "" {
			words = append(words, strings.TrimSpace(word))
		}
		remaining = rest
		state = newState
	}

	return toStringSlice[T](words)
}

func splitGraphemes[T stringer](text T) []T {
	graphemes := make([]string, 0)

	if len(text) == 0 {
		return make([]T, 0)
	}

	remaining := toString(text)
	state := -1
	for len(remaining) > 0 {
		grapheme, rest, _, newState := uniseg.FirstGraphemeClusterInString(remaining, state)
		if grapheme != "" {
			graphemes = append(graphemes, grapheme)
		}
		remaining = rest
		state = newState
	}

	return toStringSlice[T](graphemes)
}
