package common

import (
	"fmt"
	"strings"
	"math"
	"context"

	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	//iso "github.com/barbashov/iso639-3"
)

// Private because NOT NEEDED/IN USE AT THIS POINT.
// Could become needed of more sophisticated NLP providers are implemented.
// Method set needs more iterations to be defined.
type anyModule interface {
	Init() error
	InitRecreate(bool) error
	MustInit()
	ProviderNames() string
	RomanPostProcess(string, func(string) string) string
	Close() error
	
	InitWithContext(context.Context) error
	InitRecreateWithContext(context.Context, bool) error
	MustInitWithContext(context.Context)
	CloseWithContext(context.Context) error
	
	// getMaxQueryLen() int ?
	setProviders([]ProviderEntry) error
}

// Module satisfies the anyModule interface.
// It contains both Tokenization+Transliteration components.

type Module struct {
	ctx              context.Context
	Lang             string // ISO-639 Part 3: i.e. "eng", "zho", "jpn"...
	Providers        []Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	ProviderRoles    map[OperatingMode]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	progressCallback ProgressCallback
	chunkifier       *Chunkifier
}

// NewModule creates a Module for the specified language using either default Providers
// or the explicitly named ones. If providerNames is empty, default Providers are used.
// For a combined Provider, specify one name. For separate Providers, specify two names
// in the order: tokenizer, transliterator.
//
// Example usage:
//
//	module, err := NewModule("jpn") // Use defaults
//	module, err := NewModule("jpn", "ichiran") // Use combined Provider
//	module, err := NewModule("jpn", "mecab", "kakasi") // Use separate Providers
func NewModule(languageCode string, providerNames ...string) (*Module, error) {
	lang, ok := IsValidISO639(languageCode)
	if !ok {
		return nil, fmt.Errorf(errNotISO639, languageCode)
	}
	if len(providerNames) == 0 {
		return DefaultModule(lang)
	}

	module := newModule()
	module.Lang = lang

	if len(providerNames) == 1 {
		// Try to get as combined Provider
		if provider, err := getProvider(lang, CombinedMode, providerNames[0]); err == nil {
			module.Providers = append(module.Providers, provider)
			module.ProviderRoles[CombinedMode] = provider
			module.chunkifier = NewChunkifier(module.getMaxQueryLen())
			return module, nil
		}
		return nil, fmt.Errorf("single Provider %s not found as combined Provider for language %s", providerNames[0], lang)
	}

	if len(providerNames) == 2 {
		// Get tokenizer
		tokenizer, err := getProvider(lang, TokenizerMode, providerNames[0])
		if err != nil {
			return nil, fmt.Errorf("tokenizer %s not found: %w", providerNames[0], err)
		}
		
		// Get transliterator
		transliterator, err := getProvider(lang, TransliteratorMode, providerNames[1])
		if err != nil {
			return nil, fmt.Errorf("transliterator %s not found: %w", providerNames[1], err)
		}

		module.Providers = append(module.Providers, tokenizer)
		module.Providers = append(module.Providers, transliterator)
		module.ProviderRoles[TokenizerMode] = tokenizer
		module.ProviderRoles[TransliteratorMode] = transliterator
		module.chunkifier = NewChunkifier(module.getMaxQueryLen())
		return module, nil
	}

	return nil, fmt.Errorf("invalid number of Provider names: expected 1 or 2, got %d", len(providerNames))
}


func newModule() *Module {
	return &Module{
		ctx:           context.Background(),
		Providers:     make([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], 0),
		ProviderRoles: make(map[OperatingMode]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]),
	}
}

// getTokenizer returns the provider that handles tokenization
func (m *Module) getTokenizer() Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper] {
	if p, ok := m.ProviderRoles[CombinedMode]; ok {
		return p
	}
	return m.ProviderRoles[TokenizerMode]
}

// getTransliterator returns the provider that handles transliteration
func (m *Module) getTransliterator() Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper] {
	if p, ok := m.ProviderRoles[CombinedMode]; ok {
		return p
	}
	return m.ProviderRoles[TransliteratorMode]
}

// hasTokenizer returns true if the module has tokenization capability
func (m *Module) hasTokenizer() bool {
	_, hasCombined := m.ProviderRoles[CombinedMode]
	_, hasTokenizer := m.ProviderRoles[TokenizerMode]
	return hasCombined || hasTokenizer
}

