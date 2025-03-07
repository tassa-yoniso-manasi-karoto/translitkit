
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
	ctx          context.Context
	config       map[string]interface{}
	lang         string
	scriptRanges []*unicode.RangeTable
}


func (p *UnisegProvider) WithContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *UnisegProvider) WithProgressCallback(callback common.ProgressCallback) {
}

// SaveConfig stores the config and extracts the language code.
// It then retrieves the expected Unicode script ranges for that language.
func (p *UnisegProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg

	if langVal, ok := cfg["lang"].(string); ok && langVal != "" {
		p.lang = langVal
		p.scriptRanges, _ = common.GetUnicodeRangesFromLang(p.lang)
	} else {
		p.lang = "" // TODO FIXME
	}
	return nil
}

func (p *UnisegProvider) Init() error {
	return nil
}

func (p *UnisegProvider) InitRecreate(bool) error {
	return nil
}

func (p *UnisegProvider) Name() string {
	return "uniseg"
}

func (p *UnisegProvider) GetType() common.ProviderType {
	return common.TokenizerType
}

func (p *UnisegProvider) GetMaxQueryLen() int {
	return 0
}

func (p *UnisegProvider) Close() error {
	return nil
}

// ProcessFlowController implements the common.Provider interface
func (p *UnisegProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}

	if len(raw) != 0 {
		return p.process(raw)
	}

	// We don't handle already tokenized input
	return nil, fmt.Errorf("tokens not accepted as input for uniseg tokenizer")
}

// process implements the actual tokenization logic using uniseg.
// We additionally mark tokens as lexical or non-lexical.
func (p *UnisegProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}

	for _, chunk := range chunks {
		trimmed := strings.TrimSpace(chunk)
		if len(trimmed) == 0 {
			continue
		}

		// State for uniseg word segmentation
		remaining := trimmed
		state := -1

		for len(remaining) > 0 {
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
