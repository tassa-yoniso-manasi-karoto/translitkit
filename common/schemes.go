
package common

import (
	"fmt"
	"sync"
	
	"github.com/k0kubun/pp"
	"github.com/gookit/color"
)

type TranslitScheme struct {
	Name         string // e.g., "IAST", "Harvard-Kyoto"
	Description  string
	Provider     string
	NeedsDocker  bool
	NeedsScraper bool
}

// SchemeRegistry manages available transliteration schemes for languages
type SchemeRegistry struct {
	mu      sync.RWMutex
	schemes map[string][]TranslitScheme // key: ISO 639-3 language code
}

var GlobalSchemeRegistry = &SchemeRegistry{
	schemes: make(map[string][]TranslitScheme),
}

// RegisterScheme adds a transliteration scheme for a language
func RegisterScheme(languageCode string, scheme TranslitScheme) error {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return fmt.Errorf(errNotISO639, languageCode)
	}

	GlobalSchemeRegistry.mu.Lock()
	defer GlobalSchemeRegistry.mu.Unlock()

	// Initialize slice if not exists
	if _, exists := GlobalSchemeRegistry.schemes[lang]; !exists {
		GlobalSchemeRegistry.schemes[lang] = make([]TranslitScheme, 0)
	}

	// Check for duplicate scheme names
	for _, s := range GlobalSchemeRegistry.schemes[lang] {
		if s.Name == scheme.Name {
			return fmt.Errorf("scheme %s already registered for language %s", scheme.Name, lang)
		}
	}

	GlobalSchemeRegistry.schemes[lang] = append(GlobalSchemeRegistry.schemes[lang], scheme)
	return nil
}

// GetSchemes returns all available transliteration schemes for a language
func GetSchemes(languageCode string) ([]TranslitScheme, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return nil, fmt.Errorf(errNotISO639, languageCode)
	}

	GlobalSchemeRegistry.mu.RLock()
	defer GlobalSchemeRegistry.mu.RUnlock()

	schemes, exists := GlobalSchemeRegistry.schemes[lang]
	if !exists {
		return nil, fmt.Errorf("no transliteration schemes registered for language %s", lang)
	}

	return schemes, nil
}

// GetSchemeModule returns a pre-configured module for a specific transliteration scheme
func GetSchemeModule(languageCode, schemeName string) (*Module, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return nil, fmt.Errorf(errNotISO639, languageCode)
	}

	GlobalSchemeRegistry.mu.RLock()
	schemes, exists := GlobalSchemeRegistry.schemes[lang]
	GlobalSchemeRegistry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no transliteration schemes registered for language %s", lang)
	}

	var targetScheme TranslitScheme
	found := false
	for _, scheme := range schemes {
		if scheme.Name == schemeName {
			targetScheme = scheme
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("scheme %s not found for language %s", schemeName, lang)
	}

	module := &Module{
		Lang: lang,
	}

	// Try to get a combined provider first.
	if provider, err := getProvider(lang, CombinedType, targetScheme.Provider); err == nil {
		module.Combined = provider
		module.ProviderType = CombinedType
		// Save configuration for later application during provider initialization.
		if err := provider.SaveConfig(map[string]interface{}{
			"lang":   lang,
			"scheme": schemeName,
		}); err != nil {
			return nil, fmt.Errorf("failed to save configuration for combined provider: %w", err)
		}

		return module, nil
	}

	// If no combined provider, try separate providers.
	tokenizer, err := getProvider(lang, TokenizerType, "uniseg")
	if err != nil {
		return nil, fmt.Errorf("failed to get tokenizer: %w", err)
	}
	module.Tokenizer = tokenizer

	transliterator, err := getProvider(lang, TransliteratorType, targetScheme.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get transliterator: %w", err)
	}
	module.Transliterator = transliterator

	// Save configuration for the transliterator.
	if err := transliterator.SaveConfig(map[string]interface{}{
		"lang":   lang,
		"scheme": schemeName,
	}); err != nil {
		return nil, fmt.Errorf("failed to save configuration for transliterator: %w", err)
	}

	return module, nil
}


// GetSchemesNames returns a slice of strings with all Names of translit schemes
func GetSchemesNames(schemes []TranslitScheme) []string {
	var names []string
	for _, scheme := range schemes {
		names = append(names, scheme.Name)
	}
	return names
}


func placehold345654er() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}