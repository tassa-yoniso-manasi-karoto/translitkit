
package mul

import (
	"fmt"
	"strings"
	"unicode"
	"context"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"

	"github.com/rivo/uniseg"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

type UnisegProvider struct {
	ctx	context.Context
	config	map[string]interface{}
}


func (p *UnisegProvider) WithContext(ctx context.Context) {
	p.ctx = ctx
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

func (p *UnisegProvider) SetConfig(map[string]interface{}) error {
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

// process implements the actual tokenization logic
func (p *UnisegProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}

	for _, chunk := range chunks {
		if len(strings.TrimSpace(chunk)) == 0 {
			continue
		}

		// Initialize state for word segmentation
		remaining := chunk
		state := -1

		for len(remaining) > 0 {
			// Get next word
			word, rest, newState := uniseg.FirstWordInString(remaining, state)
			
			if word != "" {
				token := common.Tkn{
					Surface:  word,
					Position: struct {
						Start     int
						End       int
						Sentence  int
						Paragraph int
					}{
						Start: len(chunk) - len(remaining),
						End:   len(chunk) - len(rest),
					},
				}

				tsw.Append(&token)
			}

			remaining = rest
			state = newState
		}
	}
	return tsw, nil
}

// isSpaceOrPunct returns true if the string consists only of spaces or punctuation
func isSpaceOrPunct(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) && !unicode.IsPunct(r) {
			return false
		}
	}
	return true
}



func placehold3445654er() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}