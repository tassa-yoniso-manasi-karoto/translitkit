
package common

import (
	"fmt"
	"sync"
	"os"
	
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

	switch provType {
	case TokenizerType:
		GlobalRegistry.Providers[lang].Tokenizers[name] = entry
	case TransliteratorType:
		GlobalRegistry.Providers[lang].Transliterators[name] = entry
	case CombinedType:
		GlobalRegistry.Providers[lang].Combined[name] = entry
	default:
		return fmt.Errorf("unknown provider type: %s", provType)
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
func SetDefault(languageCode string, Providers []ProviderEntry) error {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return fmt.Errorf(errNotISO639, languageCode)
	}
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	checkCapabilities(lang, Providers, "", "")

	langProviders, exists := GlobalRegistry.Providers[lang]
	if !exists {
		return fmt.Errorf("SetDefault: no Providers registered for language: %s", lang)
	}

	if len(Providers) == 0 {
		return fmt.Errorf("cannot set empty default Providers")
	}

	// Validate Providers
	if Providers[0].Type == CombinedType {
		// For combined provider, only one entry is needed
		if len(Providers) > 1 {
			return fmt.Errorf("combined provider cannot be used with other Providers")
		}
		// Verify the provider exists in combined map
		for _, entry := range langProviders.Combined {
			if entry.Provider == Providers[0].Provider {
				goto valid
			}
		}
		return fmt.Errorf("combined provider not found in registered Providers")
	} else {
		// For separate Providers, need both tokenizer and transliterator
		if len(Providers) != 2 {
			return fmt.Errorf("separate mode requires exactly 2 Providers (tokenizer + transliterator)")
		}
		if Providers[0].Type != TokenizerType || Providers[1].Type != TransliteratorType {
			return fmt.Errorf("separate Providers must be tokenizer + transliterator in that order")
		}
		// Verify both Providers exist in their respective maps
		var found bool
		for _, entry := range langProviders.Tokenizers {
			if entry.Provider == Providers[0].Provider {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tokenizer not found in registered Providers")
		}
		found = false
		for _, entry := range langProviders.Transliterators {
			if entry.Provider == Providers[1].Provider {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("transliterator not found in registered Providers")
		}
	}

valid:
	langProviders.Defaults = Providers
	GlobalRegistry.Providers[lang] = langProviders
	return nil
}


func getProvider(lang string, provType ProviderType, name string) (Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], error) {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	langProviders, exists := GlobalRegistry.Providers[lang]
	if !exists {
		return nil, fmt.Errorf("GetProvider: no Providers registered for language: %s", lang)
	}

	var entry ProviderEntry
	var ok bool

	switch provType {
	case TokenizerType:
		entry, ok = langProviders.Tokenizers[name]
	case TransliteratorType:
		entry, ok = langProviders.Transliterators[name]
	case CombinedType:
		entry, ok = langProviders.Combined[name]
	default:
		return nil, fmt.Errorf("unknown provider type: %s", provType)
	}

	if !ok {
		return nil, fmt.Errorf("provider not found: %s (%s)", name, provType)
	}

	return entry.Provider, nil
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
			fmt.Fprintf(os.Stderr, "Warning: Registering Provider %s for %s which requires tokenization but Provider doesn't declare this capability\n", name, lang)
		}
		if mustTransliterate && !hasTransliteration && (provType == TransliteratorType || provType == CombinedType) {
			fmt.Fprintf(os.Stderr, "Warning: Registering Provider %s for %s which requires transliteration but Provider doesn't declare this capability\n", name, lang)
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
		fmt.Fprintf(os.Stderr, "Warning: Language %s requires tokenization but no Provider declares this capability\n", lang)
	}
	if mustTransliterate && !hasTransliteration {
		fmt.Fprintf(os.Stderr, "Warning: Language %s requires transliteration but no Provider declares this capability\n", lang)
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