// hasTransliterator returns true if the module has transliteration capability
func (m *Module) hasTransliterator() bool {
	_, hasCombined := m.ProviderRoles[CombinedMode]
	_, hasTransliterator := m.ProviderRoles[TransliteratorMode]
	return hasCombined || hasTransliterator
}

// ProviderNames returns the names of the provider(s) contained in the module.
// For combined providers, it returns a single name.
// For separate providers, it returns both tokenizer and transliterator names.
func (m *Module) ProviderNames() string {
	names := make([]string, 0, len(m.Providers))
	for _, p := range m.Providers {
		names = append(names, p.Name())
	}
	return strings.Join(names, "→")
}

// WithProgressCallback sets a callback function to track progress of processing operations.
// The callback will be called with the current chunk index and total chunks.
// This is useful for displaying progress bars or status updates during long-running
// operations.
//
// Returns the module for method chaining.
func (m *Module) WithProgressCallback(callback ProgressCallback) *Module {
	m.progressCallback = callback
	
	// Pass the callback to all providers
	for _, provider := range m.Providers {
		provider.WithProgressCallback(callback)
	}
	
	return m
}

// The default chunkifier is optimized for best performance but there is a case for
// using a custom chunkifier if you want smaller chunks in order to induce frequent  
// progress callbacks or if your language has some special requirements (in that case
// you may also open an issue on github).
func (m *Module) WithCustomChunkifier(chunkifier *Chunkifier) *Module {
	m.chunkifier = chunkifier
	return m
}

// serialize breaks the input text into chunks based on the maximum query length
// and returns a token slice wrapper containing the raw chunks.
// The number of chunks can be obtained by checking len(wrapper.GetRaw())
func (m *Module) serialize(input string, max int) (AnyTokenSliceWrapper, error) {
	chunks, err := m.chunkifier.Chunkify(input)
	return &TknSliceWrapper{Raw: chunks}, err
}


// InitWithContext initializes the module and its providers using the provided context.
// This allows cancellation during the initialization process.
// The module will pass the context to the appropriate providers and also set up any
// progress callbacks that have been registered.
//
// Returns an error if initialization fails or the context is canceled.
func (m *Module) InitWithContext(ctx context.Context) error {
	// Pass progress callback if set
	if m.progressCallback != nil {
		for _, provider := range m.Providers {
			provider.WithProgressCallback(m.progressCallback)
		}
	}

	// Initialize all providers
	for _, provider := range m.Providers {
		if err := provider.InitWithContext(ctx); err != nil {
			return fmt.Errorf("provider %s init failed: %w", provider.Name(), err)
		}
	}

	return nil
}

// Init initializes the module and its providers using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if initialization fails.
func (m *Module) Init() error {
	return m.InitWithContext(context.Background())
}

// InitRecreateWithContext forces reinitialization of the module's providers with the provided context.
// This can be used to recreate Docker containers or other resources.
// When noCache is true, caches will be cleared during reinitialization.
//
// Returns an error if reinitialization fails or the context is canceled.
func (m *Module) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	// Pass progress callback if set
	if m.progressCallback != nil {
		for _, provider := range m.Providers {
			provider.WithProgressCallback(m.progressCallback)
		}
	}

	// Reinitialize all providers
	for _, provider := range m.Providers {
		if err := provider.InitRecreateWithContext(ctx, noCache); err != nil {
			return fmt.Errorf("provider %s InitRecreate failed: %w", provider.Name(), err)
		}
	}

	return nil
}

// InitRecreate forces reinitialization of the module's providers using a background context.
// This is a convenience method for operations that don't need cancellation control.
// When noCache is true, caches will be cleared during reinitialization.
//
// Returns an error if reinitialization fails.
func (m *Module) InitRecreate(noCache bool) error {
	return m.InitRecreateWithContext(context.Background(), noCache)
}

// MustInitWithContext initializes the module with the given context and panics on error.
// This is useful for initialization in program startup when failure to initialize is fatal.
// The context allows for cancellation during initialization.
func (m *Module) MustInitWithContext(ctx context.Context) {
	if err := m.InitRecreateWithContext(ctx, false); err != nil {
		panic(err)
	}
}

