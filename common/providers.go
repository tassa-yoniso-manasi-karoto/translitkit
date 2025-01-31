package common

import (
	"fmt"
	"math"
)

type ProviderType string

const (
	TokenizerType      ProviderType = "tokenizer"
	TransliteratorType ProviderType = "transliterator"
	CombinedType       ProviderType = "combined"
)

// Unified interface for all providers of any type
type Provider[In AnyTokenSliceWrapper, Out AnyTokenSliceWrapper] interface {
	Init() error
	InitRecreate(noCache bool) error
	ProcessFlowController(input In) (Out, error)
	SetConfig(map[string]interface{}) error
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

