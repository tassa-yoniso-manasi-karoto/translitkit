package jpn

import (
	"fmt"
	"math"
	"context"
	
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"

	//iso "github.com/barbashov/iso639-3"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

// IchiranProvider satisfies the Provider interface
type IchiranProvider struct {
	config	map[string]interface{}
}

func (p *IchiranProvider) WithContext(ctx context.Context) {
	ichiran.Ctx = ctx
}

// SaveConfig merely stores the config to apply after init
func (p *IchiranProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}


func (p *IchiranProvider) Init() (err error) {
	if err = ichiran.Init(); err != nil {
		return fmt.Errorf("failed to initialize ichiran: %w", err)
	}
	p.applyConfig()
	return
}

func (p *IchiranProvider) InitRecreate(noCache bool) (err error) {
	if err = ichiran.InitRecreate(noCache); err != nil {
		return fmt.Errorf("failed to initialize ichiran: %w", err)
	}
	p.applyConfig()
	return
}

func (p *IchiranProvider) applyConfig() error {
	return nil
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
	return ichiran.Close()
}



// ProcessFlowController either processes raw input chunks or returns an error if tokens are passed in.
func (p *IchiranProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("ichiran: empty input was passed to processor")
	}

	switch p.GetType() {
	case common.CombinedType:
		if len(raw) != 0 {
			// We'll analyze the raw text
			outWrapper, err := p.processChunks(raw)
			if err != nil {
				return nil, err
			}
			input.ClearRaw() // mark that we've consumed the raw data
			return outWrapper, nil
		}
		// If we *already* have tokens in input, we have nowhere to pass them
		return nil, fmt.Errorf("ichiran: not implemented for pre-tokenized data (we are combined)")

	default:
		return nil, fmt.Errorf("ichiran: unsupported provider type %s", p.GetType())
	}
}

// processChunks takes the raw input chunks, runs morphological analysis, and integrates filler tokens.
func (p *IchiranProvider) processChunks(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &TknSliceWrapper{}

	for idx, chunk := range chunks {
		// 1) Ichiran morphological analysis
		jTokens, err := ichiran.Analyze(chunk)
		if err != nil {
			return nil, fmt.Errorf("ichiran: failed to analyze chunk %d: %w\nraw_chunk=>>>%s<<<", idx, err, chunk)
		}

		// Build a string slice of lexical surfaces from jTokens
		// so that we can call IntegrateProviderTokens to preserve filler
		lexSurfaces := make([]string, len(*jTokens))
		for i, jt := range *jTokens {
			lexSurfaces[i] = jt.Surface
		}

		// 2) Combine lexical tokens w/ filler
		integrated := common.IntegrateProviderTokens(chunk, lexSurfaces)

		// We'll iterate integrated tokens, filling morphological data for lexical ones
		lexCount := 0
		for _, tkn := range integrated {
			if tkn.IsLexical {
				// 3) This token corresponds to jTokens[lexCount]
				ichToken := (*jTokens)[lexCount]
				lexCount++

				// Convert to jpn.Tkn (with morphological data)
				jpnTkn := ToJapaneseToken(ichToken)
				// We also preserve the tkn positions if needed:
				jpnTkn.Position.Start = tkn.Position.Start
				jpnTkn.Position.End = tkn.Position.End

				tsw.Append(jpnTkn)
			} else {
				// 4) Non-lexical filler => just preserve as is
				fillerTkn := &Tkn{
					Tkn: *tkn, // embed the original Tkn fields
				}
				tsw.Append(fillerTkn)
			}
		}
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
		panic(fmt.Sprintf("failed to register ichiran provider: %w", err))
	}
	err = common.SetDefault(Lang, []common.ProviderEntry{IchiranEntry})
	if err != nil {
		panic(fmt.Sprintf("failed to set ichiran as default: %w", err))
	}
	
	ichiranScheme := common.TranslitScheme{
		Name: "Hepburn",
		Description: "Hepburn romanization",
		Provider: "ichiran",
		NeedsDocker: true,
	}
	if err := common.RegisterScheme(Lang, ichiranScheme); err != nil {
		common.Log.Warn().Msg("Failed to register scheme " + ichiranScheme.Name)
	}
}

func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
