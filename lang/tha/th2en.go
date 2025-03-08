package tha

import (
	"fmt"
	"net/url"
	"net/http"
	"strings"
	"slices"
	"time"
	"context"
	"regexp"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)


var (
	logger = common.Log.With().Str("provider", "thai2english.com").Logger()
	reRepetitionMark = regexp.MustCompile(`\s+(ๆ)`)
)


// TH2ENProvider satisfies the Provider interface
type TH2ENProvider struct {
	ctx              context.Context
	config           map[string]interface{}
	browser          *rod.Browser
	targetScheme     string
	progressCallback common.ProgressCallback
}


func (p *TH2ENProvider) WithContext(ctx context.Context) {
	p.ctx = ctx
}

// SaveConfig merely stores the config to apply after init
func (p *TH2ENProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}


func (p *TH2ENProvider) Init() (err error) {
	// force refresh logger
	logger = common.Log.With().Str("provider", "thai2english.com").Logger()
	if err = checkWebsiteReachable(p.ctx); err != nil {
		return
	}
	if err = p.init(); err != nil {
		return
	}
	if p.targetScheme == "" {
		if err = p.selectTranslitScheme("paiboon"); err != nil {
			return fmt.Errorf("error selecting default translit scheme: %w", err)
		}
	}
	return
}

func (p *TH2ENProvider) InitRecreate(bool) (err error) {
	return p.Init()
}

func (p *TH2ENProvider) init() (err error) {
	p.browser = rod.New()
	if err = p.browser.ControlURL(common.BrowserAccessURL).Connect(); err != nil {
		return fmt.Errorf("go-rod failed to connect to a browser instance for scraping: %w", err)
	}
	p.applyConfig()
	return
}

func (p *TH2ENProvider) applyConfig() error {
	if p.config == nil {
		return nil
	}
	targetScheme, ok := p.config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}
	if err := p.selectTranslitScheme(targetScheme); err != nil {
		return fmt.Errorf("error selecting translit scheme %s: %w", targetScheme, err)
	}

	p.targetScheme = targetScheme
	return nil
}

func (p *TH2ENProvider) Name() string {
	return "thai2english.com"
}

func (p *TH2ENProvider) GetType() common.ProviderType {
	return common.CombinedType
}

func (p *TH2ENProvider) GetMaxQueryLen() int {
	return 99
}

func (p *TH2ENProvider) Close() error {
	return p.browser.Close()
}

