
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
	return result.(*Module), nil
}

// defaultModule is an internal function that configures any module type with default providers.
// This interface-based implementation isn't needed anymore because jpn.Module embeds common.Module
// therefore, common.Module's methods can be shared to jpn.Module without constructing jpn.Module
// with defaultModule. Nonetheless I am keeping this design for now.
func defaultModule(lang string) (anyModule, error) {
	m := &Module{}
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


func getProvider(lang string, provType ProviderType, name string) (Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], error) {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	entry, ok := findProvider(lang, provType, name)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s (%s) for language %s or mul", name, provType, lang)
	}

	return entry.Provider, nil
}


// findProvider looks for a provider first in the specified language's registry,
// then falls back to multilingual providers if not found
func findProvider(lang string, provType ProviderType, name string) (ProviderEntry, bool) {
	// Try language-specific provider first
	if langProviders, exists := GlobalRegistry.Providers[lang]; exists {
		if entry, ok := getProviderFromMap(langProviders, provType, name); ok {
			return entry, true
		}
	}

	// Fallback to multilingual provider if not found and not already looking for mul
	if lang != "mul" {
		if mulProviders, exists := GlobalRegistry.Providers["mul"]; exists {
			return getProviderFromMap(mulProviders, provType, name)
		}
	}

	return ProviderEntry{}, false
}

// getProviderFromMap retrieves a provider entry from the appropriate map based on type
func getProviderFromMap(providers LanguageProviders, provType ProviderType, name string) (ProviderEntry, bool) {
	switch provType {
	case TokenizerType:
		entry, ok := providers.Tokenizers[name]
		return entry, ok
	case TransliteratorType:
		entry, ok := providers.Transliterators[name]
		return entry, ok
	case CombinedType:
		entry, ok := providers.Combined[name]
		return entry, ok
	default:
		return ProviderEntry{}, false
	}
}

// checkCapabilities validates if providers have required capabilities for a language
// and issues warnings if capabilities are missing
func checkCapabilities(lang string, entries []ProviderEntry, provType ProviderType, name string) {
	mustTokenize, _ := NeedsTokenization(lang)
	mustTransliterate, _ := NeedsTransliteration(lang)

	if !mustTokenize && !mustTransliterate {
		return
	}

	hasTokenization := false
	hasTransliteration := false

	// For Register function, we check a single entry
	if name != "" {
		for _, capability := range entries[0].Capabilities {
			if capability == "tokenization" {
				hasTokenization = true
			}
			if capability == "transliteration" {
				hasTransliteration = true
			}
		}

		if mustTokenize && !hasTokenization && (provType == TokenizerType || provType == CombinedType) {
			Log.Warn().
				Str("provider", name).
				Str("lang", lang).
				Msg("Registering provider which requires tokenization but providerType doesn't declare this capability")
		}
		if mustTransliterate && !hasTransliteration && (provType == TransliteratorType || provType == CombinedType) {
			Log.Warn().
				Str("provider", name).
				Str("lang", lang).
				Msg("Registering provider which requires transliteration but providerType doesn't declare this capability")
		}
		return
	}

	// For SetDefault function, we check all entries
	for _, p := range entries {
		for _, capability := range p.Capabilities {
			if capability == "tokenization" {
				hasTokenization = true
			}
			if capability == "transliteration" {
				hasTransliteration = true
			}
		}
	}

	if mustTokenize && !hasTokenization {
		Log.Warn().
			Str("lang", lang).
			Msg("Language requires tokenization but no provider declares this capability")
	}
	if mustTransliterate && !hasTransliteration {
		Log.Warn().
			Str("lang", lang).
			Msg("Language requires transliteration but no provider declares this capability")
	}
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

