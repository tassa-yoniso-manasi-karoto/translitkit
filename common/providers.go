package common

import (
	"fmt"
	"math"
	"context"
)

type ProviderType string

const (
	TokenizerType      ProviderType = "tokenizer"
	TransliteratorType ProviderType = "transliterator"
	CombinedType       ProviderType = "combined"
)

// Unified interface for all providers of any type
type Provider[In AnyTokenSliceWrapper, Out AnyTokenSliceWrapper] interface {
	WithContext(ctx context.Context)
	// SaveConfig just stores the config for later usage, so that
	// when we actually do .Init(), the provider can safely apply them.
	SaveConfig(cfg map[string]interface{}) error
	
	Init() error
	InitRecreate(noCache bool) error
	ProcessFlowController(input In) (Out, error)
	
	Name() string
	GetType() ProviderType
	GetMaxQueryLen() int
	Close() error
}

type LanguageProviders struct {
	Defaults        []ProviderEntry
	Tokenizers      map[string]ProviderEntry
	Transliterators map[string]ProviderEntry
	Combined        map[string]ProviderEntry
}

type ProviderEntry struct {
	Provider     Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Capabilities []string
	Type         ProviderType
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


// getQueryLenLimit returns the smallest query length limit among the provided providers.
// If no providers are given, it returns math.MaxInt64.
func getQueryLenLimit(providers ...ProviderEntry) int {
	limit := math.MaxInt64
	for _, p := range providers {
		if i := p.Provider.GetMaxQueryLen(); i > 0 && i < limit {
			limit = i
		}
	}
	return limit
}

