package common

import (
	"fmt"
	"strings"

	//iso "github.com/barbashov/iso639-3"
)

// Private because not needed at this point.
// Could become needed of more sophisticated NLP providers are implemented.
// Method set needs more iterations to be defined.
type anyModule interface {
	Init() error
	MustInit() error
	ProviderNames() string
	RomanPostProcess(string, func(string) (string)) (string)
	Close() error
}

// Module satisfies the AnyModule interface. It contains both Tokenization+Transliteration components.

type Module struct {
	Lang           string // ISO-639 Part 3: i.e. "eng", "zho", "jpn"...
	ProviderType   ProviderType
	Tokenizer      Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Transliterator Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Combined       Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	MaxLenQuery    int
}

// NewModule creates a Module for the specified language using either default Providers
// or the explicitly named ones. If providerNames is empty, default Providers are used.
// For a combined Provider, specify one name. For separate Providers, specify two names
// in the order: tokenizer, transliterator.
//
// Example usage:
//
//	module, err := NewModule("jpn") // Use defaults
//	module, err := NewModule("jpn", "ichiran") // Use combined Provider
//	module, err := NewModule("jpn", "mecab", "kakasi") // Use separate Providers
func NewModule(languageCode string, providerNames ...string) (*Module, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return nil, fmt.Errorf(errNotISO639, languageCode)
	}
	if len(providerNames) == 0 {
		return DefaultModule(lang)
	}

	module := &Module{
		Lang: lang,
	}

	if len(providerNames) == 1 {
		// Try to get as combined Provider
		if Provider, err := getProvider(lang, CombinedType, providerNames[0]); err == nil {
			module.Combined = Provider
			module.ProviderType = CombinedType
			return module, nil
		}
		return nil, fmt.Errorf("single Provider %s not found as combined Provider for language %s", providerNames[0], lang)
	}

	if len(providerNames) == 2 {
		// Get tokenizer
		tokenizer, err := getProvider(lang, TokenizerType, providerNames[0])
		if err != nil {
			return nil, fmt.Errorf("tokenizer %s not found: %v", providerNames[0], err)
		}
		
		// Get transliterator
		transliterator, err := getProvider(lang, TransliteratorType, providerNames[1])
		if err != nil {
			return nil, fmt.Errorf("transliterator %s not found: %v", providerNames[1], err)
		}

		module.Tokenizer = tokenizer
		module.Transliterator = transliterator
		return module, nil
	}

	return nil, fmt.Errorf("invalid number of Provider names: expected 1 or 2, got %d", len(providerNames))
}

// ProviderNames returns the names of the provider(s) contained in the module.
// For combined providers, it returns a single name.
// For separate providers, it returns both tokenizer and transliterator names.
func (m *Module) ProviderNames() string {
	if m.Combined != nil {
		return m.Combined.Name()
	}
	
	names := make([]string, 0, 2)
	if m.Tokenizer != nil {
		names = append(names, m.Tokenizer.Name())
	}
	if m.Transliterator != nil {
		names = append(names, m.Transliterator.Name())
	}
	return strings.Join(names, "→")
}



func (m *Module) Init() error {
	if m.Combined != nil {
		return m.Combined.Init()
	}
	if err := m.Tokenizer.Init(); err != nil {
		return fmt.Errorf("tokenizer init failed: %v", err)
	}
	if err := m.Transliterator.Init(); err != nil {
		return fmt.Errorf("transliterator init failed: %v", err)
	}
	return nil
}

func (m *Module) MustInit() {
	if err := m.Init(); err != nil {
		panic(err)
	}
}

func (m *Module) Roman(input string) (string, error) {
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}

	tkns, err := m.Tokens(input)
	if err != nil {
		return "", err
	}
	return tkns.Roman(), nil
}

func (m *Module) RomanParts(input string) ([]string, error) {
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.RomanParts(), nil
}

func (m *Module) TokenizedParts(input string) ([]string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.TokenizedParts(), nil
}

func (m *Module) Tokenized(input string) (string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	parts, err := m.TokenizedParts(input)
	if err != nil {
		return "", err
	}
	return strings.Join(parts, " "), nil // FIXME ideally place space between word-words not word-punctuation or punct-punct
}

func (m *Module) Tokens(input string) (AnyTokenSliceWrapper, error) {
	var result AnyTokenSliceWrapper
	var err error

	if m.Combined != nil {
		result, err = m.Combined.Process(m, Serialize(input))
	} else {
		intermediate, err := m.Tokenizer.Process(m, Serialize(input))
		if err != nil {
			return nil, fmt.Errorf("tokenization failed: %v", err)
		}
		result, err = m.Transliterator.Process(m, intermediate)
		if err != nil {
			return nil, fmt.Errorf("transliteration failed: %v", err)
		}
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Module) Close() error {
	if m.Combined != nil {
		return m.Combined.Close()
	}
	if err := m.Tokenizer.Close(); err != nil {
		return fmt.Errorf("tokenizer close failed: %v", err)
	}
	if err := m.Transliterator.Close(); err != nil {
		return fmt.Errorf("transliterator close failed: %v", err)
	}
	return nil
}

func (m *Module) RomanPostProcess(s string, f func(string) (string)) (string) {
	return f(s)
}
