
package mul

import (
	"fmt"
	"strings"
	"context"
	"unicode"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"

	"github.com/rivo/uniseg"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

type UnisegProvider struct {
	config       map[string]interface{}
	progressCallback common.ProgressCallback
	lang         string
	scriptRanges []*unicode.RangeTable
}


// WithProgressCallback sets a callback function for reporting progress during processing.
func (p *UnisegProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// WithDownloadProgressCallback sets a callback for download progress (no-op for Uniseg).
func (p *UnisegProvider) WithDownloadProgressCallback(callback common.DownloadProgressCallback) {
	// No-op: Uniseg doesn't require Docker downloads
}

// SaveConfig stores the configuration for later application during initialization.
// It extracts the language code and retrieves the expected Unicode script ranges for that language.
//
// Returns an error if the configuration is invalid.
func (p *UnisegProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg

	if langVal, ok := cfg["lang"].(string); ok && langVal != "" {
		p.lang = langVal
		p.scriptRanges, _ = common.GetUnicodeRangesFromLang(p.lang)
	} else {
		p.lang = "" // Default to no language-specific behavior
	}
	return nil
}

// InitWithContext initializes the provider with the given context.
// For Uniseg, this is a no-op as there are no resources to initialize.
// The context can be used for cancellation, though initialization is immediate.
//
// Returns nil as there are no initialization steps that can fail.
func (p *UnisegProvider) InitWithContext(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("uniseg: context canceled during initialization: %w", err)
	}
	return nil
}

// Init initializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns nil as there are no initialization steps that can fail.
func (p *UnisegProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider from scratch with the given context.
// For Uniseg, this is equivalent to InitWithContext as there are no persistent resources.
// The context can be used for cancellation, though reinitialization is immediate.
//
// Returns nil as there are no reinitialization steps that can fail.
func (p *UnisegProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	return p.InitWithContext(ctx)
}

// InitRecreate reinitializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns nil as there are no reinitialization steps that can fail.
func (p *UnisegProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

func (p *UnisegProvider) Name() string {
	return "uniseg"
}

func (p *UnisegProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TokenizerMode}
}

func (p *UnisegProvider) GetMaxQueryLen() int {
	return 0
}

// CloseWithContext releases resources used by the provider with the given context.
// For Uniseg, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *UnisegProvider) CloseWithContext(ctx context.Context) error {
	return nil
}

// Close releases resources used by the provider with a background context.
// For Uniseg, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *UnisegProvider) Close() error {
	return nil
}

// ProcessFlowController processes input tokens using the specified context.
// This handles raw input chunks only, as Uniseg is a tokenizer that doesn't work with pre-tokenized content.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The token slice wrapper containing raw input chunks
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails, the context is canceled, or input format is invalid
func (p *UnisegProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("uniseg: context canceled during processing: %w", err)
	}
	
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}

	if len(raw) != 0 {
		return p.process(ctx, raw)
	}

	// We don't handle already tokenized input
	return nil, fmt.Errorf("tokens not accepted as input for uniseg tokenizer")
}

// process implements the actual tokenization logic using uniseg.
// It segments text into words according to Unicode rules and marks tokens as lexical or non-lexical.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - chunks: Raw text chunks to process
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails or the context is canceled
func (p *UnisegProvider) process(ctx context.Context, chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}
	totalChunks := len(chunks)

	for idx, chunk := range chunks {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("uniseg: context canceled while processing chunk %d: %w", idx, err)
		}
		
		// Report progress if callback is set
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		trimmed := strings.TrimSpace(chunk)
		if len(trimmed) == 0 {
			continue
		}

		// State for uniseg word segmentation
		remaining := trimmed
		state := -1

		for len(remaining) > 0 {
			// Check for context cancellation in long loops
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("uniseg: context canceled during word segmentation: %w", err)
			}
			
			word, rest, newState := uniseg.FirstWordInString(remaining, state)
			if word != "" {
				token := common.Tkn{
					Surface: word,
					Position: struct {
						Start     int
						End       int
						Sentence  int
						Paragraph int
					}{
						Start: len(trimmed) - len(remaining),
						End:   len(trimmed) - len(rest),
					},
					// We decide lexical vs. non-lexical inside isLexical() helper
					IsLexical: p.isLexical(word),
				}

				tsw.Append(&token)
			}
			remaining = rest
			state = newState
		}
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(totalChunks, totalChunks)
	}
	
	return tsw, nil
}

// isLexical determines if a token should be considered linguistic content.
// It iterates over all runes in the word and returns true if at least one letter
// belongs to one of the expected script ranges. Otherwise, it returns false.
// If no language/script configuration is available, it falls back to a simple check.
func (p *UnisegProvider) isLexical(word string) bool {
	if word == "" {
		return false
	}

	// If a language and its script ranges are defined, use them.
	if p.lang != "" && len(p.scriptRanges) > 0 {
		for _, r := range word {
			// Check if the rune is a letter and is in one of the expected Unicode ranges.
			if unicode.IsLetter(r) && unicode.IsOneOf(p.scriptRanges, r) {
				return true
			}
		}
		// No letter matched the expected script ranges.
		return false
	}

	// Fallback: If no language/script configuration is available, consider the token lexical
	// if it contains any letter that isn't solely punctuation or a space.
	for _, r := range word {
		if unicode.IsLetter(r) && !isPunctuationOrSpace(r) {
			return true
		}
	}
	return false
}

// isPunctuationOrSpace returns true if the rune is punctuation, symbol, or whitespace.
func isPunctuationOrSpace(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r)
}

func placehold3445654er() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
