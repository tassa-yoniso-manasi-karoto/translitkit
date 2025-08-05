
package mul

import (
	"fmt"
	"math"
	"context"

	iuliia "github.com/mehanizm/iuliia-go"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// IuliiaProvider satisfies the Provider interface
type IuliiaProvider struct {
	config		map[string]interface{}
	Lang 		string // ISO 639-3 language code
	targetScheme	*iuliia.Schema
	progressCallback common.ProgressCallback
}

// NewIuliiaProvider creates a new provider instance
func NewIuliiaProvider(lang string) *IuliiaProvider {
	return &IuliiaProvider{
		Lang:   lang,
	}
}


// WithProgressCallback sets a callback function for reporting progress during processing.
func (p *IuliiaProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// SaveConfig stores the configuration for later application during initialization.
// This allows the provider to be configured before being initialized.
//
// Returns an error if the configuration is invalid.
func (p *IuliiaProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	lang, ok := p.config["lang"].(string)
	if !ok {
		return fmt.Errorf("lang not provided in config")
	}
	p.Lang = lang
	return nil
}

// InitWithContext initializes the provider with the given context.
// For Iuliia, this validates the language setting and applies any stored configuration.
// The context can be used for cancellation, though initialization is typically quick.
//
// Returns an error if initialization fails, language is not supported, or the context is canceled.
func (p *IuliiaProvider) InitWithContext(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("iuliia: context canceled during initialization: %w", err)
	}
	
	switch p.Lang {
	case "rus", "uzb":
	case "":
		return fmt.Errorf("language code must be set before initialization")
	default:
		return fmt.Errorf("\"%s\" is not a language code supported by Iuliia", p.Lang)
	}
	return p.applyConfig()
}

// Init initializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if initialization fails or language is not supported.
func (p *IuliiaProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider from scratch with the given context.
// For Iuliia, this is equivalent to InitWithContext as there are no persistent resources.
// The context can be used for cancellation, though reinitialization is typically quick.
//
// Returns an error if reinitialization fails, language is not supported, or the context is canceled.
func (p *IuliiaProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	return p.InitWithContext(ctx)
}

// InitRecreate reinitializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if reinitialization fails or language is not supported.
func (p *IuliiaProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

func (p *IuliiaProvider) applyConfig() error {
	if p.config == nil {
		return nil
	}
	schemeName, ok := p.config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}
	
	targetScheme, ok := russianSchemesToScript[schemeName]
	if !ok {
		return fmt.Errorf("unsupported transliteration scheme: %s", schemeName)
	}

	p.targetScheme = targetScheme
	return nil
}

func (p *IuliiaProvider) Name() string {
	return "iuliia"
}

func (p *IuliiaProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TransliteratorMode}
}

func (p *IuliiaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

// CloseWithContext releases resources used by the provider with the given context.
// For Iuliia, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *IuliiaProvider) CloseWithContext(ctx context.Context) error {
	return nil
}

// Close releases resources used by the provider with a background context.
// For Iuliia, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *IuliiaProvider) Close() error {
	return nil
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
func (p *IuliiaProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("iuliia: context canceled during processing: %w", err)
	}
	
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}

	if len(raw) != 0 {
		// switch mode {
		// case common.TransliteratorMode:
		// 	return p.process(ctx, raw)
		// default:
		return nil, fmt.Errorf("operating mode %s not supported", mode)
		// }
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

// process handles raw input strings, converting them to tokens with romanization. // FIXME see aksharamukha remark on processing raw with no tokenization
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - chunks: Raw text chunks to process
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails or the context is canceled
/*func (p *IuliiaProvider) process(ctx context.Context, chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}
	totalChunks := len(chunks)
	
	for idx, chunk := range chunks {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("iuliia: context canceled while processing chunk %d: %w", idx, err)
		}
		
		// Report progress if callback is set
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		token := common.Tkn{
			Surface: chunk,
			IsLexical: false,
		}

		romanized := p.romanize(chunk)
		token.Romanization = romanized
		if chunk != romanized {
			token.IsLexical = true
		}
		tsw.Append(&token)
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(totalChunks, totalChunks)
	}

	return tsw, nil
}*/

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
func (p *IuliiaProvider) processTokens(ctx context.Context, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	tokens := input.(*common.TknSliceWrapper).Slice
	totalTokens := len(tokens)
	
	for idx, tkn := range tokens {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("iuliia: context canceled while processing token %d: %w", idx, err)
		}
		
		// Report progress if callback is set
		if p.progressCallback != nil && idx%20 == 0 { // Report every 20 tokens to avoid excessive callbacks
			p.progressCallback(idx, totalTokens)
		}
		
		s := tkn.GetSurface()
		if !tkn.IsLexicalContent() || s == "" || tkn.Roman() != "" {
			continue
		}
		tkn.SetRoman(p.romanize(s))
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(totalTokens, totalTokens)
	}

	return input, nil
}





// romanize converts text to a romanized form using the appropriate scheme.
// It uses either the configured scheme or falls back to a default scheme based on the language.
//
// Parameters:
//   - text: The text to romanize
//
// Returns:
//   - string: The romanized text
func (p *IuliiaProvider) romanize(text string) string {
	if p.targetScheme != nil {
		return p.targetScheme.Translate(text)
	}
	// otherwise use default romanization
	if p.Lang == "uzb" {
		return iuliia.Uz.Translate(text)
	}
	return iuliia.Gost_779.Translate(text)
}