// MustInit initializes the module with a background context and panics on error.
// This is a convenience method for operations that don't need cancellation control.
func (m *Module) MustInit() {
	m.MustInitWithContext(context.Background())
}

// TokensWithContext processes the input text with the provided context and returns token analysis.
// It breaks the input into tokens and performs both tokenization and transliteration if appropriate
// for the language and provider type. The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be processed
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails or the context is canceled
func (m *Module) TokensWithContext(ctx context.Context, input string) (AnyTokenSliceWrapper, error) {
	tsw, err := m.serialize(input, m.getMaxQueryLen())
	if err != nil {
		return nil, fmt.Errorf("input serialization failed: len(input)=%d, %w", len(input), err)
	}

	// Check if we have a combined provider
	if combined, ok := m.ProviderRoles[CombinedMode]; ok {
		tsw, err = combined.ProcessFlowController(ctx, CombinedMode, tsw)
		if err != nil {
			return &TknSliceWrapper{}, fmt.Errorf("combined processing failed: %w", err)
		}
	} else {
		// Process with separate providers
		if tokenizer, ok := m.ProviderRoles[TokenizerMode]; ok {
			tsw, err = tokenizer.ProcessFlowController(ctx, TokenizerMode, tsw)
			if err != nil {
				return &TknSliceWrapper{}, fmt.Errorf("tokenization failed: %w", err)
			}
		} else {
			return &TknSliceWrapper{}, fmt.Errorf("no tokenizer available")
		}
		
		// Transliteration is optional
		if transliterator, ok := m.ProviderRoles[TransliteratorMode]; ok {
			if tsw, err = transliterator.ProcessFlowController(ctx, TransliteratorMode, tsw); err != nil {
				return &TknSliceWrapper{}, fmt.Errorf("transliteration failed: %w", err)
			}
		}
	}
	
	if tsw == nil {
		return tsw, fmt.Errorf("fatal: nil tokens returned by module: %#v", m)
	}
	return tsw, nil
}

// Tokens processes the input text using a background context and returns token analysis.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be processed
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails
func (m *Module) Tokens(input string) (AnyTokenSliceWrapper, error) {
	return m.TokensWithContext(context.Background(), input)
}

// LexicalTokensWithContext returns only tokens containing lexical content with the provided context.
// Lexical tokens are words and meaningful language units, excluding punctuation and spaces.
// The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be processed
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing only lexical tokens
//   - error: An error if processing fails or the context is canceled
func (m *Module) LexicalTokensWithContext(ctx context.Context, input string) (AnyTokenSliceWrapper, error) {
	raw, err := m.TokensWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return ToAnyLexicalTokens(raw), nil
}

// LexicalTokens returns only tokens containing lexical content using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be processed
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing only lexical tokens
//   - error: An error if processing fails
func (m *Module) LexicalTokens(input string) (AnyTokenSliceWrapper, error) {
	return m.LexicalTokensWithContext(context.Background(), input)
}

// RomanWithContext returns the input text romanized (transliterated) with the provided context.
// The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be romanized
//
// Returns:
//   - string: The romanized text
//   - error: An error if processing fails, the context is canceled, or romanization isn't supported
func (m *Module) RomanWithContext(ctx context.Context, input string) (string, error) {
	if !m.hasTransliterator() {
		return "", fmt.Errorf("romanization requires a provider with transliteration capability")
	}
	tkns, err := m.TokensWithContext(ctx, input)
	if err != nil {
		return "", err
	}
	return tkns.Roman(), nil
}

// Roman returns the input text romanized (transliterated) using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be romanized
//
// Returns:
//   - string: The romanized text
//   - error: An error if processing fails or romanization isn't supported
func (m *Module) Roman(input string) (string, error) {
	return m.RomanWithContext(context.Background(), input)
}

// RomanPartsWithContext returns an array of romanized word parts with the provided context.
// This method only returns the lexical tokens (words), not spaces or punctuation.
// The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be processed
//
// Returns:
//   - []string: An array of romanized word parts
//   - error: An error if processing fails, the context is canceled, or romanization isn't supported
func (m *Module) RomanPartsWithContext(ctx context.Context, input string) ([]string, error) {
	if !m.hasTransliterator() {
		return nil, fmt.Errorf("romanization requires a provider with transliteration capability")
	}
	tkns, err := m.LexicalTokensWithContext(ctx, input)
	if err != nil {
		return []string{}, err
	}
	return tkns.RomanParts(), nil
}

