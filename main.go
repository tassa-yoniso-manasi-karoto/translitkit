//go:generate go run generator/main.go

package translitkit

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	// language-specific pkg must be initialized for their providers to be available
	_ "github.com/tassa-yoniso-manasi-karoto/translitkit/lang/mul"
	_ "github.com/tassa-yoniso-manasi-karoto/translitkit/lang/jpn"
	_ "github.com/tassa-yoniso-manasi-karoto/translitkit/lang/tha"
	_ "github.com/tassa-yoniso-manasi-karoto/translitkit/lang/hin"
	_ "github.com/tassa-yoniso-manasi-karoto/translitkit/lang/rus"
)

// DefaultModule returns a new Module configured with the default providers
// for the specified language. The language code can be in any ISO 639 format
// (639-1, 639-2/T, 639-2/B, or 639-3).
//
// Example:
//
//	module, err := translitkit.DefaultModule("ja")  // ISO 639-1
//	module, err := translitkit.DefaultModule("jpn") // ISO 639-3
func DefaultModule(lang string) (*common.Module, error) {
	return common.DefaultModule(lang)
}

// NewModule creates a Module for the specified language using named providers.
// The language code can be in any ISO 639 format.
// For a combined provider, specify one name.
// For separate providers, specify two names in the order: tokenizer, transliterator.
//
// Example:
//
//	module, err := translitkit.NewModule("jpn", "ichiran")           // combined provider
//	module, err := translitkit.NewModule("jpn", "mecab", "kakasi")   // separate providers
func NewModule(lang string, providerNames ...string) (*common.Module, error) {
	return common.NewModule(lang, providerNames...)
}

// NeedsTokenization returns true if the given language doesn't use spaces
// to separate words and requires tokenization.
// The language code can be in any ISO 639 code format.
func NeedsTokenization(lang string) (bool, error) {
	return common.NeedsTokenization(lang)
}

// NeedsTransliteration returns true if the given language doesn't use
// the roman script and requires transliteration.
// The language code can be in any ISO 639 code format.
func NeedsTransliteration(lang string) (bool, error) {
	return common.NeedsTransliteration(lang)
}

// IsValidLanguage checks if the given language code is a valid ISO 639 code
// (in any format: 639-1, 639-2/T, 639-2/B, or 639-3).
// It returns the standardized ISO 639-3 code and true if valid.
func IsValidLanguage(lang string) (string, bool) {
	return common.IsValidISO639(lang)
}
