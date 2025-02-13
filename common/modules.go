package common

import (
	"fmt"
	"strings"
	"math"

	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	//iso "github.com/barbashov/iso639-3"
)

// Private because not needed at this point.
// Could become needed of more sophisticated NLP providers are implemented.
// Method set needs more iterations to be defined.
type anyModule interface {
	Init() error
	InitRecreate(bool) error
	MustInit()
	ProviderNames() string
	RomanPostProcess(string, func(string) string) string
	Close() error
	
	setLang(string)
	// getMaxQueryLen() int ?
	setProviders([]ProviderEntry) error
}

// Module satisfies the anyModule interface.
// It contains both Tokenization+Transliteration components.

type Module struct {
	Lang           string // ISO-639 Part 3: i.e. "eng", "zho", "jpn"...
	ProviderType   ProviderType
	Tokenizer      Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Transliterator Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Combined       Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
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



func (m *Module) InitRecreate(noCache bool) error {
	if m.Combined != nil {
		return m.Combined.InitRecreate(noCache)
	}
	if err := m.Tokenizer.InitRecreate(noCache); err != nil {
		return fmt.Errorf("tokenizer InitRecreate failed: %v", err)
	}
	if err := m.Transliterator.InitRecreate(noCache); err != nil {
		return fmt.Errorf("transliterator InitRecreate failed: %v", err)
	}
	return nil
}

func (m *Module) MustInit() {
	if err := m.InitRecreate(false); err != nil {
		panic(err)
	}
}

func (m *Module) Tokens(input string) (AnyTokenSliceWrapper, error) {
	tsw, err := serialize(input, m.getMaxQueryLen())
	if err != nil {
		return nil, fmt.Errorf("input serialization failed: len(input)=%d, %v", len(input), err)
	}

	if m.Combined != nil {
		tsw, err = m.Combined.ProcessFlowController(tsw)
		if err != nil {
			return &TknSliceWrapper{}, fmt.Errorf("combined processing failed: %v", err)
		}
	} else {
		tsw, err = m.Tokenizer.ProcessFlowController(tsw)
		if err != nil {
			return &TknSliceWrapper{}, fmt.Errorf("tokenization failed: %v", err)
		}
		if m.Transliterator != nil {
			if tsw, err = m.Transliterator.ProcessFlowController(tsw); err != nil {
				return &TknSliceWrapper{}, fmt.Errorf("transliteration failed: %v", err)
			}
		}
	}
	if tsw == nil {
		return tsw, fmt.Errorf("fatal: nil tokens returned by module: %#v", m)
	}
	return tsw, nil
}


func (m *Module) LexicalTokens(input string) (AnyTokenSliceWrapper, error) {
	raw, err := m.Tokens(input)
	if err != nil {
		return nil, err
	}
	return FilterAny(raw), nil
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
	tkns, err := m.LexicalTokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.RomanParts(), nil
}

func (m *Module) Tokenized(input string) (string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return "", err
	}
	return tkns.Tokenized(), nil 
}

func (m *Module) TokenizedParts(input string) ([]string, error) {
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.LexicalTokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.TokenizedParts(), nil
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


func (m *Module) setLang(lang string) {
	m.Lang = lang
}

// getMaxQueryLen returns the maximum query length that can be processed by the module.
// For combined providers, it returns the provider's limit.
// For separate providers, it returns the smallest limit between tokenizer and transliterator.
// If MaxQueryLen is already set, returns that value instead of recalculating.
func (m *Module) getMaxQueryLen() int {
	providers, err := m.listProviders()
	if err != nil {
		return math.MaxInt64
	}

	return getQueryLenLimit(providers...)
}

func (m *Module) setProviders(providers []ProviderEntry) error {
	if len(providers) == 0 {
		return fmt.Errorf("cannot set empty providers")
	}

	if providers[0].Type == CombinedType {
		// For combined provider, only one entry is needed
		if len(providers) > 1 {
			return fmt.Errorf("combined provider cannot be used with other providers")
		}
		m.Combined = providers[0].Provider
		m.ProviderType = CombinedType
	} else {
		// For separate providers, tokenizer is required but transliterator is optional
		if providers[0].Type != TokenizerType {
			return fmt.Errorf("first provider must be a tokenizer")
		}
		m.Tokenizer = providers[0].Provider

		// Set transliterator if provided
		if len(providers) > 1 {
			if providers[1].Type != TransliteratorType {
				return fmt.Errorf("second provider must be a transliterator")
			}
			m.Transliterator = providers[1].Provider
		}
	}
	return nil
}

func (m *Module) listProviders() (providers []ProviderEntry, err error) {
	if m.Combined != nil {
		// For combined provider, return single entry
		providers = append(providers, ProviderEntry{
			Provider: m.Combined,
			Type:     CombinedType,
		})
		return providers, nil
	}

	// For separate providers, return both tokenizer and transliterator
	if m.Tokenizer != nil {
		providers = append(providers, ProviderEntry{
			Provider: m.Tokenizer,
			Type:     TokenizerType,
		})
	}

	if m.Transliterator != nil {
		providers = append(providers, ProviderEntry{
			Provider: m.Transliterator,
			Type:     TransliteratorType,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers found in module")
	}

	return providers, nil
}


func placeholder3456456543() {
	fmt.Println("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
