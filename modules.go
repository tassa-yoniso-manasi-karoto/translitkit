package translitkit

import (
	"fmt"
	"strings"

	//iso "github.com/barbashov/iso639-3"
)

type AnyModule interface {
	Init() error
	RomanPostProcess(string, func(string) (string)) (string)
	Close() error
}

// BaseModule satisfies the AnyModule interface. It contains both Tokenization+Transliteration components.
type BaseModule struct {
	Lang           string // ISO-639 Part 3: i.e. "eng", "zho", "jpn"...
	ProviderType   ProviderType
	Tokenizer      Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Transliterator Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Combined       Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	MaxLenQuery    int
}

// ProviderNames returns the names of the provider(s) contained in the module.
// For combined providers, it returns a single name.
// For separate providers, it returns both tokenizer and transliterator names.
func (m BaseModule) ProviderNames() string {
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



func (m BaseModule) Init() error {
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

func (m BaseModule) MustInit() {
	if err := m.Init(); err != nil {
		panic(err)
	}
}

func (m BaseModule) Roman(input string) (string, error) {
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}

	tkns, err := m.Tokens(input)
	if err != nil {
		return "", err
	}
	return tkns.Roman(), nil
}

func (m BaseModule) RomanParts(input string) ([]string, error) {
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.RomanParts(), nil
}

func (m BaseModule) TokenizedParts(input string) ([]string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.TokenizedParts(), nil
}

func (m BaseModule) Tokenized(input string) (string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	parts, err := m.TokenizedParts(input)
	if err != nil {
		return "", err
	}
	return strings.Join(parts, " "), nil // FIXME ideally place space between word-words not word-punctuation or punct-punct
}

func (m BaseModule) Tokens(input string) (AnyTokenSliceWrapper, error) {
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

func (m BaseModule) Close() error {
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

func (m BaseModule) RomanPostProcess(s string, f func(string) (string)) (string) {
	return f(s)
}
