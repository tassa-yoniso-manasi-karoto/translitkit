package tha

import (
	"fmt"
	"net/url"
	"strings"
	
	"github.com/go-rod/rod"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// TH2ENProvider satisfies the Provider interface
type TH2ENProvider struct {
	config map[string]interface{}
	browser *rod.Browser
}

func (p *TH2ENProvider) Init() (err error) {
	p.browser = rod.New()
	if err = p.browser.ControlURL(common.BrowserAccessURL).Connect(); err != nil {
		return fmt.Errorf("go-rod failed to connect to a browser instance for scraping: %v", err)
	}
	// TODO Use config to scrape & select desired romanization
	return
}

func (p *TH2ENProvider) Name() string {
	return "th2en"
}

func (p *TH2ENProvider) GetType() common.ProviderType {
	return common.CombinedType
}

func (p *TH2ENProvider) GetMaxQueryLen() int {
	return 999
}

func (p *TH2ENProvider) Close() error {
	return p.browser.Close()
}


func (p *TH2ENProvider) ProcessFlowController(input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
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
		input.ClearRaw()
	} else { // generic token processor: take common.Tkn as well as lang-specic tokens that have common.Tkn as their embedded field
		switch ProviderType {
		case common.TokenizerType:
			// Either refuse because it is already tokenized or add linguistic annotations
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a tokenizer")
		case common.TransliteratorType:
		case common.CombinedType:
			// Refuse because it is already tokenized
			return nil, fmt.Errorf("not implemented atm: Tokens is not accepted as input type for a provider that combines tokenizer+transliterator")
		}
	}
	return nil, fmt.Errorf("handling not implemented for '%s' with ProviderType '%s'", p.Name(), ProviderType)
}


func (p *TH2ENProvider) process(chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &TknSliceWrapper{}
	for idx, chunk := range chunks {
		page := p.browser.MustPage(fmt.Sprintf("https://www.thai2english.com/?q=%s", url.QueryEscape(chunk))).MustWaitLoad()
		page.MustWaitRequestIdle()
		page.MustElement(".word-breakdown_line-meanings__1RADe")
		elements := page.MustElements(".word-breakdown_line-meaning__NARMM")
		if len(elements) == 0 {
			return tsw, fmt.Errorf("elements are empty. idx=%d", idx)
		}
		for _, element := range elements {
			thNode, err       :=  element.Element(".thai")
			if err != nil {
				continue
			}
			th := thNode.MustText()
			
			tlitNode, err := element.Element(".tlit")
			if err != nil {
				color.Redln("no transliteration element exists")
				continue
			}
			tlit := tlitNode.MustText()
			
			glossNode, err := element.Element(".meanings")
			if err != nil {
				color.Redln("no gloss element exists")
				continue
			}
			glossRaw := strings.Split(glossNode.MustText(), "\n")
			// there is always a final newline → would create empty gloss
			glossRaw = removeEmptyStrings(glossRaw)
			var glossSlice []common.Gloss
			for _, gloss := range glossRaw {
				glossSlice = append(glossSlice, common.Gloss{
					Definition: gloss,
				})
			}
			tsw.Append(Tkn{ Tkn: common.Tkn{
				Surface: th,
				Romanization: tlit,
				IsToken: true,
				Glosses: glossSlice,
			}})
		}
		page.MustClose()
	}
	return tsw, nil
}



func init() {
	TH2ENentry := common.ProviderEntry{
		Provider:     &TH2ENProvider{},
		Capabilities: []string{"tokenization", "transliteration"},
		Type:         common.CombinedType,
	}
	err := common.Register(Lang, common.CombinedType, "th2en", TH2ENentry)
	if err != nil {
		panic(fmt.Sprintf("failed to register ichiran provider: %v", err))
	}
	err = common.SetDefault(Lang, []common.ProviderEntry{TH2ENentry}) // TODO add robepike/nihongo to force romanization after
	
	if err != nil {
		panic(fmt.Sprintf("failed to set ichiran as default: %v", err))
	}
}


func removeEmptyStrings(strings []string) []string {
	result := make([]string, 0, len(strings))
	for _, str := range strings {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

// func (p *ThaiProvider) AnalyzeWord(word string) ThaiToken {
// 	token := NewThaiToken()

// 	// Example analysis (simplified)
// 	token.Surface = word
// 	token.TokenType = common.WordToken
// 	token.Language = "tha"

// 	// Thai-specific analysis
// 	token.Tone = determineTone(word)
// 	token.ConsonantClass = determineConsonantClass(word)
// 	token.RegisterLevel = determineRegisterLevel(word)

// 	// Set common fields
// 	token.PartOfSpeech = determinePartOfSpeech(word)
// 	token.Romanization = romanize(word)

// 	return token
// }




// // Helper functions (to be implemented)
// func determineTone(word string) int              { /* ... */ return 0 }
// func determineConsonantClass(word string) string { /* ... */ return "" }
// func determineRegisterLevel(word string) string  { /* ... */ return "" }
// func determinePartOfSpeech(word string) string   { /* ... */ return "" }
// func romanize(word string) string                { /* ... */ return "" }



func placeholder333() {
	fmt.Print("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
