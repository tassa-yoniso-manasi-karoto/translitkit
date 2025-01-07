package translitkit

import (
	"fmt"
)

type ProviderType string

const (
	TokenizerType      ProviderType = "tokenizer"
	TransliteratorType ProviderType = "transliterator"
	CombinedType       ProviderType = "combined"
)

// Unified interface for all providers of any type
type Provider[In AnyTokenSlice, Out AnyTokenSlice] interface {
	Init() error
	Process(m Module, input In) (Out, error)
	Name() string
	GetType() ProviderType
	//GetCapabilities() []string
	Close() error
}

type DefaultProviders []ProviderEntry

type SeparateProviders struct {
	Tokenizer      ProviderEntry
	Transliterator ProviderEntry
}

type LanguageProviders struct {
	Defaults        DefaultProviders
	Tokenizers      map[string]ProviderEntry
	Transliterators map[string]ProviderEntry
	Combined        map[string]ProviderEntry
}

type ProviderEntry struct {
	Provider     Provider[AnyTokenSlice, AnyTokenSlice]
	Capabilities []string
	Type         ProviderType
}

type ProviderConfig struct {
	Name         string
	ProviderType ProviderType
	Params       map[string]interface{}
}

type DefaultProvider struct {
	CombinedName       string
	TokenizerName      string
	TransliteratorName string
}

// // FIXME WIP
// func GenericQuerySplitter(input []MyString, max int) (QuerySliced [][]MyString, err error) {
// 	for _, s := range input {
// 		var chunks = []MyString{s}
// 		if notTooBig(chunks, max) {
// 			QuerySliced = append(QuerySliced, chunks)
// 			continue
// 		}
// 		chunks = SplitSpace(s)
// 		if notTooBig(chunks, max) {
// 			QuerySliced = append(QuerySliced, chunks)
// 			continue
// 		}
// 		chunks = SplitSentences(s)
// 		if notTooBig(chunks, max) {
// 			QuerySliced = append(QuerySliced, chunks)
// 			continue
// 		}
// 		chunks = SplitWords(s)
// 		if notTooBig(chunks, max) {
// 			QuerySliced = append(QuerySliced, chunks)
// 			continue
// 		}
// 		chunks = SplitGraphemes(s)
// 		if notTooBig(chunks, max) {
// 			QuerySliced = append(QuerySliced, chunks)
// 			continue
// 		}
// 		return nil, fmt.Errorf("couldn't decompose string into smaller parts: →%s←" +
// 			"SplitGraphemes did at most: %#v", s, chunks)
// 	}
// 	return
// }

// func GenericQueriesChunkify(s MyString, max int) (QuerySliced []MyString, err error) {
// 	var chunks = []MyString{s}
// 	if notTooBig(chunks, max) {
// 		return chunks, nil
// 	}
// 	chunks = SplitSpace(s)
// 	if notTooBig(chunks, max) {
// 		return chunks, nil
// 	}
// 	chunks = SplitSentences(s)
// 	if notTooBig(chunks, max) {
// 		return chunks, nil
// 	}
// 	chunks = SplitWords(s)
// 	if notTooBig(chunks, max) {
// 		return chunks, nil
// 	}
// 	chunks = SplitGraphemes(s)
// 	if notTooBig(chunks, max) {
// 		return chunks, nil
// 	}
// 	return nil, fmt.Errorf("couldn't decompose string into smaller parts: →%s←" +
// 		"SplitGraphemes did at most: %#v", s, chunks)
// }

// // FIXME WIP
// func GenericTokenProcessor(p Provider[MyString, Tkn], Query []MyString, f Module) (results []Tkn, err error) {
// 	for _, chunk := range Query {
// 		tokens, err := p.process(f, chunk)
// 		if err != nil {
// 			return nil, fmt.Errorf("running tokenProcessor() failed for chunk: %#v", chunk)
// 		}
// 		results = append(results, tokens...)
// 	}
// 	return
// }

// // FIXME WIP
// func GenericMyStringProcessor(p Provider[MyString, MyString], Query []MyString, f Module) (results []MyString, err error) {
// 	var sb strings.Builder
// 	for _, chunk := range Query {
// 		s, err := p.process(f, chunk)
// 		if err != nil {
// 			return nil, fmt.Errorf("running genericSplittedQueryProcessor() failed for chunk: %#v", chunk)
// 		}
// 		// idk why but apparently it assumes "s" as "variable of type []MyString"
// 		sb.WriteString(string(s[0]))
// 	}
// 	results = append(results, MyString(sb.String()))
// 	return
// }
