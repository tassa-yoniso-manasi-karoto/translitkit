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

// ProgressCallback is a function that reports the progress of a processing operation
// current is the index of the chunk currently being processed (0-based)
// total is the total number of chunks to process
// If a provider returns 0 or math.MaxInt32 (or greater) from GetMaxQueryLen(), 
// the progress cannot be accurately tracked.
type ProgressCallback func(current, total int)

// Provider is a unified interface for all providers of any type in the library.
// It handles tokenization, transliteration, or both combined, for specific languages.
// Providers process text input and return processed tokens with linguistic annotations.
// All providers must be context-aware and support cancellation via context.
type Provider[In AnyTokenSliceWrapper, Out AnyTokenSliceWrapper] interface {
	// SaveConfig stores configuration options for later application during initialization.
	// This allows providers to maintain their configuration separately from initialization.
	// Returns an error if the configuration is invalid.
	SaveConfig(cfg map[string]interface{}) error
	
	// Init initializes the provider with a background context.
	// This is a convenience method that calls InitWithContext with context.Background().
	// Returns an error if initialization fails.
	Init() error
	
	// InitWithContext initializes the provider with the specified context.
	// The context can be used to cancel initialization or set deadlines.
	// Returns an error if initialization fails or the context is canceled.
	InitWithContext(ctx context.Context) error
	
	// InitRecreate reinitializes the provider from scratch with a background context,
	// optionally clearing any caches when noCache is true.
	// This is a convenience method that calls InitRecreateWithContext with context.Background().
	// Returns an error if reinitialization fails.
	InitRecreate(noCache bool) error
	
	// InitRecreateWithContext reinitializes the provider from scratch with the specified context,
	// optionally clearing any caches when noCache is true. This can be used to recreate
	// Docker containers or other resources.
	// Returns an error if reinitialization fails or the context is canceled.
	InitRecreateWithContext(ctx context.Context, noCache bool) error
	
	// Close releases resources used by the provider with a background context.
	// This is a convenience method that calls CloseWithContext with context.Background().
	// Returns an error if closing fails.
	Close() error
	
	// CloseWithContext releases resources used by the provider with the specified context.
	// The context can be used to cancel the closing operation or set deadlines.
	// Returns an error if closing fails or the context is canceled.
	CloseWithContext(ctx context.Context) error
	
	// ProcessFlowController processes the input tokens using the specified context.
	// This is the core processing method of the provider. It handles either raw input
	// chunks or pre-tokenized content based on the provider type.
	// The context can be used to cancel processing or set deadlines.
	// Returns processed tokens and an error if processing fails or the context is canceled.
	ProcessFlowController(ctx context.Context, input In) (Out, error)
	
	// WithProgressCallback sets a callback function to report processing progress.
	// The callback will be called with the current chunk index and total chunks
	// during processing operations. This can be used for status reporting or
	// progress bars in user interfaces.
	WithProgressCallback(callback ProgressCallback)
	
	// Name returns the unique identifier of the provider.
	// This is used for registration and lookup in the provider registry.
	Name() string
	
	// GetType returns the type of the provider (TokenizerType, TransliteratorType, or CombinedType).
	// This is used to determine the provider's capabilities and role in processing.
	GetType() ProviderType
	
	// GetMaxQueryLen returns the maximum input length the provider can handle in a single operation.
	// This is used to determine chunking strategies for large inputs.
	// A return value of 0 indicates no known limit.
	GetMaxQueryLen() int
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

// SupportsProgress checks if a provider can meaningfully report progress
// based on its GetMaxQueryLen() value. This function can be used by client code
// to determine if progress reporting is available for a given provider.
func SupportsProgress(provider Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]) bool {
	maxQueryLen := provider.GetMaxQueryLen()
	// If maxQueryLen is 0 or >= MaxInt32, the provider doesn't use chunks
	// and therefore can't report meaningful progress
	return maxQueryLen > 0 && maxQueryLen < math.MaxInt32
}

