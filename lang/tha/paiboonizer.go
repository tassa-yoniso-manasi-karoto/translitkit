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

	totalTokens := input.Len()

	// =======================================================================
	// TOKENIZATION CORRECTION PASS
	// =======================================================================
	// Collect lexical tokens, apply correction, then map back to indices.
	// This fixes pythainlp segmentation errors before transliteration.

	// Step 1: Collect lexical token indices and surfaces
	type lexicalInfo struct {
		index   int
		surface string
	}
	var lexicals []lexicalInfo
	for i := 0; i < totalTokens; i++ {
		token := input.GetIdx(i)
		if token != nil && token.IsLexicalContent() {
			lexicals = append(lexicals, lexicalInfo{index: i, surface: token.GetSurface()})
		}
	}

	// Step 2: Extract surfaces and apply correction
	surfaces := make([]string, len(lexicals))
	for i, lex := range lexicals {
		surfaces[i] = lex.surface
	}
	correctedSurfaces := correctTokenization(surfaces)

	// Step 3: Build mapping from original index to corrected surface
	// If correction merged tokens, some indices will map to "" (skip)
	correctedMap := make(map[int]string)

	// After correction, we may have fewer surfaces than original lexicals.
	// Walk through original lexicals and match with corrected.
	correctedIdx := 0
	for i := 0; i < len(lexicals); i++ {
		if correctedIdx >= len(correctedSurfaces) {
			// This token was merged away
			correctedMap[lexicals[i].index] = ""
			continue
		}

		// Check if this token matches the current corrected surface
		// or if it was merged into the previous one
		if correctedSurfaces[correctedIdx] == lexicals[i].surface {
			// Unchanged
			correctedMap[lexicals[i].index] = lexicals[i].surface
			correctedIdx++
		} else if i > 0 && strings.HasSuffix(correctedSurfaces[correctedIdx-1], lexicals[i].surface) {
			// This token was merged into previous - skip it
			correctedMap[lexicals[i].index] = ""
		} else {
			// The corrected surface is different (merged or modified)
			correctedMap[lexicals[i].index] = correctedSurfaces[correctedIdx]
			correctedIdx++
		}
	}

	// =======================================================================
	// TRANSLITERATION PASS
	// =======================================================================

	tsw := &TknSliceWrapper{}

	// Track previous romanization for ๆ (mai yamok) handling
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

		token := input.GetIdx(i)
		if token == nil {
			continue
		}

		// Check if this lexical token should be skipped (merged into previous)
		if token.IsLexicalContent() {
			if corrected, ok := correctedMap[i]; ok && corrected == "" {
				// Token was merged - skip it entirely
				continue
			}
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
			// Use corrected surface if available
			text := token.GetSurface()
			if corrected, ok := correctedMap[i]; ok && corrected != "" {
				text = corrected
				thaiToken.Surface = corrected // Update surface to corrected form
			}

			// Handle ๆ (mai yamok) as standalone token from word tokenizer
			if text == "ๆ" {
				if lastRomanization != "" {
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

// WithDownloadProgressCallback sets a callback for download progress (no-op for Paiboonizer).
func (p *PaiboonizerProvider) WithDownloadProgressCallback(callback common.DownloadProgressCallback) {
	// No-op: Paiboonizer is a pure Go implementation, doesn't require Docker downloads
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

// =============================================================================
// TOKENIZATION CORRECTION
// =============================================================================
//
// pythainlp word tokenizer sometimes incorrectly segments closing consonants:
//   - Pattern A: Consonant split off as isolated token (แม่ง → ["แม่", "ง"])
//   - Pattern B: Consonant attached to next word (บอกว่า → ["บอ", "กว่า"])
//
// These functions post-process pythainlp's output to fix common errors.
// =============================================================================

// closingConsonants are Thai consonants that commonly appear as word-final sounds.
// When we see one of these as an isolated single-character token, it's likely
// a pythainlp segmentation error.
var closingConsonants = map[rune]bool{
	'ง': true, // ng - very common final
	'น': true, // n - common final
	'ม': true, // m - common final
	'ก': true, // k - common final
	'บ': true, // p - common final
	'ด': true, // t - common final
	'ย': true, // y - in some words
	'ว': true, // w - in diphthongs
}

// knownMissegmentation describes a word that pythainlp commonly splits incorrectly.
type knownMissegmentation struct {
	fullWord  string // The correct merged word
	splitChar rune   // The consonant that gets incorrectly attached to next word
}

// knownMissegmentations maps truncated forms to their correct full forms.
// Used to fix Pattern B errors where closing consonant attaches to next word.
var knownMissegmentations = map[string]knownMissegmentation{
	"บอ": {"บอก", 'ก'}, // บอกว่า → ["บอ", "กว่า"] should be ["บอก", "ว่า"]
	// Add more as discovered from test failures
}

// isSingleThaiConsonant checks if the string is exactly one Thai consonant.
func isSingleThaiConsonant(s string) (rune, bool) {
	runes := []rune(s)
	if len(runes) != 1 {
		return 0, false
	}
	r := runes[0]
	// Thai consonants range: ก (0x0E01) to ฮ (0x0E2E)
	if r >= 'ก' && r <= 'ฮ' {
		return r, true
	}
	return 0, false
}

// correctTokenization fixes common pythainlp word segmentation errors.
// It modifies the input slice in place and returns it.
func correctTokenization(tokens []string) []string {
	if len(tokens) < 2 {
		return tokens
	}

	// Pattern A: Merge isolated closing consonants back into previous word
	// e.g., ["แม่", "ง"] → ["แม่ง"]
	i := 1
	for i < len(tokens) {
		consonant, isSingle := isSingleThaiConsonant(tokens[i])
		if isSingle && closingConsonants[consonant] {
			candidate := tokens[i-1] + tokens[i]
			// Only merge if the result is a known dictionary word
			if _, found := paiboonizer.LookupDictionary(candidate); found {
				tokens[i-1] = candidate
				tokens = append(tokens[:i], tokens[i+1:]...)
				// Don't increment i - check same position again
				continue
			}
		}
		i++
	}

	// Pattern B: Fix known missegmentations where consonant attaches to next word
	// e.g., ["บอ", "กว่า"] → ["บอก", "ว่า"]
	for i := 0; i < len(tokens)-1; i++ {
		fix, ok := knownMissegmentations[tokens[i]]
		if !ok {
			continue
		}

		nextRunes := []rune(tokens[i+1])
		if len(nextRunes) == 0 {
			continue
		}

		// Check if next token starts with the expected split character
		if nextRunes[0] != fix.splitChar {
			continue
		}

		// Get remainder after removing the split character
		remainder := string(nextRunes[1:])

		// Only fix if remainder is non-empty and contains Thai
		// (empty remainder would mean the whole next token was just the consonant)
		if len(remainder) > 0 && containsThai(remainder) {
			tokens[i] = fix.fullWord
			tokens[i+1] = remainder
		}
	}

	return tokens
}

// Note: Dictionaries and transliteration rules are provided by the paiboonizer package.
// See github.com/tassa-yoniso-manasi-karoto/paiboonizer for the full implementation.
