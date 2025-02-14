
package common

import (
	"fmt"
	"sync"
	"context"
	
	iso "github.com/barbashov/iso639-3"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

const errNotISO639 = "\"%s\" isn't a ISO-639 language code"

var GlobalRegistry = &Registry{
	Providers: make(map[string]LanguageProviders),
}

type Registry struct {
	mu        sync.RWMutex
	Providers map[string]LanguageProviders
}

var BrowserAccessURL = ""

// Register adds a new Provider to the global registry for the specified language.
// It performs capability validation and warns if the Provider's capabilities
// don't match the language requirements.
func Register(languageCode string, provType ProviderType, name string, entry ProviderEntry) error {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return fmt.Errorf(errNotISO639, languageCode)
	}
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	checkCapabilities(lang, []ProviderEntry{entry}, provType, name)

	// Initialize language Providers if not exists
	if _, exists := GlobalRegistry.Providers[lang]; !exists {
		GlobalRegistry.Providers[lang] = LanguageProviders{
			Tokenizers:      make(map[string]ProviderEntry),
			Transliterators: make(map[string]ProviderEntry),
			Combined:        make(map[string]ProviderEntry),
			Defaults:        make([]ProviderEntry, 0),
		}
	}

	// Verify entry type matches provType
	if entry.Type != provType {
		return fmt.Errorf("provider type mismatch: expected %s, got %s", provType, entry.Type)
	}

	// Verify Provider interface is implemented
	if entry.Provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	// Register the provider
	switch provType {
	case TokenizerType:
		GlobalRegistry.Providers[lang].Tokenizers[name] = entry
	case TransliteratorType:
		GlobalRegistry.Providers[lang].Transliterators[name] = entry
	case CombinedType:
		GlobalRegistry.Providers[lang].Combined[name] = entry
	}

	return nil
}


// DefaultModule returns a new Module configured with the default providers
// for the specified language.
func DefaultModule(languageCode string) (*Module, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return nil, fmt.Errorf(errNotISO639, languageCode)
	}
	result, err := defaultModule(lang)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// defaultModule is an internal function that configures a common with default providers for a given language.
func defaultModule(lang string) (*Module, error) {
	m := &Module{ ctx: context.Background()}
	m.setLang(lang)

	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	langProviders, exists := GlobalRegistry.Providers[lang]
	if !exists {
		return nil, fmt.Errorf("defaultModule: no providers registered for language: %s", lang)
	}

	if len(langProviders.Defaults) == 0 {
		return nil, fmt.Errorf("no default providers set for language: %s", lang)
	}

	if err := m.setProviders(langProviders.Defaults); err != nil {
		return nil, fmt.Errorf("failed to set providers: %v", err)
	}

	return m, nil
}

// SetDefault configures the default Providers for a language in the global registry.
// It validates that the Providers have the necessary capabilities for the language.
func SetDefault(languageCode string, providers []ProviderEntry) error {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return fmt.Errorf(errNotISO639, languageCode)
	}
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	checkCapabilities(lang, providers, "", "")

	// Initialize language providers if not exists
	if _, exists := GlobalRegistry.Providers[lang]; !exists {
		GlobalRegistry.Providers[lang] = LanguageProviders{
			Tokenizers:      make(map[string]ProviderEntry),
			Transliterators: make(map[string]ProviderEntry),
			Combined:        make(map[string]ProviderEntry),
			Defaults:        make([]ProviderEntry, 0),
		}
	}

	if len(providers) == 0 {
		return fmt.Errorf("cannot set empty default providers")
	}

	// Validate providers
	name := providers[0].Provider.Name()
	if providers[0].Type == CombinedType {
		if len(providers) > 1 {
			return fmt.Errorf("combined provider cannot be used with other providers")
		}
		// Verify the provider exists
		if _, ok := findProvider(lang, CombinedType, providers[0].Provider.Name()); !ok {
			return fmt.Errorf("combined provider \"%s\" not found in registered providers", name)
		}
	} else {
		// Require tokenizer but make transliterator optional
		if providers[0].Type != TokenizerType {
			return fmt.Errorf("first provider must be a tokenizer")
		}
		if _, ok := findProvider(lang, TokenizerType, name); !ok {
			return fmt.Errorf("tokenizer \"%s\" not found in registered providers", name)
		}

		// Check transliterator if provided
		if len(providers) > 1 {
			name := providers[1].Provider.Name()
			if providers[1].Type != TransliteratorType {
				return fmt.Errorf("second provider must be a transliterator")
			}
			if _, ok := findProvider(lang, TransliteratorType, name); !ok {
				return fmt.Errorf("transliterator \"%s\" not found in registered providers", name)
			}
		}
	}

	langProviders := GlobalRegistry.Providers[lang]
	langProviders.Defaults = providers
	GlobalRegistry.Providers[lang] = langProviders
	return nil
}


func IsValidISO639(lang string) (stdLang string, ok bool) {
	code := iso.FromAnyCode(lang)
	if code == nil {
		return
	}
	return code.Part3, true
}


// NeedsTokenization returns true if the given language doesn't use space to
// separate words and requires tokenization.
// The language code can be in any ISO 639 code format.
func NeedsTokenization(languageCode string) (bool, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return false, fmt.Errorf(errNotISO639, languageCode)
	}
	for _, code := range langsNeedTokenization {
		if lang == code {
			return true, nil
		}
	}
	return false, nil
}

// NeedsTransliteration returns true if the given language doesn't use the roman
// script and requires transliteration.
// The language code can be in any ISO 639 code format.
func NeedsTransliteration(languageCode string) (bool, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return false, fmt.Errorf(errNotISO639, languageCode)
	}
	for _, code := range langsNeedTransliteration {
		if lang == code {
			return true, nil
		}
	}
	return false, nil
}




func placeholder23456ui() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

