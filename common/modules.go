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
	ProviderType     ProviderType
	Tokenizer        Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Transliterator   Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
	Combined         Provider[AnyTokenSliceWrapper, AnyTokenSliceWrapper]
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
		if Provider, err := getProvider(lang, CombinedType, providerNames[0]); err == nil {
			module.Combined = Provider
			module.ProviderType = CombinedType
			module.chunkifier = NewChunkifier(module.getMaxQueryLen())
			return module, nil
		}
		return nil, fmt.Errorf("single Provider %s not found as combined Provider for language %s", providerNames[0], lang)
	}

	if len(providerNames) == 2 {
		// Get tokenizer
		tokenizer, err := getProvider(lang, TokenizerType, providerNames[0])
		if err != nil {
			return nil, fmt.Errorf("tokenizer %s not found: %w", providerNames[0], err)
		}
		
		// Get transliterator
		transliterator, err := getProvider(lang, TransliteratorType, providerNames[1])
		if err != nil {
			return nil, fmt.Errorf("transliterator %s not found: %w", providerNames[1], err)
		}

		module.Tokenizer = tokenizer
		module.Transliterator = transliterator
		module.chunkifier = NewChunkifier(module.getMaxQueryLen())
		return module, nil
	}

	return nil, fmt.Errorf("invalid number of Provider names: expected 1 or 2, got %d", len(providerNames))
}


func newModule() *Module {
	return &Module{ ctx:  context.Background()}
}

