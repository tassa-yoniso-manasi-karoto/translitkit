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
	// NOTE: PyThaiNLPProvider OWNS the Docker container lifecycle - see pythainlp.go
	pythainlpProvider := NewPyThaiNLPProvider()
	pythainlpEntry := common.ProviderEntry{
		Provider:     pythainlpProvider,
		Capabilities: []string{"tokenization", "transliteration"},
	}

	if err := common.Register(Lang, pythainlpEntry); err != nil {
		panic(fmt.Sprintf("failed to register pythainlp: %v", err))
	}

	// Register Paiboonizer provider (transliterator only, ~83% accuracy, experimental)
	// NOTE: PaiboonizerProvider does NOT own any Docker container - it reuses
	// the container started by PyThaiNLPProvider via package-level functions.
	// See paiboonizer.go for lifecycle details.
	paiboonizerProvider := NewPaiboonizerProvider()
	paiboonizerEntry := common.ProviderEntry{
		Provider:     paiboonizerProvider,
		Capabilities: []string{"transliteration"},
	}

	if err := common.Register(Lang, paiboonizerEntry); err != nil {
		panic(fmt.Sprintf("failed to register paiboonizer: %v", err))
	}

	registerThaiSchemes()
	setDefaultProviders()
}

func registerThaiSchemes() {
	// ==========================================================================
	// HYBRID SCHEME: PyThaiNLP tokenizer + Paiboonizer transliterator
	// ==========================================================================
	// This scheme uses pythainlp for word tokenization, then paiboonizer for
	// Paiboon-style romanization. It's experimental with ~83% accuracy but
	// runs fully locally without external API dependencies.
	//
	// LIFECYCLE: pythainlp provider MUST be initialized first (starts Docker
	// container). Paiboonizer then reuses the same container via package-level
	// go-pythainlp functions. See pythainlp.go and paiboonizer.go for details.
	// ==========================================================================
	hybridScheme := common.TranslitScheme{
		Name:        "paiboon-hybrid",
		Description: "Paiboon (exp.🧪, accuracy ~95%, local, fast)",
		Providers:   []string{"pythainlp", "paiboonizer"},
		NeedsDocker: true,
	}

	if err := common.RegisterScheme(Lang, hybridScheme); err != nil {
		common.Log.Warn().
			Str("pkg", Lang).
			Str("scheme", hybridScheme.Name).
			Msg("Failed to register hybrid paiboonizer scheme")
	}

	// PyThaiNLP (lightweight mode only)
	pythainlpSchemes := []common.TranslitScheme{
		{
			Name:        "royin",
			Description: "Royal Thai General System of Transcription (pythainlp)",
			Providers:   []string{"pythainlp"},
			NeedsDocker: true,
		},
		{
			Name:        "tltk",
			Description: "Thai Language Toolkit romanization (pythainlp)",
			Providers:   []string{"pythainlp"},
			NeedsDocker: true,
		},
		{
			Name:        "lookup",
			Description: "Dictionary-based romanization with fallback (pythainlp)",
			Providers:   []string{"pythainlp"},
			NeedsDocker: true,
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
			Description:  "Paiboon-esque transliteration (thai2english.com)",
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
			Description:  "Royal Thai General System of Transcription (thai2english.com)",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "ipa",
			Description:  "International Phonetic Alphabet representation (thai2english.com)",
			Providers:    []string{"thai2english.com"},
			NeedsScraper: true,
		},
		{
			Name:         "simplified-ipa",
			Description:  "Simplified phonetic notation (thai2english.com)",
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
	// Use paiboon-hybrid as default: pythainlp for tokenization, paiboonizer for transliteration
	// Even if not 100% accurate, it is faster than th2en's paiboon and produces more learner-friendly output than pythainlp's RTGS
	pythainlpProvider := NewPyThaiNLPProvider()
	tokenizerEntry := common.ProviderEntry{
		Provider:     pythainlpProvider,
		Capabilities: []string{"tokenization"},
	}

	paiboonizerProvider := NewPaiboonizerProvider()
	transliteratorEntry := common.ProviderEntry{
		Provider:     paiboonizerProvider,
		Capabilities: []string{"transliteration"},
	}

	// Set paiboon-hybrid (pythainlp + paiboonizer) as default
	if err := common.SetDefault(Lang, []common.ProviderEntry{tokenizerEntry, transliteratorEntry}); err != nil {
		common.Log.Error().
			Err(err).
			Msg("Failed to set default provider")
	}

	common.Log.Info().
		Str("lang", Lang).
		Str("scheme", "paiboon-hybrid").
		Msg("Set paiboon-hybrid as default Thai provider.")
}
