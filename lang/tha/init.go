package tha

import (
	"fmt"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

func init() {
	// Register thai2english.com provider
	th2enProvider := &TH2ENProvider{}
	th2enEntry := common.ProviderEntry{
		Provider:     th2enProvider,
		Capabilities: []string{"tokenization", "transliteration"},
		Type:         common.CombinedType,
	}
	
	if err := common.Register(Lang, common.CombinedType, th2enProvider.Name(), th2enEntry); err != nil {
		panic(fmt.Sprintf("failed to register thai2english.com: %v", err))
	}
	
	// Register PyThaiNLP as tokenizer only
	pythainlpTokenizer := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
	tokenizerEntry := common.ProviderEntry{
		Provider:     pythainlpTokenizer,
		Capabilities: []string{"tokenization"},
		Type:         common.TokenizerType,
	}
	
	if err := common.Register(Lang, common.TokenizerType, pythainlpTokenizer.Name(), tokenizerEntry); err != nil {
		panic(fmt.Sprintf("failed to register pythainlp-tokenizer: %v", err))
	}
	
	// Register PyThaiNLP as combined provider
	pythainlpCombined := &PyThaiNLPProvider{operatingMode: common.CombinedType}
	combinedEntry := common.ProviderEntry{
		Provider:     pythainlpCombined,
		Capabilities: []string{"tokenization", "transliteration"},
		Type:         common.CombinedType,
	}
	
	if err := common.Register(Lang, common.CombinedType, pythainlpCombined.Name(), combinedEntry); err != nil {
		panic(fmt.Sprintf("failed to register pythainlp: %v", err))
	}
	
	// Register all Thai transliteration schemes
	registerThaiSchemes()
	
	// Set default: PyThaiNLP tokenizer + thai2english.com transliterator
	// This gives the best of both worlds: reliable local tokenization + 
	// high-quality web-based transliteration
	setDefaultProviders()
}

func registerThaiSchemes() {
	// Register thai2english.com schemes (moved from th2en.go)
	thai2englishSchemes := []common.TranslitScheme{
		{Name: "paiboon", Description: "Paiboon-esque transliteration", Provider: "thai2english.com", NeedsScraper: true},
		{Name: "thai2english", Description: "thai2english's custom transliteration system", Provider: "thai2english.com", NeedsScraper: true},
		{Name: "rtgs", Description: "Royal Thai General System of transcription", Provider: "thai2english.com", NeedsScraper: true},
		{Name: "ipa", Description: "International Phonetic Alphabet representation", Provider: "thai2english.com", NeedsScraper: true},
		{Name: "simplified-ipa", Description: "Simplified phonetic notation", Provider: "thai2english.com", NeedsScraper: true},
	}
	
	for _, scheme := range thai2englishSchemes {
		if err := common.RegisterScheme(Lang, scheme); err != nil {
			common.Log.Warn().
				Str("pkg", Lang).
				Str("scheme", scheme.Name).
				Msg("Failed to register thai2english.com scheme")
		}
	}
	
	// Register PyThaiNLP romanization schemes
	pythainlpSchemes := []common.TranslitScheme{
		{
			Name:        "royin",
			Description: "Royal Thai General System of Transcription (RTGS)",
			Provider:    "pythainlp",
		},
		{
			Name:        "thai2rom",
			Description: "Deep learning-based Thai romanization",
			Provider:    "pythainlp",
		},
		{
			Name:        "tltk",
			Description: "Thai Language Toolkit romanization",
			Provider:    "pythainlp",
		},
		{
			Name:        "lookup",
			Description: "Dictionary-based romanization with fallback",
			Provider:    "pythainlp",
		},
	}
	
	for _, scheme := range pythainlpSchemes {
		if err := common.RegisterScheme(Lang, scheme); err != nil {
			common.Log.Warn().
				Str("pkg", Lang).
				Str("scheme", scheme.Name).
				Msg("Failed to register PyThaiNLP scheme")
		}
	}
}

func setDefaultProviders() {
	// Create fresh instances for default configuration
	// Use PyThaiNLP tokenizer + thai2english.com transliterator as default
	
	// pythainlpTokenizer := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
	// tokenizerEntry := common.ProviderEntry{
	// 	Provider:     pythainlpTokenizer,
	// 	Capabilities: []string{"tokenization"},
	// 	Type:         common.TokenizerType,
	// }
	
	// Note: We need to check if thai2english.com is registerable as a transliterator
	// It's currently only registered as CombinedType, but we want to use it as transliterator
	// For now, we'll set the combined pythainlp as default and let users configure as needed
	
	pythainlpCombined := &PyThaiNLPProvider{operatingMode: common.CombinedType}
	combinedEntry := common.ProviderEntry{
		Provider:     pythainlpCombined,
		Capabilities: []string{"tokenization", "transliteration"},
		Type:         common.CombinedType,
	}
	
	// Set PyThaiNLP combined as default (users can override to use pythainlp-tokenizer + thai2english)
	if err := common.SetDefault(Lang, []common.ProviderEntry{combinedEntry}); err != nil {
		common.Log.Error().
			Err(err).
			Msg("Failed to set default provider")
	}
	
	common.Log.Info().
		Str("lang", Lang).
		Str("provider", "pythainlp").
		Msg("Set PyThaiNLP as default Thai provider. Users can configure pythainlp-tokenizer + thai2english.com for hybrid approach")
}