func (p *TH2ENProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

func (p *TH2ENProvider) selectTranslitScheme(scheme string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
	defer cancel()
	
	// Normalize the input scheme
	scheme = strings.ToLower(strings.TrimSpace(scheme))

	// Validate the scheme
	if !slices.Contains(common.GetSchemesNames(translitSchemes), scheme) {
		return fmt.Errorf("invalid transliteration scheme: %s", scheme)
	}
	
	logger.Trace().Msg("Creating new page")
	page, err := p.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	
	// TODO
	page = page.Context(ctx)

	logger.Trace().Msg("Navigating to website")
	if err := page.Navigate("https://www.thai2english.com/"); err != nil {
		return fmt.Errorf("failed to navigate to website: %w", err)
	}

	logger.Trace().Msg("Waiting for page to load")
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	logger.Trace().Msg("Looking for settings button and clicking via JavaScript")
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while trying to click settings button: %v", ctx.Err())
	default:
		_, err = page.Eval(`() => {
			const buttons = Array.from(document.querySelectorAll('button'));
			const settingsBtn = buttons.find(btn => btn.textContent.includes('Settings'));
			if (!settingsBtn) {
				throw new Error('Settings button not found');
			}
			settingsBtn.click();
			return true;
		}`)
		if err != nil {
			return fmt.Errorf("failed to click settings button via JavaScript: %w", err)
		}
	}

	logger.Trace().Msg("Waiting for dialog to appear")
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for dialog: %w", ctx.Err())
	case <-time.After(500 * time.Millisecond):
	}

	logger.Trace().Msgf("Looking for radio button with value %s and clicking via JavaScript", scheme)
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while trying to click radio button: %w", ctx.Err())
	default:
		_, err = page.Eval(fmt.Sprintf(`() => {
			const radio = document.querySelector('input[type="radio"][value="%s"]');
			if (!radio) {
				throw new Error('Radio button not found');
			}
			radio.click();
			return true;
		}`, scheme))
		if err != nil {
			return fmt.Errorf("failed to click radio button via JavaScript: %w", scheme, err)
		}
	}

	logger.Trace().Msg("Successfully changed transliteration scheme")
	return nil
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
	providerTokenSlice := []string{}
	dicTlit := make(map[string]string)
	dicGloss := make(map[string][]common.Gloss)
	totalChunks := len(chunks)
	
	for idx, chunk := range chunks {
		chunks[idx] = reRepetitionMark.ReplaceAllString(chunk, "$1")
	}
	
	for idx, chunk := range chunks {
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		if err := p.ctx.Err(); err != nil {
			return nil, err
		}

		logger.Trace().Msgf("Processing chunk %d/%d: %s", idx+1, totalChunks, chunk)
		
		page, err := p.browser.Page(proto.TargetCreateTarget{})
		if err != nil {
			return nil, fmt.Errorf("failed to create page: %w", err)
		}
		defer page.Close()

		logger.Trace().Msg("Navigate to URL")
		url := fmt.Sprintf("https://www.thai2english.com/?q=%s", url.QueryEscape(chunk))
		if err := page.Navigate(url); err != nil {
			return nil, fmt.Errorf("failed to navigate to URL: %w", err)
		}

		// Waits for the `window.onload` event
		logger.Trace().Msg("Wait for page load")
		if err := page.WaitLoad(); err != nil {
			return nil, fmt.Errorf("failed to wait for page load: %w", err)
		}

		// Waits until all network requests including dynamic requests
		// (AJAX, fetch, or WebSockets) stop for a set duration
		logger.Trace().Msg("Wait for RequestIdle (300 ms)")
		page.MustWaitRequestIdle()
		
		
		logger.Trace().Msg("Wait for main element to be present")
		_, err = page.Element(".word-breakdown_line-meanings__1RADe")
		if err != nil {
			return nil, fmt.Errorf("failed to find main element: %w", err)
		}

		logger.Trace().Msg("Get all meaning elements")
		elements, err := page.Elements(".word-breakdown_line-meaning__NARMM")
		if err != nil {
			return nil, fmt.Errorf("failed to get meaning elements: %w", err)
		}
		if len(elements) == 0 {
			return tsw, fmt.Errorf("elements are empty. idx=%d", idx)
		}

		// Process each element
		for _, element := range elements {
			thNode, err := element.Element(".thai")
			if err != nil {
				// seems to be caused by punctuation
				//logger.Warn().Err(err).Msg("no Thai element exists, skipping")
				continue
			}
			th, err := thNode.Text()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to get Thai text, skipping")
				continue
			}
			providerTokenSlice = append(providerTokenSlice, th)
			
			tlitNode, err := element.Element(".tlit")
			if err != nil {
				logger.Warn().Err(err).Msg("no transliteration element exists, skipping")
				continue
			}
			tlit, err := tlitNode.Text()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to get transliteration text, skipping")
				continue
			}
			dicTlit[th] = tlit
			
			// Get gloss
			glossNode, err := element.Element(".meanings")
			if err != nil {
				logger.Warn().Err(err).Msg("no gloss element exists, skipping")
				continue
			}
			glossText, err := glossNode.Text()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to get gloss text, skipping")
				continue
			}

			// Process gloss text
			glossRaw := strings.Split(glossText, "\n")
			glossRaw = removeEmptyStrings(glossRaw)
			
			for _, gloss := range glossRaw {
				dicGloss[th] = append(dicGloss[th], common.Gloss{
					Definition: gloss,
				})
			}
		}

		// Close page after processing
		if err := page.Close(); err != nil {
			logger.Warn().Err(err).Msg("failed to close page")
		}
	}
	// Simple interleaving of the strings (joined chunks) that
	//	- allows to discriminate true lexical content from what isn't
	//	- retain non-lexical content, properly tagged
	tkns := common.IntegrateProviderTokens(strings.Join(chunks, ""), providerTokenSlice)

	for _, tkn := range tkns {
		tkn.Romanization = dicTlit[tkn.Surface]
		tkn.Glosses = dicGloss[tkn.Surface]
		tsw.Append(tkn)
	}
	
	// Report completion (all chunks processed)
	if p.progressCallback != nil {
		p.progressCallback(totalChunks, totalChunks)
	}
	
	return tsw, nil
}



var translitSchemes = []common.TranslitScheme{
	{ Name:"paiboon", Description:"Paiboon-esque transliteration"},
	{ Name:"thai2english", Description: "thai2english's custom transliteration system"},
	{ Name:"rtgs", Description: "Royal Thai General System of transcription"},
	{ Name: "ipa", Description:"International Phonetic Alphabet representation"},
	{ Name:"simplified-ipa",Description:"Simplified phonetic notation"},
}

func init() {
	p := common.ProviderEntry{
		Provider:     &TH2ENProvider{ ctx: context.Background() },
		Capabilities: []string{"tokenization", "transliteration"},
		Type:         common.CombinedType,
	}
	err := common.Register(Lang, common.CombinedType, p.Provider.Name(), p)
	if err != nil {
		panic(fmt.Sprintf("failed to register %s provider: %w", p.Provider.Name(), err))
	}
	err = common.SetDefault(Lang, []common.ProviderEntry{p})
	if err != nil {
		panic(fmt.Sprintf("failed to set %s as default: %w", p.Provider.Name(), err))
	}
	
	for _, scheme := range translitSchemes {
		scheme.Provider = p.Provider.Name()
		scheme.NeedsScraper = true
		if err := common.RegisterScheme(Lang, scheme); err != nil {
			common.Log.Warn().
				Str("pkg", Lang).
				Msg("Failed to register scheme " + scheme.Name)
		}
	}

}


func checkWebsiteReachable(ctx context.Context) error {
	URL := "https://www.thai2english.com/"
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", URL,  nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach th2en: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("website returned non-200 status code: %d", resp.StatusCode)
	}
	
	return nil
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





func placeholder333() {
	fmt.Print("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
