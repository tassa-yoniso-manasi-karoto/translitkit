package common

import (
	"strings"
	"unicode"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)


type AnyTokenSliceWrapper interface {
	GetIdx(int)		AnyToken
	GetRaw()		[]string
	ClearRaw()
	Append(...AnyToken)
	Len()			int

	Roman()			string
	RomanParts()		[]string
	Tokenized()		string
	TokenizedParts()	[]string
}

type AnyToken interface {
	GetSurface()		string
	Roman()			string
	SetRoman(string)
	IsLexicalContent()	bool
}

// FilterAny receives any token slice wrapper and returns a new wrapper
// containing only tokens that contain lexical content (ie. it excludes space, punctuations...)
func ToAnyLexicalTokens(wrapper AnyTokenSliceWrapper) AnyTokenSliceWrapper {
	filtered := &TknSliceWrapper{}
	for i := 0; i < wrapper.Len(); i++ {
		token := wrapper.GetIdx(i)
		if token.IsLexicalContent() {
			filtered.Append(token)
		}
	}
	return filtered
}


// Filter receives *common.TknSliceWrapper and returns a new wrapper
// containing only tokens that contain lexical content (ie. it excludes space, punctuations...)
func ToLexicalTokens(wrapper *TknSliceWrapper) *TknSliceWrapper {
	filtered := &TknSliceWrapper{}
	for i := 0; i < wrapper.Len(); i++ {
		token := wrapper.GetIdx(i)
		if token.IsLexicalContent() {
			filtered.Append(token)
		}
	}
	return filtered
}

type TknSliceWrapper struct {
	Slice []AnyToken //alt.: Sentences [][]AnyToken ?
	Raw   []string
}

// TODO maybe make some of these methods private

func (tokens *TknSliceWrapper) GetIdx(i int) AnyToken {
	if len(tokens.Slice) == 0 {
		return nil
	}
	return tokens.Slice[i]
}
func (tokens *TknSliceWrapper) Len() int {
	return len(tokens.Slice)
}
func (tokens *TknSliceWrapper) GetRaw() []string {
	return tokens.Raw
}
func (tokens *TknSliceWrapper) ClearRaw() {
	tokens.Raw = []string{}
}
func (tokens *TknSliceWrapper) Append(tkn ...AnyToken) {
	if tokens.Slice == nil {
		tokens.Slice = make([]AnyToken, 0)
	}
	tokens.Slice = append(tokens.Slice, tkn...)
}


// return the unwrapped slice contained by the wrapper
//func (tokens TknSliceWrapper) Tokens() []AnyToken // FIXME may come in handy?

func (tokens TknSliceWrapper) Roman() string {
	return defaultRoman(tokens.Slice)
}
func (tokens TknSliceWrapper) RomanParts() []string {
	return romanParts(tokens.Slice)
}

func (tokens TknSliceWrapper) Tokenized() string {
	return defaultTokenized(tokens.Slice)
}

func (tokens TknSliceWrapper) TokenizedParts() []string {
	return tokenizedParts(tokens.Slice)
}


// (common.)Tkn represents the common, generic Token containing basic linguistic
// annotations / features for all languages.
// Languages specific token types (ie. jpn.Tkn) all have Tkn as their embedded
// (unnamed) field therefore the methods of Tkn are available for every token
// type regardless.
type Tkn struct {
	// The actual text segment
	Surface    string 
	
	// IsLexicalToken indicates whether this token represents genuine linguistic content,
	// such as a word or phrase recognized by the tokenization provider.
	// A value of false means the token consists of non-lexical elements
	// (e.g., punctuation, spaces, other filler characters...).
	IsLexical    bool
	
	// Normalized form (e.g., lowercase, trimmed)
	Normalized string
	
	// Type of token (word, punctuation, etc.)
	// TokenType  TokenType 
	
	Position struct {
		Start     int // Start position in original text
		End       int // End position in original text
		Sentence  int // Index of containing sentence
		Paragraph int // Index of containing paragraph
	}

	// Linguistic Features
	Romanization  string            // Latin alphabet representation
	Lemma         string            // Base/dictionary form
	PartOfSpeech  string            // Grammatical category (noun, verb, etc.)
	MorphFeatures map[string]string // Morphological features (gender, number, tense, etc.)
	Glosses       []Gloss           // Definitions/meanings with associated metadata

	// Semantic Information
	NamedEntity string  // Named entity type (if applicable)
	Sentiment   float64 // Sentiment score (if applicable)

	// Dependency Parsing
	DependencyRole string // Syntactic role in dependency tree
	HeadPosition   int    // Position of syntactic head

	// Word Composition
	Components []Tkn // For compound words or complex tokens
	IsCompound bool  // Whether this is a compound token

	// Additional Information
	Confidence float64                // Confidence score of the analysis
	Script     string                 // Writing system used (Latin, Cyrillic, etc.)
	Language   string                 // ISO 639-3 code of the token's language
	Metadata   map[string]interface{} // Provider-specific additional data
}


