package zho

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// toneNumberRegex extracts the tone number from numeric pinyin notation like "hao3"
var toneNumberRegex = regexp.MustCompile(`(\d)$`)

// GoPinyinProvider implements the Provider interface for Chinese Pinyin transliteration.
// It uses the go-pinyin library to convert Chinese characters to Pinyin romanization.
// This provider chooses the "most frequent" reading for Tkn.Pinyin while also storing
// all alternative readings in Tkn.PinyinAll and Tkn.PinyinNumAll.
type GoPinyinProvider struct {
	config           map[string]interface{}
	progressCallback common.ProgressCallback
	initialized      bool

	chosenScheme string
	mainStyle    int
	numStyle     int

	mainArgs pinyin.Args
	numArgs  pinyin.Args
}

// WithProgressCallback sets a callback function for reporting progress during processing.
// This is a no-op for GoPinyin as it typically processes text very quickly.
func (p *GoPinyinProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// WithDownloadProgressCallback sets a callback for download progress (no-op for GoPinyin).
func (p *GoPinyinProvider) WithDownloadProgressCallback(callback common.DownloadProgressCallback) {
	// No-op: GoPinyin doesn't require Docker downloads
}

// SaveConfig stores the configuration for later application during initialization.
// This allows the provider to be configured before being initialized.
//
// Returns an error if the configuration is invalid.
func (p *GoPinyinProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}

// InitWithContext initializes the provider with the given context.
// This sets up the pinyin styles and configurations based on the stored configuration.
// The context can be used for cancellation, though initialization is typically quick.
//
// Returns an error if initialization fails or the context is canceled.
func (p *GoPinyinProvider) InitWithContext(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gopinyin: context canceled during initialization: %w", err)
	}

	if p.initialized {
		return nil
	}

	schemeName, _ := p.config["scheme"].(string)
	if schemeName == "" {
		schemeName = "tone" // default diacritic
	}
	p.chosenScheme = schemeName

	style, ok := PinyinSchemes[strings.ToLower(schemeName)]
	if !ok {
		style = pinyin.Tone
	}
	p.mainStyle = style
	p.numStyle = pinyin.Tone2

	// Prepare mainArgs
	p.mainArgs = pinyin.NewArgs()
	p.mainArgs.Style = p.mainStyle
	p.mainArgs.Heteronym = true // gather multiple possible pronunciations

	// Prepare numArgs
	p.numArgs = pinyin.NewArgs()
	p.numArgs.Style = p.numStyle
	p.numArgs.Heteronym = true // also gather multiple numeric variants

	p.initialized = true
	return nil
}

// Init initializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if initialization fails.
func (p *GoPinyinProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider from scratch with the given context.
// This clears all configuration and resets the provider to its initial state.
// The context can be used for cancellation, though reinitialization is typically quick.
//
// Returns an error if reinitialization fails or the context is canceled.
func (p *GoPinyinProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gopinyin: context canceled during reinitialization: %w", err)
	}

	p.initialized = false
	p.mainArgs = pinyin.Args{}
	p.numArgs = pinyin.Args{}
	p.mainStyle = 0
	p.numStyle = 0
	return p.InitWithContext(ctx)
}

