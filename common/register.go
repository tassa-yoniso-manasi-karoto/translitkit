
package common

import (
	"fmt"
	"sync"
	
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
func Register(languageCode string, entry ProviderEntry) error {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return fmt.Errorf(errNotISO639, languageCode)
	}
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	// Check capabilities based on supported modes
	modes := entry.Provider.SupportedModes()
	if len(modes) > 0 {
		checkCapabilities(lang, []ProviderEntry{entry}, modes[0], entry.Provider.Name())
	}

	// Initialize language Providers if not exists
	if _, exists := GlobalRegistry.Providers[lang]; !exists {
		GlobalRegistry.Providers[lang] = LanguageProviders{
			Providers: make([]ProviderEntry, 0),
			Defaults:  make([]ProviderEntry, 0),
		}
	}

	// Verify Provider interface is implemented
	if entry.Provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	// Check if provider already registered (avoid duplicates)
	providers := GlobalRegistry.Providers[lang]
	for i, existing := range providers.Providers {
		if existing.Provider.Name() == entry.Provider.Name() {
			// Update existing entry
			providers.Providers[i] = entry
			GlobalRegistry.Providers[lang] = providers
			return nil
		}
	}

	// Add new provider
	providers.Providers = append(providers.Providers, entry)
	GlobalRegistry.Providers[lang] = providers

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
	m := newModule()
	m.Lang = lang

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
		return nil, fmt.Errorf("failed to set providers: %w", err)
	}
	m.chunkifier = NewChunkifier(m.getMaxQueryLen())
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
			Providers: make([]ProviderEntry, 0),
			Defaults:  make([]ProviderEntry, 0),
		}
	}

	if len(providers) == 0 {
		return fmt.Errorf("cannot set empty default providers")
	}

	// Extract provider interfaces for validation
	providerInterfaces := make([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], len(providers))
	for i, entry := range providers {
		providerInterfaces[i] = entry.Provider
	}
	
	// Validate the provider setup for this language
	if err := validateProviderSetup(lang, providerInterfaces); err != nil {
		return err
	}
	
	// Verify providers are registered
	if len(providers) == 1 {
		// Check if it's a combined provider
		modes := providers[0].Provider.SupportedModes()
		hasCombined := false
		for _, mode := range modes {
			if mode == CombinedMode {
				hasCombined = true
				break
			}
		}
		
		if hasCombined {
			if _, ok := findProvider(lang, CombinedMode, providers[0].Provider.Name()); !ok {
				return fmt.Errorf("combined provider \"%s\" not found in registered providers", providers[0].Provider.Name())
			}
		} else {
			// Check as transliterator
			if _, ok := findProvider(lang, TransliteratorMode, providers[0].Provider.Name()); !ok {
				return fmt.Errorf("provider \"%s\" not found in registered providers", providers[0].Provider.Name())
			}
		}
	} else if len(providers) >= 2 {
		// First should be tokenizer
		if _, ok := findProvider(lang, TokenizerMode, providers[0].Provider.Name()); !ok {
			return fmt.Errorf("tokenizer \"%s\" not found in registered providers", providers[0].Provider.Name())
		}
		
		// Second should be transliterator
		if _, ok := findProvider(lang, TransliteratorMode, providers[1].Provider.Name()); !ok {
			return fmt.Errorf("transliterator \"%s\" not found in registered providers", providers[1].Provider.Name())
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