// Some tokenization providers have a lossy tokenization that offers only the core, lexical content.
// IntegrateProviderTokens combines the tokens produced by the provider with the
// intervening text (such as punctuation, spaces, or other characters) that the provider
// did not tokenize by tracking their positions and capturing any gaps as filler tokens.
func IntegrateProviderTokens(original string, providerTokens []string) []*Tkn {
	var result []*Tkn
	pos := 0

	for _, token := range providerTokens {
		// Find the token starting from the current position.
		idx := strings.Index(original[pos:], token)
		if idx == -1 {
			// If the token is not found, skip to the next token.
			continue
		}
		// Adjust index relative to the whole string.
		idx += pos

		// Capture any text between the current position and the token's start as a fake token.
		if pos < idx {
			fake := original[pos:idx]
			result = append(result, &Tkn{Surface: fake, IsLexical: false})
		}

		// Append the provider token.
		result = append(result, &Tkn{Surface: token, IsLexical: true})

		// Update the position after the token.
		pos = idx + len(token)
	}

	// Capture any trailing characters as a fake token.
	if pos < len(original) {
		fake := original[pos:]
		result = append(result, &Tkn{Surface: fake, IsLexical: false})
	}
	return result
}


type Gloss struct {
	PartOfSpeech	string  // Part of speech
	Definition	string  // Definition/meaning
	Info		string  // Additional information
}

func (t *Tkn) GetSurface() string {
	return t.Surface
}

func (t *Tkn) Roman() string {
	if !t.IsLexical || t.Surface == t.Romanization {
		return ""
	}
	return t.Romanization
}

func (t *Tkn) SetRoman(roman string) {
	t.Romanization = roman
}

func (t *Tkn) IsLexicalContent() bool {
	return t.IsLexical
}




// ###########################################################################



func romanParts(tokens []AnyToken) []string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		if r := t.Roman(); r != "" {
			parts[i] = t.Roman()
		} else {
			parts[i] = t.GetSurface()
		}
	}
	return parts
}

func tokenizedParts(tokens []AnyToken) []string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = t.GetSurface()
	}
	return parts
}

// roman constructs the romanized string intelligently using the provided spacing rule.
func defaultRoman(tokens []AnyToken) string {
	spacingRule := defaultSpacingRule
	var builder strings.Builder
	var prev string

	for i, token := range tokens {
		var text string
		// Use token.Roman() if available; otherwise, use token.GetSurface().
		if r := token.Roman(); r != "" {
			text = r
		} else {
			text = token.GetSurface()
		}

		if i > 0 && spacingRule(prev, text) {
			builder.WriteRune(' ')
		}
		builder.WriteString(text)
		prev = text
	}
	return builder.String()
}

// defaultTokenized constructs the tokenized string intelligently using the provided spacing rule.
func defaultTokenized(tokens []AnyToken) string {
	spacingRule := defaultSpacingRule
	var builder strings.Builder
	var prev string

	for i, token := range tokens {
		text := token.GetSurface()
		if i > 0 && spacingRule(prev, text) {
			builder.WriteRune(' ')
		}
		builder.WriteString(text)
		prev = text
	}
	return builder.String()
}


// SpacingRule defines a function signature for deciding if a space is needed between tokens.
type SpacingRule func(prev, current string) bool