// InitRecreate reinitializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if reinitialization fails.
func (p *GoPinyinProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

// ProcessFlowController processes input tokens using the specified context.
// This processes pre-tokenized input, adding Pinyin romanization to Chinese tokens.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The token slice wrapper to process
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails, the context is canceled, or initialization fails
func (p *GoPinyinProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("gopinyin: context canceled during processing: %w", err)
	}

	if err := p.InitWithContext(ctx); err != nil {
		return nil, fmt.Errorf("gopinyin init failed: %w", err)
	}

	tokens := input.Len()
	for i := 0; i < tokens; i++ {
		// Periodically check for context cancellation
		if i%100 == 0 && ctx.Err() != nil {
			return nil, fmt.Errorf("gopinyin: context canceled while processing token %d: %w", i, ctx.Err())
		}
		
		// Report progress if callback is set (throttler handles batching)
		if p.progressCallback != nil {
			p.progressCallback(i, tokens)
		}
		
		anyTkn := input.GetIdx(i)
		if !anyTkn.IsLexicalContent() {
			continue
		}

		zhoTkn, ok := anyTkn.(*Tkn)
		if !ok {
			// Not our specialized token => fallback
			anyTkn.SetRoman(anyTkn.GetSurface())
			continue
		}

		if !zhoTkn.IsChinese() {
			zhoTkn.SetRoman(zhoTkn.Surface)
			continue
		}

		// 1) Retrieve diacritic multi-pronunciation data
		allSyllables := pinyin.Pinyin(zhoTkn.Surface, p.mainArgs) // 2D slice
		zhoTkn.PinyinAll = allSyllables

		// 2) Retrieve numeric multi-pronunciation data
		allNumSyllables := pinyin.Pinyin(zhoTkn.Surface, p.numArgs)
		zhoTkn.PinyinNumAll = allNumSyllables

		// 3) The "most frequent" reading is the *first* in each sub-slice.
		// We'll build Tkn.Pinyin from that.
		var chosenDiacritic []string
		var chosenNumeric []string

		for idxChar, arr := range allSyllables {
			if len(arr) > 0 {
				chosenDiacritic = append(chosenDiacritic, arr[0])
			} else {
				// fallback if no reading
				chosenDiacritic = append(chosenDiacritic, "")
			}

			numArr := allNumSyllables[idxChar]
			if len(numArr) > 0 {
				chosenNumeric = append(chosenNumeric, numArr[0])
			} else {
				chosenNumeric = append(chosenNumeric, "")
			}
		}

		zhoTkn.Pinyin = strings.Join(chosenDiacritic, " ")
		zhoTkn.PinyinNum = strings.Join(chosenNumeric, " ")

		// 4) If single-syllable, parse numeric tone
		if len(chosenNumeric) == 1 {
			toneVal := parseToneNumber(chosenNumeric[0])
			if toneVal > 0 {
				zhoTkn.Tone = Tone(toneVal)
				zhoTkn.OriginalTone = zhoTkn.Tone
				zhoTkn.HasToneSandhi = false
			}
		}

		// 5) Put the final reading in Tkn.Romanization
		zhoTkn.SetRoman(zhoTkn.Pinyin)
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(tokens, tokens)
	}

	return input, nil
}


// Name identifies this provider as "gopinyin".
func (p *GoPinyinProvider) Name() string {
	return "gopinyin"
}

func (p *GoPinyinProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TransliteratorMode}
}

func (p *GoPinyinProvider) GetMaxQueryLen() int {
	return 0
}

// CloseWithContext releases resources used by the provider with the given context.
// For GoPinyin, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *GoPinyinProvider) CloseWithContext(ctx context.Context) error {
	return nil
}

// Close releases resources used by the provider with a background context.
// For GoPinyin, this is a no-op as there are no persistent resources to release.
//
// Returns nil as there are no resources to release.
func (p *GoPinyinProvider) Close() error {
	return nil
}


// PinyinSchemes maps user-friendly scheme names to pinyin int constants.
var PinyinSchemes = map[string]int{
	"normal":       pinyin.Normal,
	"tone":         pinyin.Tone,
	"tone2":        pinyin.Tone2,
	"tone3":        pinyin.Tone3,
	"initials":     pinyin.Initials,
	"firstletter":  pinyin.FirstLetter,
	"finals":       pinyin.Finals,
	"finalstone":   pinyin.FinalsTone,
	"finalstone2":  pinyin.FinalsTone2,
	"finalstone3":  pinyin.FinalsTone3,
}

// parseToneNumber picks the last digit [1..5] from a tone2 syllable like "hao3".
// This is a helper function for extracting tone numbers from numeric Pinyin notation.
//
// Parameters:
//   - s: The syllable with numeric tone marking
//
// Returns:
//   - int: The tone number (1-5), or 0 if no valid tone number is found
func parseToneNumber(s string) int {
	match := toneNumberRegex.FindStringSubmatch(s)
	if len(match) < 2 {
		return 0
	}
	num, _ := strconv.Atoi(match[1])
	return num
}
