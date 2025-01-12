package jpn

import (
	"fmt"

	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	common "github.com/tassa-yoniso-manasi-karoto/translitkit"

	//iso "github.com/barbashov/iso639-3"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

const Lang = "jpn"

// IchiranProvider satisfies the Provider interface
type IchiranProvider struct {
	config map[string]interface{}
	docker *ichiran.Docker
}

func (p *IchiranProvider) Init() (err error) {
	p.docker, err = ichiran.NewDocker()
	if err != nil {
		return fmt.Errorf("failed to create API client for Ichiran Docker: %v", err)
	}

	if err = p.docker.Init(); err != nil {
		return fmt.Errorf("failed to initialize: %v", err)
	}
	return
}

func (p *IchiranProvider) Name() string {
	return "ichiran"
}

func (p *IchiranProvider) GetType() common.ProviderType {
	return common.CombinedType
}

func (p *IchiranProvider) Close() error {
	return p.docker.Close()
}

// FIXME passing m *common.Module no longer useful?
// this method should probably private at first glance

func (p *IchiranProvider) Process(m *common.Module, input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
	raw := input.GetRaw()
	if input.Len() == 0 && raw == "" {
		return nil, fmt.Errorf("empty input was passed to processor")
	}
	ProviderType := p.GetType()
	if raw != "" {
		switch ProviderType {
		case common.TokenizerType:
			//results = p.process(ToTokenSlice, input[i].Surface)
		case common.TransliteratorType:
			// note: the output format will be Tkns so due to the absence of tokenization everything will be stuffed into Tkns[0]
			// results = []TokenContainer{new(common.Tkn)}
			// results[0] = p.process((*ichiran.JSONTokens).Roman, input[i].Surface)
		case common.CombinedType:
			return p.process(ToTokenSlice, raw)
		}
		input = input.ClearRaw()
	} else { // generic token processor: take common.Tkn as well as lang-specic tokens that have common.Tkn as their embedded field
		switch ProviderType {
		case common.TokenizerType:
			// Either refuse or add linguistic annotations
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a tokenizer")
		case common.TransliteratorType:
			/*for i, _ := range input {
				input[i].Romanization = p.process((*ichiran.JSONTokens).Roman, input[i].Surface)
			}
			results = input*/
		case common.CombinedType:
			// Refuse because it is already tokenized
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a provider that combines tokenizer+transliterator")
		}
	}
	return nil, fmt.Errorf("handling not implemented for '%s' with ProviderType '%s'", p.Name(), ProviderType)
}

// process accepts a transformation function: the desired ichiran method to use
func (p *IchiranProvider) process(transform func(*ichiran.JSONTokens) common.AnyTokenSliceWrapper, input string) (results common.AnyTokenSliceWrapper, err error) {
	text, err := ichiran.Analyze(input)
	if err != nil {
		return nil, fmt.Errorf("failed to analyse: %v", err)
	}
	return transform(text), nil
}

func init() {
	ichiran := &IchiranProvider{}

	err := common.Register(Lang, common.CombinedType, "ichiran", common.ProviderEntry{
		Provider:     ichiran,
		Capabilities: []string{"tokenization", "reading", "romaji"},
		Type:         common.CombinedType,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to register ichiran provider: %v", err))
	}
	err = common.SetDefault(Lang, []common.ProviderEntry{
		{
			Provider: ichiran,
			Type:     common.CombinedType,
		}, // TODO add robepike/nihongo to force romanization after
	})
	if err != nil {
		panic(fmt.Sprintf("failed to set ichiran as default: %v", err))
	}
}

// // Example of registering separate providers:
// func init() {
// 	ja := iso.FromAnyCode("ja")

// 	// Create provider instances
// 	mecab := &MecabProvider{}
// 	hepburn := &HepburnProvider{}

// 	// Register MeCab tokenizer
// 	err := common.Register(ja, common.TokenizerType, "mecab", common.ProviderEntry{
// 		Provider:     mecab,
// 		Capabilities: []string{"tokenization"},
// 		Type:        common.TokenizerType,
// 	})
// 	if err != nil {
// 		panic(fmt.Sprintf("failed to register mecab provider: %v", err))
// 	}

// 	// Register Hepburn transliterator
// 	err = common.Register(ja, common.TransliteratorType, "hepburn", common.ProviderEntry{
// 		Provider:     hepburn,
// 		Capabilities: []string{"romaji"},
// 		Type:        common.TransliteratorType,
// 	})
// 	if err != nil {
// 		panic(fmt.Sprintf("failed to register hepburn provider: %v", err))
// 	}

// 	// Set as default providers (both needed for separate mode)
// 	err = common.SetDefault(ja, []common.ProviderEntry{
// 		{
// 			Provider: mecab,
// 			Type:    common.TokenizerType,
// 		},
// 		{
// 			Provider: hepburn,
// 			Type:    common.TransliteratorType,
// 		},
// 	})
// 	if err != nil {
// 		panic(fmt.Sprintf("failed to set mecab+hepburn as default: %v", err))
// 	}
// }

func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