// defaultSpacingRule is our configurable rule for whether to insert a space between two tokens.
func defaultSpacingRule(prev, current string) bool {
	// If either token is empty, no space is needed
	if prev == "" || current == "" {
		return false
	}

	prevRunes := []rune(prev)
	currRunes := []rune(current)
	
	if len(prevRunes) == 0 || len(currRunes) == 0 {
		return false
	}

	lastPrev := prevRunes[len(prevRunes)-1]
	firstCurr := currRunes[0]

	// 1. Handle punctuation that should not have spacing
	
	// 1.1 No space before closing punctuation or punctuation that follows text
	switch firstCurr {
	case '.', ',', '!', '?', ':', ';', ')', ']', '}', '»', '・', '…', '"', '\'', '」',
		 '】', '々', '～', '。', '、', '：', '；', '，', '．', '！', '？', '）', '］', '｝', '』', '》', '〉':
		return false
	}

	// 1.2 No space after opening punctuation or punctuation that precedes text
	switch lastPrev {
	case '(', '[', '{', '«', '"', '\'', '「', '【', '（', '［', '『', '《', '〈':
		return false
	}
	
	// 2. Detect Unicode character categories
	
	// 2.1 No space between punctuation and text
	if unicode.IsPunct(lastPrev) && !unicode.IsPunct(firstCurr) {
		return false
	}
	if !unicode.IsPunct(lastPrev) && unicode.IsPunct(firstCurr) {
		return false
	}
	
	// 2.2 No space between consecutive punctuation marks
	if unicode.IsPunct(lastPrev) && unicode.IsPunct(firstCurr) {
		return false
	}
	
	// 3. Language-specific handling for scripts that don't use spaces
	
	// 3.1 East Asian scripts (Chinese, Japanese, Korean)
	if unicode.Is(unicode.Han, lastPrev) || unicode.Is(unicode.Hiragana, lastPrev) || 
	   unicode.Is(unicode.Katakana, lastPrev) || unicode.Is(unicode.Hangul, lastPrev) {
		if unicode.Is(unicode.Han, firstCurr) || unicode.Is(unicode.Hiragana, firstCurr) || 
		   unicode.Is(unicode.Katakana, firstCurr) || unicode.Is(unicode.Hangul, firstCurr) {
			// Force space between consecutive characters of these scripts = tokenization
			return true
		}
	}
	
	// 3.2 No spaces between Latin characters and certain scripts
	//isLastLatinAlpha := unicode.Is(unicode.Latin, lastPrev) && unicode.IsLetter(lastPrev)
	//isFirstLatinAlpha := unicode.Is(unicode.Latin, firstCurr) && unicode.IsLetter(firstCurr)
	
	// 3.3 Thai script doesn't use spaces between words
	if unicode.Is(unicode.Thai, lastPrev) && unicode.Is(unicode.Thai, firstCurr) {
		// Force space between consecutive characters of these scripts = tokenization
		return true
	}
	
	// 3.4 No spaces in Indic scripts between characters of the same script
	isLastIndic := unicode.Is(unicode.Devanagari, lastPrev) || unicode.Is(unicode.Bengali, lastPrev) || 
	              unicode.Is(unicode.Gurmukhi, lastPrev) || unicode.Is(unicode.Gujarati, lastPrev) || 
	              unicode.Is(unicode.Tamil, lastPrev) || unicode.Is(unicode.Telugu, lastPrev) || 
	              unicode.Is(unicode.Kannada, lastPrev) || unicode.Is(unicode.Malayalam, lastPrev) || 
	              unicode.Is(unicode.Sinhala, lastPrev)
	              
	isFirstIndic := unicode.Is(unicode.Devanagari, firstCurr) || unicode.Is(unicode.Bengali, firstCurr) || 
	               unicode.Is(unicode.Gurmukhi, firstCurr) || unicode.Is(unicode.Gujarati, firstCurr) || 
	               unicode.Is(unicode.Tamil, firstCurr) || unicode.Is(unicode.Telugu, firstCurr) || 
	               unicode.Is(unicode.Kannada, firstCurr) || unicode.Is(unicode.Malayalam, firstCurr) || 
	               unicode.Is(unicode.Sinhala, firstCurr)
	               
	if isLastIndic && isFirstIndic {
		return false
	}
	
	// 4. Number and symbol handling
	
	// 4.1 No space between numbers and certain symbols
	if unicode.IsDigit(lastPrev) && (firstCurr == '.' || firstCurr == ',' || firstCurr == '%' || firstCurr == '°' || 
	   firstCurr == ':' || firstCurr == '-' || firstCurr == '/' || firstCurr == '×' || firstCurr == '⁄') {
		return false
	}
	
	// 4.2 No space between certain symbols and numbers
	if (lastPrev == '+' || lastPrev == '-' || lastPrev == '±' || lastPrev == '=' || lastPrev == '<' || lastPrev == '>' || 
	    lastPrev == '~' || lastPrev == '$' || lastPrev == '€' || lastPrev == '£' || lastPrev == '¥' || 
	    lastPrev == '₹' || lastPrev == '₽' || lastPrev == '¢' || lastPrev == '/' || lastPrev == ':' || 
	    lastPrev == '#' || lastPrev == '№') && unicode.IsDigit(firstCurr) {
		return false
	}
	
	// 4.3 No space between consecutive numbers
	if unicode.IsDigit(lastPrev) && unicode.IsDigit(firstCurr) {
		return false
	}
	
	// 5. Special cases
	
	// 5.1 No space in contractions with apostrophes
	if lastPrev == '\'' || firstCurr == '\'' {
		return false
	}
	
	// 5.2 No space in hyphenated words
	if lastPrev == '-' || firstCurr == '-' {
		return false
	}
	
	// 5.3 No space for Arabic/Persian scripts connecting characters
	if unicode.Is(unicode.Arabic, lastPrev) && unicode.Is(unicode.Arabic, firstCurr) {
		return false
	}
	
	// 5.4 No space for Hebrew characters
	if unicode.Is(unicode.Hebrew, lastPrev) && unicode.Is(unicode.Hebrew, firstCurr) {
		return false
	}
	
	// 5.5 No space for Cyrillic special cases
	if unicode.Is(unicode.Cyrillic, lastPrev) && lastPrev == 'ь' && unicode.Is(unicode.Cyrillic, firstCurr) {
		return false
	}
	
	// 6. By default, insert a space between lexical tokens
	return true
}


func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

