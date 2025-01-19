package jpn

import (
	"fmt"
	"math"
	
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"

	//iso "github.com/barbashov/iso639-3"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

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

// Returns a limit based on the max of a 32 bit integer.
// Ichiran's developper doesn't know of a length limit to the input of the CLI
// but I am setting this just in case. It could also be MaxInt64.
func (p *IchiranProvider) GetMaxQueryLen() int {
	return math.MaxInt32-2
}

func (p *IchiranProvider) Close() error {
	return p.docker.Close()
}



func (p *IchiranProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input was passed to processor")
	}
	ProviderType := p.GetType()
	if len(raw) != 0 {
		switch ProviderType {
		case common.TokenizerType:
		case common.TransliteratorType:
		case common.CombinedType:
			return p.process(raw)
		}
		// Important to clear the field Raw, otherwise Tkn would be ignored by next processor
		input.ClearRaw()
	} else { // generic token processor: take common.Tkn as well as lang-specic tokens that have common.Tkn as their embedded field
		switch ProviderType {
		case common.TokenizerType:
			// Either refuse or add linguistic annotations
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a tokenizer")
		case common.TransliteratorType:
		case common.CombinedType:
			// Refuse because it is already tokenized
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a provider that combines tokenizer+transliterator")
		}
	}
	return nil, fmt.Errorf("handling not implemented for '%s' with ProviderType '%s'", p.Name(), ProviderType)
}

// returns jpn.TknSliceWrapper that satisfies AnyTokenSliceWrapper
func (p *IchiranProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &TknSliceWrapper{}
	for idx, chunk := range chunks {
		JSONTokens, err := ichiran.Analyze(chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to analyse chunk %d of %d: %v\nraw_chunk=>>>%s<<<", idx, len(chunks), err, chunk)
		}
		tsw.Append(ToAnyTokenSlice(JSONTokens)...)
	}
	return tsw, nil
}

func init() {
	IchiranEntry := common.ProviderEntry{
		Provider:     &IchiranProvider{},
		Capabilities: []string{"tokenization", "transliteration", "romaji"},
		Type:         common.CombinedType,
	}
	err := common.Register(Lang, common.CombinedType, "ichiran", IchiranEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register ichiran provider: %v", err))
	}
	err = common.SetDefault(Lang, []common.ProviderEntry{IchiranEntry}) // TODO add robepike/nihongo to force romanization after
	
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