// ProviderNames returns the names of the provider(s) contained in the module.
// For combined providers, it returns a single name.
// For separate providers, it returns both tokenizer and transliterator names.
func (m *Module) ProviderNames() string {
	if m.Combined != nil {
		return m.Combined.Name()
	}
	
	names := make([]string, 0, 2)
	if m.Tokenizer != nil {
		names = append(names, m.Tokenizer.Name())
	}
	if m.Transliterator != nil {
		names = append(names, m.Transliterator.Name())
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
	
	// Pass the callback to the appropriate provider(s)
	if m.Combined != nil {
		m.Combined.WithProgressCallback(callback)
	} else {
		// For separate providers, pass it to the transliterator
		// as it's usually the one doing the chunked processing
		if m.Transliterator != nil {
			m.Transliterator.WithProgressCallback(callback)
		}
		if m.Tokenizer != nil {
			m.Tokenizer.WithProgressCallback(callback)
		}
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
		if m.Combined != nil {
			m.Combined.WithProgressCallback(m.progressCallback)
		} else {
			if m.Tokenizer != nil {
				m.Tokenizer.WithProgressCallback(m.progressCallback)
			}
			if m.Transliterator != nil {
				m.Transliterator.WithProgressCallback(m.progressCallback)
			}
		}
	}

	if m.Combined != nil {
		return m.Combined.InitWithContext(ctx)
	}

	if err := m.Tokenizer.InitWithContext(ctx); err != nil {
		return fmt.Errorf("tokenizer init failed: %w", err)
	}

	if err := m.Transliterator.InitWithContext(ctx); err != nil {
		return fmt.Errorf("transliterator init failed: %w", err)
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
		if m.Combined != nil {
			m.Combined.WithProgressCallback(m.progressCallback)
		} else {
			if m.Tokenizer != nil {
				m.Tokenizer.WithProgressCallback(m.progressCallback)
			}
			if m.Transliterator != nil {
				m.Transliterator.WithProgressCallback(m.progressCallback)
			}
		}
	}

	if m.Combined != nil {
		return m.Combined.InitRecreateWithContext(ctx, noCache)
	}

	if err := m.Tokenizer.InitRecreateWithContext(ctx, noCache); err != nil {
		return fmt.Errorf("tokenizer InitRecreate failed: %w", err)
	}

	if err := m.Transliterator.InitRecreateWithContext(ctx, noCache); err != nil {
		return fmt.Errorf("transliterator InitRecreate failed: %w", err)
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

	if m.Combined != nil {
		tsw, err = m.Combined.ProcessFlowController(ctx, tsw)
		if err != nil {
			return &TknSliceWrapper{}, fmt.Errorf("combined processing failed: %w", err)
		}
	} else {
		tsw, err = m.Tokenizer.ProcessFlowController(ctx, tsw)
		if err != nil {
			return &TknSliceWrapper{}, fmt.Errorf("tokenization failed: %w", err)
		}
		if m.Transliterator != nil {
			if tsw, err = m.Transliterator.ProcessFlowController(ctx, tsw); err != nil {
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
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
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
	if m.Transliterator == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("romanization requires either a transliterator or combined provider (got %s)", m.ProviderType)
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
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return "", fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
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
	if m.Tokenizer == nil && m.ProviderType != CombinedType {
		return nil, fmt.Errorf("tokenization requires either a tokenizer or combined provider (got %s)", m.ProviderType)
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
	if m.Combined != nil {
		return m.Combined.CloseWithContext(ctx)
	}
	if err := m.Tokenizer.CloseWithContext(ctx); err != nil {
		return fmt.Errorf("tokenizer close failed: %w", err)
	}
	if err := m.Transliterator.CloseWithContext(ctx); err != nil {
		return fmt.Errorf("transliterator close failed: %w", err)
	}
	return nil
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
// For combined providers, it returns the provider's limit.
// For separate providers, it returns the smallest limit between tokenizer and transliterator.
// If MaxQueryLen is already set, returns that value instead of recalculating.
func (m *Module) getMaxQueryLen() int {
	providers, err := m.listProviders()
	if err != nil {
		return math.MaxInt64
	}

	return getQueryLenLimit(providers...)
}

// SupportsProgress checks if this module's providers can report progress during processing.
// Returns true if at least one provider supports progress reporting, false otherwise.
func (m *Module) SupportsProgress() bool {
	if m.Combined != nil {
		return SupportsProgress(m.Combined)
	}
	
	// For separate providers, check if the transliterator supports progress
	// since it's usually the one doing the chunked processing
	if m.Transliterator != nil {
		return SupportsProgress(m.Transliterator)
	}
	
	if m.Tokenizer != nil {
		return SupportsProgress(m.Tokenizer)
	}
	
	return false
}

func (m *Module) setProviders(providers []ProviderEntry) error {
	if len(providers) == 0 {
		return fmt.Errorf("cannot set empty providers")
	}

	if providers[0].Type == CombinedType {
		// For combined provider, only one entry is needed
		if len(providers) > 1 {
			return fmt.Errorf("combined provider cannot be used with other providers")
		}
		m.Combined = providers[0].Provider
		m.ProviderType = CombinedType
	} else {
		// For separate providers, tokenizer is required but transliterator is optional
		if providers[0].Type != TokenizerType {
			return fmt.Errorf("first provider must be a tokenizer")
		}
		m.Tokenizer = providers[0].Provider

		// Set transliterator if provided
		if len(providers) > 1 {
			if providers[1].Type != TransliteratorType {
				return fmt.Errorf("second provider must be a transliterator")
			}
			m.Transliterator = providers[1].Provider
		}
	}
	return nil
}

func (m *Module) listProviders() (providers []ProviderEntry, err error) {
	if m.Combined != nil {
		// For combined provider, return single entry
		providers = append(providers, ProviderEntry{
			Provider: m.Combined,
			Type:     CombinedType,
		})
		return providers, nil
	}

	// For separate providers, return both tokenizer and transliterator
	if m.Tokenizer != nil {
		providers = append(providers, ProviderEntry{
			Provider: m.Tokenizer,
			Type:     TokenizerType,
		})
	}

	if m.Transliterator != nil {
		providers = append(providers, ProviderEntry{
			Provider: m.Transliterator,
			Type:     TransliteratorType,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers found in module")
	}

	return providers, nil
}


func placeholder3456456543() {
	fmt.Println("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