// RomanParts returns an array of romanized word parts using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be processed
//
// Returns:
//   - []string: An array of romanized word parts
//   - error: An error if processing fails or romanization isn't supported
func (m *Module) RomanParts(input string) ([]string, error) {
	return m.RomanPartsWithContext(context.Background(), input)
}

// TokenizedWithContext returns the input text tokenized with the provided context.
// Tokenization breaks the text into individual linguistic units with appropriate spacing.
// The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be tokenized
//
// Returns:
//   - string: The tokenized text
//   - error: An error if processing fails, the context is canceled, or tokenization isn't supported
func (m *Module) TokenizedWithContext(ctx context.Context, input string) (string, error) {
	if !m.hasTokenizer() {
		return "", fmt.Errorf("tokenization requires a provider with tokenization capability")
	}
	tkns, err := m.TokensWithContext(ctx, input)
	if err != nil {
		return "", err
	}
	return tkns.Tokenized(), nil
}

// Tokenized returns the input text tokenized using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be tokenized
//
// Returns:
//   - string: The tokenized text
//   - error: An error if processing fails or tokenization isn't supported
func (m *Module) Tokenized(input string) (string, error) {
	return m.TokenizedWithContext(context.Background(), input)
}

// TokenizedPartsWithContext returns an array of tokenized word parts with the provided context.
// This method only returns the lexical tokens (words), not spaces or punctuation.
// The context allows cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The text to be processed
//
// Returns:
//   - []string: An array of tokenized word parts
//   - error: An error if processing fails, the context is canceled, or tokenization isn't supported
func (m *Module) TokenizedPartsWithContext(ctx context.Context, input string) ([]string, error) {
	if !m.hasTokenizer() {
		return nil, fmt.Errorf("tokenization requires a provider with tokenization capability")
	}
	tkns, err := m.LexicalTokensWithContext(ctx, input)
	if err != nil {
		return []string{}, err
	}
	return tkns.TokenizedParts(), nil
}

// TokenizedParts returns an array of tokenized word parts using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Parameters:
//   - input: The text to be processed
//
// Returns:
//   - []string: An array of tokenized word parts
//   - error: An error if processing fails or tokenization isn't supported
func (m *Module) TokenizedParts(input string) ([]string, error) {
	return m.TokenizedPartsWithContext(context.Background(), input)
}

// CloseWithContext closes the module and its providers with the provided context.
// This releases any resources used by the module and its providers, such as
// database connections or containerized services.
// The context allows cancellation during closing.
//
// Returns an error if closing fails or the context is canceled.
func (m *Module) CloseWithContext(ctx context.Context) error {
	var lastErr error
	// Close all providers, collecting errors
	for _, provider := range m.Providers {
		if err := provider.CloseWithContext(ctx); err != nil {
			lastErr = fmt.Errorf("provider %s close failed: %w", provider.Name(), err)
		}
	}
	return lastErr
}

// Close closes the module and its providers using a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if closing fails.
func (m *Module) Close() error {
	return m.CloseWithContext(context.Background())
}

func (m *Module) RomanPostProcess(s string, f func(string) (string)) (string) {
	return f(s)
}

// getMaxQueryLen returns the maximum query length that can be processed by the module.
// It returns the smallest limit among all providers.
func (m *Module) getMaxQueryLen() int {
	limit := math.MaxInt64
	for _, p := range m.Providers {
		if i := p.GetMaxQueryLen(); i > 0 && i < limit {
			limit = i
		}
	}
	return limit
}

// SupportsProgress checks if this module's providers can report progress during processing.
// Returns true if at least one provider supports progress reporting, false otherwise.
func (m *Module) SupportsProgress() bool {
	for _, provider := range m.Providers {
		if SupportsProgress(provider) {
			return true
		}
	}
	return false
}

