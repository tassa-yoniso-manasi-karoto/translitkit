package tha

import (
	"context"
	"fmt"
	"strings"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"github.com/tassa-yoniso-manasi-karoto/paiboonizer"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// =============================================================================
// DOCKER CONTAINER LIFECYCLE - IMPORTANT FOR FUTURE DEVELOPERS/LLMs
// =============================================================================
//
// PaiboonizerProvider does NOT own any Docker container lifecycle.
// It is a TRANSLITERATOR-ONLY provider that depends on pythainlp for syllable
// tokenization.
//
// CRITICAL: This provider uses go-pythainlp's PACKAGE-LEVEL functions
// (e.g., pythainlp.SyllableTokenize()) instead of creating its own manager.
// This ensures it reuses any existing Docker container started by PyThaiNLPProvider.
//
// In hybrid schemes like "paiboon-hybrid":
//   1. PyThaiNLPProvider initializes first → starts Docker container
//   2. PaiboonizerProvider initializes → NO new container (uses existing)
//   3. During processing: pythainlp does word tokenization, paiboonizer does transliteration
//   4. PaiboonizerProvider closes → NO container action (doesn't own it)
//   5. PyThaiNLPProvider closes → container stops
//
// This design prevents lifecycle conflicts. DO NOT change this to create
// a pythainlp.PyThaiNLPManager - that would cause container conflicts.
//
// Accuracy: ~83% on dictionary dataset (experimental, fast, fully local)
//
// =============================================================================

// PaiboonizerProvider implements the Provider interface for Thai using paiboonizer
// It operates as a transliterator only (requires tokenized input from pythainlp)
type PaiboonizerProvider struct {
	config           map[string]interface{}
	progressCallback common.ProgressCallback
	// NOTE: No pythainlp manager here - we use package-level functions
}

// NewPaiboonizerProvider creates a new provider
func NewPaiboonizerProvider() *PaiboonizerProvider {
	return &PaiboonizerProvider{
		config: make(map[string]interface{}),
	}
}

// SaveConfig stores configuration for later application during initialization
func (p *PaiboonizerProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}

// InitWithContext initializes the provider with context
// NOTE: This does NOT start any Docker container - we rely on PyThaiNLPProvider
// having already started the pythainlp container in hybrid schemes.
func (p *PaiboonizerProvider) InitWithContext(ctx context.Context) error {
	// No manager creation needed!
	// Paiboonizer uses go-pythainlp's package-level functions which
	// automatically reuse any existing container via the default manager.
	//
	// See lifecycle comments at top of file for details.
	return nil
}

// Init initializes the provider with background context
func (p *PaiboonizerProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider
func (p *PaiboonizerProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	// Nothing to recreate - we don't own any resources
	return nil
}

// InitRecreate reinitializes with background context
func (p *PaiboonizerProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

// CloseWithContext releases resources
// NOTE: This does NOT stop any Docker container - PyThaiNLPProvider owns that.
func (p *PaiboonizerProvider) CloseWithContext(ctx context.Context) error {
	// Nothing to close - we don't own any resources
	return nil
}

// Close releases resources with background context
func (p *PaiboonizerProvider) Close() error {
	return p.CloseWithContext(context.Background())
}

// ProcessFlowController processes input tokens for transliteration
func (p *PaiboonizerProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	// Paiboonizer only supports transliteration mode
	if mode != common.TransliteratorMode {
		return nil, fmt.Errorf("paiboonizer only supports transliterator mode, got %s", mode)
	}

	// Check if we have tokenized input
	if input.Len() == 0 {
		return nil, fmt.Errorf("paiboonizer requires tokenized input")
	}

	tsw := &TknSliceWrapper{}
	totalTokens := input.Len()

	// Track previous romanization for ๆ (mai yamok) handling
	// When pythainlp word tokenizer returns ๆ as a separate token,
	// we need to repeat the previous word's romanization
	var lastRomanization string

	// Process each token
	for i := 0; i < totalTokens; i++ {
		if p.progressCallback != nil {
			p.progressCallback(i, totalTokens)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get the token using GetIdx (AnyTokenSliceWrapper interface)
		token := input.GetIdx(i)
		if token == nil {
			continue
		}

		// Create Thai token
		thaiToken := &Tkn{
			Tkn: common.Tkn{
				Surface:   token.GetSurface(),
				IsLexical: token.IsLexicalContent(),
			},
		}

		// Transliterate if it's a lexical token with Thai text
		if token.IsLexicalContent() {
			text := token.GetSurface()

			// Handle ๆ (mai yamok) as standalone token from word tokenizer
			// When pythainlp splits "แคบๆ" into ["แคบ", "ๆ"], repeat previous
			if text == "ๆ" {
				if lastRomanization != "" {
					// Get last syllable from previous romanization to repeat
					lastParts := strings.Split(lastRomanization, "-")
					lastSyl := lastParts[len(lastParts)-1]
					thaiToken.Romanization = lastSyl
				}
			} else if containsThai(text) {
				romanized := p.transliterateWord(ctx, text)
				thaiToken.Romanization = romanized
				lastRomanization = romanized
			} else {
				// Non-Thai text passes through unchanged
				thaiToken.Romanization = text
			}
		}

		tsw.Append(thaiToken)
	}

	return tsw, nil
}

// transliterateWord transliterates a single Thai word.
// Flow:
//   1. Handle ๆ (mai yamok) repetition marker at word level
//   2. Check the word dictionary (~5000 entries) for exact match
//   3. If not found, use pythainlp syllable tokenization + paiboonizer rules
//
// IMPORTANT: Uses package-level pythainlp.SyllableTokenize() to reuse existing container.
func (p *PaiboonizerProvider) transliterateWord(ctx context.Context, word string) string {
	// STEP 0: Handle ๆ (mai yamok) at word level
	// Words like "ชิ้นๆ" should become "chín-chín"
	// This handles cases where pythainlp doesn't separate ๆ as its own syllable
	if strings.HasSuffix(word, "ๆ") {
		baseWord := strings.TrimSuffix(word, "ๆ")
		if baseWord != "" {
			baseTrans := p.transliterateWord(ctx, baseWord)
			if baseTrans != "" {
				// Get the last syllable to repeat
				lastParts := strings.Split(baseTrans, "-")
				lastSyl := lastParts[len(lastParts)-1]
				return baseTrans + "-" + lastSyl
			}
		}
	}

	// STEP 1: Check word dictionary first (has ~5000 whole word entries)
	// This handles common words like หน้าต่าง → nâa-dtàang correctly
	if trans, found := paiboonizer.LookupDictionary(word); found {
		return trans
	}

	// STEP 2: Word not in dictionary - use pythainlp syllable tokenization
	// Use go-pythainlp's package-level function - this reuses the default manager
	// which connects to the already-running Docker container started by PyThaiNLPProvider.
	result, err := pythainlp.SyllableTokenize(word)
	if err != nil || result == nil || len(result.Syllables) == 0 {
		// Fall back to pure rule-based transliteration using paiboonizer package
		return paiboonizer.ComprehensiveTransliterate(word)
	}

	// STEP 3: Transliterate each syllable using the paiboonizer package
	var parts []string
	var lastTrans string

	for _, syllable := range result.Syllables {
		// Handle ๆ (mai yamok) - repeat previous syllable
		// This catches cases where pythainlp returns ๆ as separate syllable
		if syllable == "ๆ" {
			if lastTrans != "" {
				parts = append(parts, lastTrans)
			}
			continue
		}

		// Check if syllable ends with ๆ (pythainlp might not separate it)
		if strings.HasSuffix(syllable, "ๆ") {
			baseSyl := strings.TrimSuffix(syllable, "ๆ")
			cleanSyl := paiboonizer.RemoveSilentConsonants(baseSyl)
			if cleanSyl != "" {
				trans := p.transliterateSyllable(cleanSyl)
				if trans != "" {
					parts = append(parts, trans)
					parts = append(parts, trans) // Repeat for ๆ
					lastTrans = trans
				}
			}
			continue
		}

		// Strip trailing silent consonants (consonant + ์) before lookup
		cleanSyllable := paiboonizer.RemoveSilentConsonants(syllable)
		if cleanSyllable == "" {
			continue
		}

		trans := p.transliterateSyllable(cleanSyllable)
		if trans != "" {
			parts = append(parts, trans)
			lastTrans = trans
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "-")
}

// transliterateSyllable transliterates a single syllable using dictionary lookup then rules
func (p *PaiboonizerProvider) transliterateSyllable(syllable string) string {
	// Try syllable dictionary first, then special cases, then rules
	if t, found := paiboonizer.LookupSyllable(syllable); found {
		return t
	}
	if t, found := paiboonizer.LookupSpecialCase(syllable); found {
		return t
	}
	// Use the paiboonizer package's comprehensive transliteration
	return paiboonizer.ComprehensiveTransliterate(syllable)
}

// Note: RemoveSilentConsonants and other helper functions are provided by
// the paiboonizer package. See paiboonizer.RemoveSilentConsonants().

// WithProgressCallback sets the progress callback
func (p *PaiboonizerProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// Name returns the provider name
func (p *PaiboonizerProvider) Name() string {
	return "paiboonizer"
}

// SupportedModes returns the operating modes this provider supports
func (p *PaiboonizerProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TransliteratorMode}
}

// GetMaxQueryLen returns the maximum query length
func (p *PaiboonizerProvider) GetMaxQueryLen() int {
	// Paiboonizer can handle any length since it processes token by token
	return 0
}

// containsThai checks if a string contains Thai characters
func containsThai(text string) bool {
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	return false
}

// Note: Dictionaries and transliteration rules are provided by the paiboonizer package.
// See github.com/tassa-yoniso-manasi-karoto/paiboonizer for the full implementation.
