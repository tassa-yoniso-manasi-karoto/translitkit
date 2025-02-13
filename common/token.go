package common

import (
	"strings"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
)

func serialize(input string, max int) (AnyTokenSliceWrapper, error) {
	chunks, err := chunkify(input, max)
	return &TknSliceWrapper{Raw: chunks}, err
}


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
	return roman(tokens.Slice)
}
func (tokens TknSliceWrapper) RomanParts() []string {
	return romanParts(tokens.Slice)
}

func (tokens TknSliceWrapper) Tokenized() string {
	return tokenized(tokens.Slice)
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




// due to common.Tkn.embedding and methods inheritance, interfaces are overkill here
func roman(tokens []AnyToken) string {
	return strings.Join(romanParts(tokens), " ")
}

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

func tokenized(tokens []AnyToken) string {
	return strings.Join(tokenizedParts(tokens), " ")
}


func tokenizedParts(tokens []AnyToken) []string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = t.GetSurface()
	}
	return parts
}




func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

