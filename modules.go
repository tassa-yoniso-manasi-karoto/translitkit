package translitkit

import (
	"fmt"
	"strings"

	iso "github.com/barbashov/iso639-3"
)

type Module interface {
	Init() error
	RomanPostProcess(string, func(string) (string, error)) (string, error)
	Close() error
}

// BaseModule satisfies the Module interface. It contains both Tokenization+Transliteration components.
type BaseModule struct {
	Lang           *iso.Language
	ProviderType   ProviderType
	Tokenizer      Provider[AnyTokenSlice, AnyTokenSlice]
	Transliterator Provider[AnyTokenSlice, AnyTokenSlice]
	Combined       Provider[AnyTokenSlice, AnyTokenSlice]
	MaxLenQuery    int
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
	return tkns.GetRomanization(), nil
}

func (m BaseModule) RomanParts(input string) ([]string, error) {
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.GetRomanizationParts(), nil
}

func (m BaseModule) TokenizedParts(input string) ([]string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.GetTokenizedParts(), nil
}

func (m BaseModule) TokenizedStr(input string) (string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	parts, err := m.TokenizedParts(input)
	if err != nil {
		return "", err
	}
	return strings.Join(parts, " "), nil // FIXME ideally place space between word-words not word-punctuation or punct-punct
}

func (m BaseModule) Tokens(input string) (AnyTokenSlice, error) {
	var result AnyTokenSlice
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

func (m BaseModule) RomanPostProcess(s string, f func(string) (string, error)) (string, error) {
	return f(s)
}
