package mul

import (
	"fmt"
	"math"
	"context"

	"github.com/tassa-yoniso-manasi-karoto/go-aksharamukha"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

// AksharamukhaProvider satisfies the Provider interface
type AksharamukhaProvider struct {
	config		    map[string]interface{}
	Lang		    string // ISO 639-3 language code
	targetScheme	aksharamukha.Script
	progressCallback common.ProgressCallback
}


// NewAksharamukhaProvider creates a new provider instance with the specified language
func NewAksharamukhaProvider(lang string) *AksharamukhaProvider {
	return &AksharamukhaProvider{
		Lang:   lang,
	}
}

// SaveConfig stores the configuration for later application during initialization.
// This allows the provider to be configured before being initialized.
//
// Returns an error if the configuration is invalid.
func (p *AksharamukhaProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	lang, ok := p.config["lang"].(string)
	if !ok {
		return fmt.Errorf("lang not provided in config")
	}
	p.Lang = lang
	return nil
}

// InitWithContext initializes the provider with the given context.
// This sets up the aksharamukha library and applies any stored configuration.
// The context is used for cancellation during initialization.
//
// Returns an error if initialization fails, language is not set, or the context is canceled.
func (p *AksharamukhaProvider) InitWithContext(ctx context.Context) (err error) {
	if p.Lang == "" {
		return fmt.Errorf("language code must be set before initialization")
	}

	// Pre-pull images with retry logic for slow/unreliable connections
	if err = aksharamukha.PullImagesWithContext(ctx); err != nil {
		return fmt.Errorf("failed to pull aksharamukha images: %w", err)
	}

	// Now using the context-aware version
	if err = aksharamukha.InitWithContext(ctx); err != nil {
		return fmt.Errorf("failed to initialize aksharamukha: %w", err)
	}
	p.applyConfig()
	return
}

// Init initializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if initialization fails or language is not set.
func (p *AksharamukhaProvider) Init() (err error) {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider from scratch with the given context.
// This can be used to recreate any resources and optionally clear caches when noCache is true.
// The context is used for cancellation during reinitialization.
//
// Returns an error if reinitialization fails, language is not set, or the context is canceled.
func (p *AksharamukhaProvider) InitRecreateWithContext(ctx context.Context, noCache bool) (err error) {
	if p.Lang == "" {
		return fmt.Errorf("language code must be set before initialization")
	}

	// Now using the context-aware version
	if err = aksharamukha.InitRecreateWithContext(ctx, noCache); err != nil {
		return fmt.Errorf("failed to initialize aksharamukha: %w", err)
	}
	p.applyConfig()
	return
}

// InitRecreate reinitializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if reinitialization fails or language is not set.
func (p *AksharamukhaProvider) InitRecreate(noCache bool) (err error) {
	return p.InitRecreateWithContext(context.Background(), noCache)
}


func (p *AksharamukhaProvider) applyConfig() error {
	if p.config == nil {
		return nil
	}
	schemeName, ok := p.config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}
	
	// Convert scheme name to target aksharamukha.Script
	targetScheme, ok := indicSchemesToScript[schemeName]
	if !ok {
		return fmt.Errorf("unsupported transliteration scheme: %s", schemeName)
	}

	p.targetScheme = targetScheme
	return nil
}


func (p *AksharamukhaProvider) Name() string {
	return "aksharamukha"
}

func (p *AksharamukhaProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TransliteratorMode}
}

func (p *AksharamukhaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

// CloseWithContext releases resources used by the provider with the given context.
// The context is used for cancellation during resource release.
//
// Returns an error if closing fails or the context is canceled.
func (p *AksharamukhaProvider) CloseWithContext(ctx context.Context) error {
	// Using the normal Close() for now as there's no context-aware version in the API yet
	// When aksharamukha adds CloseWithContext, we can use that instead
	return aksharamukha.Close()
}

// Close releases resources used by the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if closing fails.
func (p *AksharamukhaProvider) Close() error {
	return p.CloseWithContext(context.Background())
}

// WithProgressCallback sets a callback function for reporting progress during processing.
// The callback will be invoked with the current chunk index and total number of chunks.
func (p *AksharamukhaProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// ProcessFlowController processes input tokens using the specified context.
// This handles either raw input chunks or pre-tokenized content.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The token slice wrapper to process
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails, the context is canceled, or input format is invalid
func (p *AksharamukhaProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}
	if len(raw) != 0 {
		//switch mode {
		//case common.TransliteratorMode:
		//	return p.process(ctx, raw)
		//default:
		return nil, fmt.Errorf("operating mode %s not supported", mode)
		//}
		input.ClearRaw()
	} else {
		switch mode {
		case common.TransliteratorMode:
			return p.processTokens(ctx, input)
		default:
			return nil, fmt.Errorf("operating mode %s not supported", mode)
		}
	}
	return nil, fmt.Errorf("handling not implemented for '%s' with OperatingMode '%s'", p.Name(), mode)
}

// processTokens handles pre-tokenized input, adding romanization to tokens.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: Pre-tokenized input to process
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails or the context is canceled
func (p *AksharamukhaProvider) processTokens(ctx context.Context, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	tokens := input.(*common.TknSliceWrapper).Slice
	totalTokens := len(tokens)
	
	for idx, tkn := range tokens {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("aksharamukha: context canceled while processing token %d: %w", idx, err)
		}
		
		// Report progress if callback is set
		if p.progressCallback != nil && idx%10 == 0 { // Report every 10 tokens to avoid excessive callbacks
			p.progressCallback(idx, totalTokens)
		}
		
		s := tkn.GetSurface()
		if !tkn.IsLexicalContent() || s == "" || tkn.Roman() != "" {
			continue
		}
		romanized, err := p.romanize(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("romanization failed for token %s: %w", s, err)
		}
		tkn.SetRoman(romanized)
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(totalTokens, totalTokens)
	}
	
	return input, nil
}

// romanize converts text to a romanized form using the appropriate scheme.
// It uses either the configured scheme or falls back to the default romanization.
// Now accepts a context for cancellation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - text: The text to romanize
//
// Returns:
//   - string: The romanized text
//   - error: An error if romanization fails
func (p *AksharamukhaProvider) romanize(ctx context.Context, text string) (string, error) {
	if p.targetScheme != "" {
		script, err := aksharamukha.DefaultScriptFor(p.Lang)
		if err != nil {
			return "", fmt.Errorf("DefaultScriptFor failed for lang \"%s\": %w", p.Lang, err)
		}
		
		// Use the context-aware version
		romanized, err := aksharamukha.TranslitWithContext(ctx, text, script, p.targetScheme, aksharamukha.DefaultOptions())
		if err != nil {
			return "", fmt.Errorf("romanization failed for token \"%s\" with scheme %s: %w", text, p.targetScheme, err)
		}
		return romanized, err
	}
	// Use the context-aware version for default romanization
	return aksharamukha.RomanWithContext(ctx, text, p.Lang, aksharamukha.DefaultOptions())
}


func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}