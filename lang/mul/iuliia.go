
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
	targetScheme *iuliia.Schema
}

// NewIuliiaProvider creates a new provider instance
func NewIuliiaProvider(lang string) *IuliiaProvider {
	return &IuliiaProvider{
		Config: make(map[string]interface{}),
		Lang:   lang,
	}
}

func (p *IuliiaProvider) Init() error {
	switch p.Lang {
	case "rus", "uzb":
	case "":
		return fmt.Errorf("language code must be set before initialization")
	default:
		return fmt.Errorf("\"%s\" is not a language code supported by Iuliia", p.Lang)
	}
	return nil
}

func (p *IuliiaProvider) InitRecreate(bool) error {
	return p.Init()
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

func (p *IuliiaProvider) SetConfig(config map[string]interface{}) error {
	schemeName, ok := config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}

	lang, ok := config["lang"].(string)
	if !ok {
		return fmt.Errorf("lang not provided in config")
	}
	p.Lang = lang
	
	targetScheme, ok := russianSchemesToScript[schemeName]
	if !ok {
		return fmt.Errorf("unsupported transliteration scheme: %s", schemeName)
	}

	p.targetScheme = targetScheme
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
			IsLexical: false,
		}

		romanized := iuliia.Gost_779.Translate(chunk)
		token.Romanization = romanized
		if chunk != romanized {
			token.IsLexical = true
		}
		tsw.Append(&token)
	}

	return tsw, nil
}

// processTokens handles pre-tokenized input
func (p *IuliiaProvider) processTokens(input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	for _, tkn := range input.(*common.TknSliceWrapper).Slice {
		s := tkn.GetSurface()
		if !tkn.IsLexicalContent() || s == "" || tkn.Roman() != "" {
			continue
		}
		tkn.SetRoman(p.romanize(s))
	}

	return input, nil
}


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