package zho

import (
	"context"
	"fmt"
	"math"

	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	"github.com/yanyiwu/gojieba"
)

// GoJiebaProvider tokenizes Chinese text with gojieba, preserving filler tokens
// and populating relevant fields in our custom Tkn struct where possible.
type GoJiebaProvider struct {
	ctx    context.Context
	config map[string]interface{}
	jieba  *gojieba.Jieba
}

// WithContext updates the provider's context.
func (p *GoJiebaProvider) WithContext(ctx context.Context) {
	if ctx != nil {
		p.ctx = ctx
	}
}

func (p *GoJiebaProvider) WithProgressCallback(callback common.ProgressCallback) {
}


// SaveConfig stores configuration (e.g., custom dict paths) for use by Init().
func (p *GoJiebaProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}

// Init initializes the gojieba engine (if not already done).
func (p *GoJiebaProvider) Init() error {
	if p.jieba != nil {
		return nil
	}
	// If your config includes dictPaths, parse them here, e.g.:
	//   dictPaths, _ := p.config["dictPaths"].([]string)
	//   p.jieba = gojieba.NewJieba(dictPaths...)
	p.jieba = gojieba.NewJieba()
	return nil
}

// InitRecreate frees existing gojieba resources and re-initializes from scratch.
func (p *GoJiebaProvider) InitRecreate(noCache bool) error {
	if p.jieba != nil {
		p.jieba.Free()
		p.jieba = nil
	}
	return p.Init()
}

// ProcessFlowController performs tokenization on raw chunks, preserving filler tokens.
// It populates Tkn fields that can be informed by gojieba (e.g. part-of-speech).
func (p *GoJiebaProvider) ProcessFlowController(
	input common.AnyTokenSliceWrapper,
) (common.AnyTokenSliceWrapper, error) {

	// Ensure gojieba is initialized
	if p.jieba == nil {
		if err := p.Init(); err != nil {
			return nil, fmt.Errorf("failed to init gojieba: %w", err)
		}
	}

	rawChunks := input.GetRaw()
	if len(rawChunks) == 0 {
		return input, nil
	}

	outWrapper := &TknSliceWrapper{}

	for _, chunk := range rawChunks {
		if chunk == "" {
			continue
		}

		// 1) Use gojieba for lexical segmentation + POS tags
		words := p.jieba.Cut(chunk, true) // "precise" mode with HMM
		tags := p.jieba.Tag(chunk)
		if len(words) != len(tags) {
			return nil, fmt.Errorf("gojieba mismatch: len(words)=%d, len(tags)=%d", len(words), len(tags))
		}

		// 2) Integrate lexical tokens with filler
		integrated := common.IntegrateProviderTokens(chunk, words)

		// We'll attach each recognized lexical token's POS from 'tags' in order
		lexCount := 0
		for _, fillerOrLex := range integrated {
			// Build a new zho.Tkn from the integrated token
			zhoTkn := &Tkn{
				Tkn: *fillerOrLex,

				// For Chinese tokens, we can at least guess that 'Surface' is both
				// the simplified and traditional form if we have no external DB:
				Simplified:  fillerOrLex.Surface,
				Traditional: fillerOrLex.Surface,

				// We won't fill `NumStrokes`, `Radical`, etc. because gojieba
				// doesn't supply stroke or radical data.
				// We'll also leave morphological + idiomatic fields at defaults.
			}

			if fillerOrLex.IsLexical {
				// The next POS tag in 'tags' corresponds to this lexical word
				pos := tags[lexCount]
				lexCount++

				// Store generic POS in Tkn.PartOfSpeech
				zhoTkn.PartOfSpeech = pos

				// If the user wants, store that same POS in some custom field, or
				// interpret it further:
				// For instance, if POS == "q", we might guess it's a classifier.
				if pos == "q" {
					// Mark it as a classifier
					zhoTkn.ClassifierType = "indiv" // naive assumption
				}

				// If you want to guess if it's an idiom or something from the POS,
				// you could do so here, though gojieba's default dict rarely marks that.

				// If we see 'a' (形容词), we might guess it's a stative verb in Chinese:
				if pos == "a" {
					zhoTkn.IsStative = true
				}
			}

			// Append the new token
			outWrapper.Append(zhoTkn)
		}
	}

	// Clear raw chunks to mark they've been processed
	input.ClearRaw()

	return outWrapper, nil
}

// Name returns the unique name of this provider.
func (p *GoJiebaProvider) Name() string {
	return "gojieba"
}

// GetType returns the ProviderType: Tokenizer.
func (p *GoJiebaProvider) GetType() common.ProviderType {
	return common.TokenizerType
}

// GetMaxQueryLen returns a large number so the module can handle big input.
func (p *GoJiebaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

// Close frees the gojieba instance.
func (p *GoJiebaProvider) Close() error {
	if p.jieba != nil {
		p.jieba.Free()
		p.jieba = nil
	}
	return nil
}
