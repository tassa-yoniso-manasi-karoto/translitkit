package translitkit

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

/* CLAUDE SUMMARY

Core Requirements:
	Modular design with common interface methods: Init(), Query(), PostProcess(), Close()
	Metadata storage for each module (language, provider info, max query length etc.)
	Support for multiple providers per language
	Text processing with two main operations:
		→ Tokenization (i.e. for ja supported by Mecab, Kagome)
		→ Transliteration (i.e. for ja supported by Ichiran)
	Ability to combine different providers (e.g., Mecab for tokenization + another one for transliteration)
	
Technical Considerations:
	Need to handle cases where one provider can do both operations (Ichiran)
	Need to handle cases where operations must be split between providers (Mecab+Ichiran)
	Must maintain a clean interface while supporting different provider combinations
	Should be extensible for adding new providers or capabilities
*/

type ProviderType string

const (
	TokenizerType      ProviderType = "tokenizer"
	TransliteratorType ProviderType = "transliterator"
	CombinedType       ProviderType = "combined"
)

// Unified interface for all providers of any type
type Provider[In AnyTokenSliceWrapper, Out AnyTokenSliceWrapper] interface {
	Init() error
	Process(m AnyModule, input In) (Out, error)
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
	Provider     Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
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



func placeholder2345w4567ui() {
	fmt.Print("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

