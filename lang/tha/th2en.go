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
	"github.com/go-rod/rod/lib/launcher"
	
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
	config           map[string]interface{}
	browser          *rod.Browser
	targetScheme     string
	progressCallback common.ProgressCallback
}

// SaveConfig merely stores the config to apply after init
func (p *TH2ENProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}


// InitWithContext initializes with the provided context
func (p *TH2ENProvider) InitWithContext(ctx context.Context) (err error) {
	// Get a browser instance (either via BrowserAccessURL or automatic download)
	var browserURL string

	if common.BrowserAccessURL == "" {
		// Use launcher to automatically download and manage browser
		logger.Info().Msg("BrowserAccessURL not set, using automatic browser management")

		// Create a new launcher that will download the browser if needed
		l := launcher.New()
		// Configure the launcher with appropriate options
		l = l.Headless(true)

		// Launch the browser and get the WebSocket URL
		url, err := l.Launch()
		if err != nil {
			return fmt.Errorf("failed to launch browser automatically: %w", err)
		}

		browserURL = url
		logger.Info().Str("browser_url", url).Msg("Browser launched automatically")
	} else {
		// Use provided BrowserAccessURL
		browserURL = common.BrowserAccessURL
		logger.Info().Str("browser_url", browserURL).Msg("Using provided browser URL")
	}

	// Initialize browser with proper error handling
	p.browser = rod.New().ControlURL(browserURL).Context(ctx)
	if p.browser == nil {
		return fmt.Errorf("failed to create browser instance")
	}

	// Connect to the browser - this is a critical step
	if err = p.browser.Connect(); err != nil {
		return fmt.Errorf("go-rod failed to connect to browser: %w", err)
	}

	// Apply config only after successful connection
	if err = p.applyConfig(ctx); err != nil {
		p.browser.Close() // Clean up on error
		p.browser = nil
		return fmt.Errorf("failed to apply config: %w", err)
	}

	return nil
}


// Init initializes with background context
func (p *TH2ENProvider) Init() (err error) {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes with the provided context
func (p *TH2ENProvider) InitRecreateWithContext(ctx context.Context, noCache bool) (err error) {
	return p.InitWithContext(ctx)
}

// InitRecreate reinitializes with background context
func (p *TH2ENProvider) InitRecreate(bool) (err error) {
	return p.Init()
}

// init initializes the provider with the given context
func (p *TH2ENProvider) init(ctx context.Context) (err error) {
	// Check if BrowserAccessURL is available
	if common.BrowserAccessURL == "" {
		return fmt.Errorf("BrowserAccessURL is not set - required for web scraping")
	}

	// Initialize browser with proper error handling
	p.browser = rod.New().ControlURL(common.BrowserAccessURL).Context(ctx)
	if p.browser == nil {
		return fmt.Errorf("failed to create browser instance")
	}
	
	// Connect to the browser - this is a critical step
	if err = p.browser.Connect(); err != nil {
		return fmt.Errorf("go-rod failed to connect to browser: %w", err)
	}
	
	// Apply config only after successful connection
	if err = p.applyConfig(ctx); err != nil {
		p.browser.Close() // Clean up on error
		p.browser = nil
		return fmt.Errorf("failed to apply config: %w", err) 
	}
	
	return nil
}


// applyConfig applies the stored configuration to the provider.
// This includes selecting the transliteration scheme if specified.
// The context is used for cancellation during configuration.
//
// Returns an error if configuration application fails or the context is canceled.
func (p *TH2ENProvider) applyConfig(ctx context.Context) error {
	if p.config == nil {
		return nil
	}
	targetScheme, ok := p.config["scheme"].(string)
	if !ok {
		return fmt.Errorf("scheme name not provided in config")
	}
	if err := p.selectTranslitScheme(ctx, targetScheme); err != nil {
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
	return 120
}

// CloseWithContext closes the provider with the given context
func (p *TH2ENProvider) CloseWithContext(ctx context.Context) error {
	if p.browser != nil {
		return p.browser.Context(ctx).Close()
	}
	return nil
}

// Close closes the provider with background context
func (p *TH2ENProvider) Close() error {
	return p.CloseWithContext(context.Background())
}


func (p *TH2ENProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// selectTranslitScheme selects the transliteration scheme with provided context
func (p *TH2ENProvider) selectTranslitScheme(ctx context.Context, scheme string) error {
	// Protect against nil browser
	if p.browser == nil {
		return fmt.Errorf("browser not initialized, call Init first")
	}

	// Create a derived context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Normalize the input scheme
	scheme = strings.ToLower(strings.TrimSpace(scheme))

	// Validate the scheme
	if !slices.Contains(common.GetSchemesNames(translitSchemes), scheme) {
		return fmt.Errorf("invalid transliteration scheme: %s", scheme)
	}
	
	logger.Trace().Msg("Creating new page")
	// IMPORTANT: We use the original browser instance directly, not a new one with context
	// The context is already set in the main browser instance during init
	// Trying to slap a new one on top will cause runtime panics
	page, err := p.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

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
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("context cancelled while trying to click settings button: %v", ctxWithTimeout.Err())
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
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("context cancelled while waiting for dialog: %w", ctxWithTimeout.Err())
	case <-time.After(500 * time.Millisecond):
	}

	logger.Trace().Msgf("Looking for radio button with value %s and clicking via JavaScript", scheme)
	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("context cancelled while trying to click radio button: %w", ctxWithTimeout.Err())
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
			return fmt.Errorf("failed to click radio button via JavaScript: %w", err)
		}
	}

	logger.Trace().Msg("Successfully changed transliteration scheme")
	return nil
}


