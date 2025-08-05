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
	}

	if err := common.Register(Lang, th2enEntry); err != nil {
		panic(fmt.Sprintf("failed to register thai2english.com: %v", err))
	}

	// Register PyThaiNLP provider (supports both tokenizer and combined modes)
	pythainlpProvider := NewPyThaiNLPProvider()
	pythainlpEntry := common.ProviderEntry{
		Provider:     pythainlpProvider,
		Capabilities: []string{"tokenization", "transliteration"},
	}

	if err := common.Register(Lang, pythainlpEntry); err != nil {
		panic(fmt.Sprintf("failed to register pythainlp: %v", err))
	}

	registerThaiSchemes()
	setDefaultProviders()
}

func registerThaiSchemes() {
	// PyThaiNLP (lightweight mode only)
	pythainlpSchemes := []common.TranslitScheme{
		{
			Name:        "royin",
			Description: "Royal Thai General System of Transcription (RTGS)",
			Providers:   []string{"pythainlp"},
		},
		{
			Name:        "tltk",
			Description: "Thai Language Toolkit romanization",
			Providers:   []string{"pythainlp"},
		},
		{
			Name:        "lookup",
			Description: "Dictionary-based romanization with fallback",
			Providers:   []string{"pythainlp"},
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

	thai2englishSchemes := []common.TranslitScheme{
		{
			Name:         "paiboon",
			Description:  "Paiboon-esque transliteration",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "thai2english",
			Description:  "thai2english's custom transliteration system",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "rtgs",
			Description:  "Royal Thai General System of transcription",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "ipa",
			Description:  "International Phonetic Alphabet representation",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "simplified-ipa",
			Description:  "Simplified phonetic notation",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
	}

	for _, scheme := range thai2englishSchemes {
		if err := common.RegisterScheme(Lang, scheme); err != nil {
			common.Log.Warn().
				Str("pkg", Lang).
				Str("scheme", scheme.Name).
				Msg("Failed to register thai2english.com scheme")
		}
	}

}

func setDefaultProviders() {
	// Create a new provider instance for default
	pythainlpProvider := NewPyThaiNLPProvider()
	combinedEntry := common.ProviderEntry{
		Provider:     pythainlpProvider,
		Capabilities: []string{"tokenization", "transliteration"},
		Mode:         common.CombinedMode,
	}

	// Set PyThaiNLP combined as default
	if err := common.SetDefault(Lang, []common.ProviderEntry{combinedEntry}); err != nil {
		common.Log.Error().
			Err(err).
			Msg("Failed to set default provider")
	}

	common.Log.Info().
		Str("lang", Lang).
		Str("provider", "pythainlp").
		Msg("Set PyThaiNLP as default Thai provider.")
}
