
package mul

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

type UnisegProvider struct {
	config map[string]interface{}
}

func (p *UnisegProvider) Init() error {
	return nil // No initialization needed
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
				trimmed := strings.TrimSpace(word)
				isToken := len(trimmed) > 0 && !isSpaceOrPunct(trimmed)
				
				// Create token for the word
				token := common.Tkn{
					Surface:    strings.TrimSpace(word),
					IsToken:    isToken,
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

				// For non-Latin scripts, the Surface becomes the romanization target
				// if !isLatinScript(word) {
				// 	token.Romanization = "" // Will be filled by transliterator
				// }

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

// isLatinScript is a basic check for Latin script
// This should be enhanced with proper script detection
// func isLatinScript(s string) bool {
// 	for _, r := range s {
// 		if r > 0x7F { // Basic Latin ends at 0x7F
// 			return false
// 		}
// 	}
// 	return true
// }