// ProcessFlowController processes input with the given context
func (p *TH2ENProvider) ProcessFlowController(ctx context.Context, input common.AnyTokenSliceWrapper) (results common.AnyTokenSliceWrapper, err error) {
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
			return p.process(ctx, raw)
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

// process processes chunks with the given context
func (p *TH2ENProvider) process(ctx context.Context, chunks []string) (common.AnyTokenSliceWrapper, error) {
	tsw := &TknSliceWrapper{}
	totalChunks := len(chunks)
	
	for idx, chunk := range chunks {
		chunks[idx] = reRepetitionMark.ReplaceAllString(chunk, "$1")
	}
	
	for idx, chunk := range chunks {
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		logger.Trace().Msgf("Processing chunk %d/%d: %s", idx+1, totalChunks, chunk)
		
		// IMPORTANT: We use the original browser instance directly, not a new one with context
		// The context is already set in the main browser instance during init
		// Trying to slap a new one on top will cause runtime panics
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

		providerTokenSlice := []string{}
		dicTlit := make(map[string]string)
		dicGloss := make(map[string][]common.Gloss)
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
		// Simple interleaving of the strings (joined chunks) that
		//	- allows to discriminate true lexical content from what isn't
		//	- retain non-lexical content, properly tagged
		
		// IMPORTANT: keep this in the for loop to prevent mysterious bug, see commit msg 6bf9a50
		tkns, err := common.IntegrateProviderTokensV2(chunk, providerTokenSlice)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Token integration had issues, romanization may be incomplete")
			// Continue despite errors - we still want to return partial results
		}


		for _, tkn := range tkns {
			tkn.Romanization = dicTlit[tkn.Surface]
			tkn.Glosses = dicGloss[tkn.Surface]
			tsw.Append(tkn)
		}

		// Close page after processing
		if err := page.Close(); err != nil {
			logger.Warn().Err(err).Msg("failed to close page")
		}
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
		Provider:     &TH2ENProvider{},
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

	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn().Err(err).Msg("Could not reach thai2english.com - will attempt to proceed anyway using automatic browser management")
		return nil // Return nil instead of error to allow automatic browser management to try
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn().Int("status_code", resp.StatusCode).Msg("Website returned non-200 status code - will attempt to proceed anyway")
		return nil // Return nil to allow automatic browser management to try
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
