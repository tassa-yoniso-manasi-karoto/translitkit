
package mul

import (
	"fmt"
	"math"

	iuliia "github.com/mehanizm/iuliia-go"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// IuliiaProvider satisfies the Provider interface
type IuliiaProvider struct {
	Config map[string]interface{}
	Lang   string // ISO 639-3 language code
	Schema *iuliia.Schema
}

// NewIuliiaProvider creates a new provider instance
func NewIuliiaProvider(lang string) *IuliiaProvider {
	return &IuliiaProvider{
		Config: make(map[string]interface{}),
		Lang:   lang,
		Schema: iuliia.Gost_779, // Default Schema // TODO Make configurable
	}
}

func (p *IuliiaProvider) Init() error {
	switch p.Lang {
	case "rus":
	case "uzb":
		p.Schema = iuliia.Uz
	case "":
		return fmt.Errorf("language code must be set before initialization")
	default:
		return fmt.Errorf("\"%s\" is not a language code supported by Iuliia", p.Lang)
	}
	return nil
}

func (p *IuliiaProvider) Name() string {
	return "iuliia"
}

func (p *IuliiaProvider) GetType() common.ProviderType {
	return common.TransliteratorType
}

func (p *IuliiaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

func (p *IuliiaProvider) Close() error {
	return nil
}

func (p *IuliiaProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}

	providerType := p.GetType()
	if len(raw) != 0 {
		switch providerType {
		case common.TransliteratorType:
			return p.process(raw)
		default:
			return nil, fmt.Errorf("provider type %s not supported", providerType)
		}
		input.ClearRaw()
	} else {
		switch providerType {
		case common.TransliteratorType:
			return p.processTokens(input)
		default:
			return nil, fmt.Errorf("provider type %s not supported", providerType)
		}
	}
	return nil, fmt.Errorf("handling not implemented for '%s' with ProviderType '%s'", p.Name(), providerType)
}

// process handles raw input strings // TODO see aksharamukha remark on processing raw with no tokenization
func (p *IuliiaProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}
	
	for _, chunk := range chunks {
		token := common.Tkn{
			Surface: chunk,
			IsToken: true,
		}

		romanized := p.Schema.Translate(chunk)
		token.Romanization = romanized
		tsw.Append(&token)
	}

	return tsw, nil
}

// processTokens handles pre-tokenized input
func (p *IuliiaProvider) processTokens(input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	for _, tkn := range input.(*common.TknSliceWrapper).Slice {
		s := tkn.GetSurface()
		if !tkn.IsTokenType() || s == "" || tkn.Roman() != "" {
			continue
		}
		romanized := p.Schema.Translate(s)
		tkn.SetRoman(romanized)
	}

	return input, nil
}