// validateProviderSetup validates that providers are suitable for a language
func validateProviderSetup(lang string, providers []Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]) error {
	if len(providers) == 0 {
		return fmt.Errorf("no providers specified")
	}
	
	needsTokenization, _ := NeedsTokenization(lang)
	
	// Single provider case
	if len(providers) == 1 {
		modes := providers[0].SupportedModes()
		hasCombined := false
		hasTokenizer := false
		hasTransliterator := false
		
		for _, mode := range modes {
			switch mode {
			case CombinedMode:
				hasCombined = true
			case TokenizerMode:
				hasTokenizer = true
			case TransliteratorMode:
				hasTransliterator = true
			}
		}
		
		// Combined provider is always valid
		if hasCombined {
			return nil
		}
		
		// Single transliterator is only valid if language doesn't need tokenization
		if hasTransliterator && !hasTokenizer {
			if needsTokenization {
				return fmt.Errorf("language %s requires tokenization but provider only supports transliteration", lang)
			}
			return nil
		}
		
		// Single tokenizer is valid - useful for NLP tasks that don't need transliteration
		if hasTokenizer && !hasTransliterator {
			return nil
		}
	}
	
	// Multiple providers case
	if len(providers) >= 2 {
		// First provider should typically be a tokenizer for languages that need tokenization
		firstModes := providers[0].SupportedModes()
		hasTokenizer := false
		for _, mode := range firstModes {
			if mode == TokenizerMode {
				hasTokenizer = true
				break
			}
		}
		
		// If the language needs tokenization, the first provider should support it
		if needsTokenization && !hasTokenizer {
			return fmt.Errorf("first provider should support tokenizer mode for language %s", lang)
		}
		
		// Second provider is typically a transliterator, but it's optional
		// This allows for tokenizer-only setups for future NLP tasks
		// No validation required for the second provider - it could be another tokenizer,
		// a transliterator, or any future provider type (sentiment analyzer, NER, etc.)
	}
	
	return nil
}

func (m *Module) setProviders(providers []ProviderEntry) error {
	if len(providers) == 0 {
		return fmt.Errorf("cannot set empty providers")
	}

	// Extract provider interfaces for validation
	providerInterfaces := make([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], len(providers))
	for i, entry := range providers {
		providerInterfaces[i] = entry.Provider
	}
	
	// Validate the provider setup for this language
	if err := validateProviderSetup(m.Lang, providerInterfaces); err != nil {
		return err
	}

	// Clear existing providers
	m.Providers = make([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper], 0, len(providers))
	m.ProviderRoles = make(map[OperatingMode]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper])

	// Assign providers to the module
	for _, entry := range providers {
		m.Providers = append(m.Providers, entry.Provider)
		
		// Map provider to its supported roles
		for _, mode := range entry.Provider.SupportedModes() {
			// If multiple providers support the same mode, the last one wins
			m.ProviderRoles[mode] = entry.Provider
		}
	}

	// Special handling for single transliterator without tokenizer
	if len(providers) == 1 {
		modes := providers[0].Provider.SupportedModes()
		hasOnlyTransliterator := false
		for _, mode := range modes {
			if mode == TransliteratorMode && !contains(modes, TokenizerMode) && !contains(modes, CombinedMode) {
				hasOnlyTransliterator = true
				break
			}
		}
		
		if hasOnlyTransliterator {
			// Check if language needs tokenization
			needsTokenization, _ := NeedsTokenization(m.Lang)
			if !needsTokenization {
				// Add uniseg tokenizer
				if uniseg, err := getProvider("mul", TokenizerMode, "uniseg"); err == nil {
					m.Providers = append([]Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]{uniseg}, m.Providers...)
					m.ProviderRoles[TokenizerMode] = uniseg
				}
			}
		}
	}
	
	m.chunkifier = NewChunkifier(m.getMaxQueryLen())
	return nil
}

// contains checks if a slice contains a specific mode
func contains(modes []OperatingMode, mode OperatingMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

func (m *Module) listProviders() (providers []ProviderEntry, err error) {
	if len(m.Providers) == 0 {
		return nil, fmt.Errorf("no providers found in module")
	}

	// Return all providers as ProviderEntry
	for _, p := range m.Providers {
		providers = append(providers, ProviderEntry{
			Provider: p,
		})
	}

	return providers, nil
}


func placeholder3456456543() {
	fmt.Println("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
