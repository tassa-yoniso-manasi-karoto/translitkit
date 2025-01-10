package translitkit

import (
	"fmt"
	"sync"

	iso "github.com/barbashov/iso639-3"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

var GlobalRegistry = &Registry{
	Providers: make(map[*iso.Language]LanguageProviders),
}

type Registry struct {
	mu        sync.RWMutex
	Providers map[*iso.Language]LanguageProviders
}

func Register(lang *iso.Language, provType ProviderType, name string, entry ProviderEntry) error {
	GlobalRegistry.mu.Lock()
	defer GlobalRegistry.mu.Unlock()

	// Initialize language Providers if not exists
	if _, exists := GlobalRegistry.Providers[lang]; !exists {
		GlobalRegistry.Providers[lang] = LanguageProviders{
			Tokenizers:      make(map[string]ProviderEntry),
			Transliterators: make(map[string]ProviderEntry),
			Combined:        make(map[string]ProviderEntry),
			Defaults:        DefaultProviders(make([]ProviderEntry, 0)),
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

func GetProvider(lang *iso.Language, provType ProviderType, name string) (Provider[AnyTokenSlice, AnyTokenSlice], error) {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	langProviders, exists := GlobalRegistry.Providers[lang]
	if !exists {
		return nil, fmt.Errorf("GetProvider: no Providers registered for language: %s", lang.Part3)
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

func GetDefault(lang *iso.Language) (*BaseModule, error) {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	langProviders, exists := GlobalRegistry.Providers[lang]
	// color.Blueln("langProviders currently available:")
	// pp.Println(langProviders)
	if !exists {
		return nil, fmt.Errorf("GetDefault: no Providers registered for language: %s", lang.Part3)
	}

	if len(langProviders.Defaults) == 0 {
		return nil, fmt.Errorf("no default Providers set for language: %s", lang.Part3)
	}

	module := &BaseModule{
		Lang: lang,
	}

	if langProviders.Defaults[0].Type == CombinedType {
		module.Combined = langProviders.Defaults[0].Provider
		module.ProviderType = CombinedType
	} else {
		if len(langProviders.Defaults) < 2 {
			return nil, fmt.Errorf("insufficient default Providers for separate mode")
		}
		module.Tokenizer = langProviders.Defaults[0].Provider
		module.Transliterator = langProviders.Defaults[1].Provider
	}

	return module, nil
}

func ListProviders(lang *iso.Language) map[ProviderType][]string {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	result := make(map[ProviderType][]string)

	if langProviders, exists := GlobalRegistry.Providers[lang]; exists {
		result[TokenizerType] = make([]string, 0, len(langProviders.Tokenizers))
		for name := range langProviders.Tokenizers {
			result[TokenizerType] = append(result[TokenizerType], name)
		}

		result[TransliteratorType] = make([]string, 0, len(langProviders.Transliterators))
		for name := range langProviders.Transliterators {
			result[TransliteratorType] = append(result[TransliteratorType], name)
		}

		result[CombinedType] = make([]string, 0, len(langProviders.Combined))
		for name := range langProviders.Combined {
			result[CombinedType] = append(result[CombinedType], name)
		}
	}

	return result
}

func SetDefault(lang *iso.Language, Providers []ProviderEntry) error {
	GlobalRegistry.mu.RLock()
	defer GlobalRegistry.mu.RUnlock()

	langProviders, exists := GlobalRegistry.Providers[lang]
	if !exists {
		return fmt.Errorf("SetDefault: no Providers registered for language: %s", lang.Part3)
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

func placeholder23456ui() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

/*
func HasCapability(lang *iso.Language, provType ProviderType, name string, capability string) bool {
	...
}*/
