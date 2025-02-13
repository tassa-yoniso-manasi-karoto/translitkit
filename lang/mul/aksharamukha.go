
package mul

import (
	"fmt"
	"strings"
	"math"
	"context"

	"github.com/tassa-yoniso-manasi-karoto/go-aksharamukha"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

// AksharamukhaProvider satisfies the Provider interface
type AksharamukhaProvider struct {
	ctx		context.Context
	Config		map[string]interface{}
	Lang		string // ISO 639-3 language code
	targetScheme	aksharamukha.Script
}


func (p *AksharamukhaProvider) WithContext(ctx context.Context) {
	aksharamukha.Ctx = ctx
}

// NewAksharamukhaProvider creates a new provider instance with the specified language
func NewAksharamukhaProvider(lang string) *AksharamukhaProvider {
	return &AksharamukhaProvider{
		Config: make(map[string]interface{}),
		Lang:   lang,
	}
}

func (p *AksharamukhaProvider) Init() (err error) {
	if p.Lang == "" {
		return fmt.Errorf("language code must be set before initialization")
	}

	if err = aksharamukha.Init(); err != nil {
		return fmt.Errorf("failed to initialize aksharamukha: %v", err)
	}
	return
}

func (p *AksharamukhaProvider) InitRecreate(noCache bool) (err error) {
	if p.Lang == "" {
		return fmt.Errorf("language code must be set before initialization")
	}

	if err = aksharamukha.InitRecreate(noCache); err != nil {
		return fmt.Errorf("failed to initialize aksharamukha: %v", err)
	}
	return
}

func (p *AksharamukhaProvider) Name() string {
	return "aksharamukha"
}

func (p *AksharamukhaProvider) GetType() common.ProviderType {
	return common.TransliteratorType
}

func (p *AksharamukhaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

func (p *AksharamukhaProvider) Close() error {
	return aksharamukha.Close()
}

func (p *AksharamukhaProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
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


// process handles raw input strings // FIXME THIS WILL TURN INTO TOKENS AND TRANSLITERATE ENTIRE CHUNKS
func (p *AksharamukhaProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &common.TknSliceWrapper{}
	
	for _, chunk := range chunks {
		if len(strings.TrimSpace(chunk)) == 0 {
			continue
		}

		token := common.Tkn{
			Surface: chunk,
			IsLexical: false,
		}

		romanized, err := p.romanize(chunk)
		if err != nil {
			return nil, fmt.Errorf("romanization failed: %w", err)
		}

		token.Romanization = romanized
		if chunk != romanized {
			token.IsLexical = true
		}
		tsw.Append(&token)
	}

	return tsw, nil
}

// processTokens handles pre-tokenized input
func (p *AksharamukhaProvider) processTokens(input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	for _, tkn := range input.(*common.TknSliceWrapper).Slice {
		s := tkn.GetSurface()
		if !tkn.IsLexicalContent() || s == "" || tkn.Roman() != "" {
			continue
		}
		romanized, err := p.romanize(s)
		if err != nil {
			return nil, fmt.Errorf("romanization failed for token %s: %w", s, err)
		}
		tkn.SetRoman(romanized)
		//fmt.Printf("TOKEN %s: %s\n\n\n", color.Bold.Sprint(idx), pp.Sprint(tkn))
	}
	return input, nil
}

func (p *AksharamukhaProvider) romanize(text string) (string, error) {
	if p.targetScheme != "" {
		script, err := aksharamukha.DefaultScriptFor(p.Lang)
		if err != nil {
			return "", fmt.Errorf("DefaultScriptFor failed for lang \"%s\": %w", p.Lang, err)
		}
		romanized, err := aksharamukha.Translit(text, script, p.targetScheme)
		if err != nil {
			return "", fmt.Errorf("romanization failed for token \"%s\" with scheme %s: %w", text, p.targetScheme, err)
		}
		return romanized, err
	}
	// otherwise use default romanization
	return aksharamukha.Roman(text, p.Lang)
}

func (p *AksharamukhaProvider) SetConfig(config map[string]interface{}) error {
	schemeName, ok := config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}

	lang, ok := config["lang"].(string)
	if !ok {
		return fmt.Errorf("lang not provided in config")
	}
	p.Lang = lang
	
	// Convert scheme name to target aksharamukha.Script
	targetScheme, ok := indicSchemesToScript[schemeName]
	if !ok {
		return fmt.Errorf("unsupported transliteration scheme: %s", schemeName)
	}

	p.targetScheme = targetScheme
	return nil
}


func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}