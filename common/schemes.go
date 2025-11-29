
package common

import (
	"fmt"
	"sync"
	"errors"

	"github.com/k0kubun/pp"
	"github.com/gookit/color"
)

var ErrNoSchemesRegistered = errors.New("no transliteration schemes registered for provided language")

type TranslitScheme struct {
	Name         string   // e.g., "IAST", "Harvard-Kyoto"
	Description  string
	Providers    []string // Provider names in order (tokenizer, transliterator)
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
		return nil, ErrNoSchemesRegistered
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
		return nil, ErrNoSchemesRegistered
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

	module := newModule()
	module.Lang = lang

	// Handle based on number of providers
	switch len(targetScheme.Providers) {
	case 0:
		return nil, fmt.Errorf("scheme %s has no providers configured", schemeName)
		
	case 1:
		// Single provider - try as combined first
		providerName := targetScheme.Providers[0]
		
		// Try to get as combined provider
		if provider, err := getProvider(lang, CombinedMode, providerName); err == nil {
			module.Providers = append(module.Providers, provider)
			module.ProviderRoles[CombinedMode] = provider
			module.chunkifier = NewChunkifier(module.getMaxQueryLen())
			
			// Save configuration
			if err := provider.SaveConfig(map[string]interface{}{
				"lang":   lang,
				"scheme": schemeName,
			}); err != nil {
				return nil, fmt.Errorf("failed to save configuration for combined provider: %w", err)
			}
			return module, nil
		}
		
		// Not found as combined, try as transliterator
		if provider, err := getProvider(lang, TransliteratorMode, providerName); err == nil {
			// Validate single transliterator setup
			if err := validateProviderSetup(lang, []Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]{provider}); err != nil {
				return nil, err
			}
			
			module.Providers = append(module.Providers, provider)
			module.ProviderRoles[TransliteratorMode] = provider
			
			// Use uniseg as tokenizer if language doesn't need special tokenization
			needsTokenization, _ := NeedsTokenization(lang)
			if !needsTokenization {
				tokenizer, err := getProvider("mul", TokenizerMode, "uniseg")
				if err != nil {
					return nil, fmt.Errorf("failed to get uniseg tokenizer: %w", err)
				}
				module.Providers = append([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]{tokenizer}, module.Providers...)
				module.ProviderRoles[TokenizerMode] = tokenizer
			}
			
			module.chunkifier = NewChunkifier(module.getMaxQueryLen())
			
			// Save configuration for transliterator
			if err := provider.SaveConfig(map[string]interface{}{
				"lang":   lang,
				"scheme": schemeName,
			}); err != nil {
				return nil, fmt.Errorf("failed to save configuration: %w", err)
			}
			return module, nil
		}
		
		return nil, fmt.Errorf("provider %s not found as combined or transliterator for language %s", providerName, lang)
		
	case 2:
		// Two providers - first must be tokenizer, second transliterator
		tokenizer, err := getProvider(lang, TokenizerMode, targetScheme.Providers[0])
		if err != nil {
			return nil, fmt.Errorf("first provider must be tokenizer, %s not found: %w", targetScheme.Providers[0], err)
		}
		
		transliterator, err := getProvider(lang, TransliteratorMode, targetScheme.Providers[1])
		if err != nil {
			return nil, fmt.Errorf("second provider must be transliterator, %s not found: %w", targetScheme.Providers[1], err)
		}
		
		// Validate the provider combination
		if err := validateProviderSetup(lang, []Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]{tokenizer, transliterator}); err != nil {
			return nil, err
		}
		
		module.Providers = append(module.Providers, tokenizer)
		module.Providers = append(module.Providers, transliterator)
		module.ProviderRoles[TokenizerMode] = tokenizer
		module.ProviderRoles[TransliteratorMode] = transliterator
		module.chunkifier = NewChunkifier(module.getMaxQueryLen())
		
		// Save configuration for transliterator
		if err := transliterator.SaveConfig(map[string]interface{}{
			"lang":   lang,
			"scheme": schemeName,
		}); err != nil {
			return nil, fmt.Errorf("failed to save configuration: %w", err)
		}
		return module, nil
		
	default:
		return nil, fmt.Errorf("unsupported provider configuration: %d providers", len(targetScheme.Providers))
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