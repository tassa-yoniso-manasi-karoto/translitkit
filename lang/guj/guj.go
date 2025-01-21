
package guj

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const (
	ScriptGujarati = "Gujr" // Gujarati script
	ScriptLatin    = "Latn" // Romanized/Latin script
)

// Tkn extends the common Token with Gujarati-specific features
type Tkn struct {
	common.Tkn

	// Phonological features specific to Gujarati
	Phonology struct {
		HasNukta       bool   // Whether the character has a nukta
		HasAnusvara    bool   // Presence of anusvara (અં)
		HasVisarga     bool   // Presence of visarga (અઃ)
		HasAspiration  bool   // Whether consonants are aspirated
		HasMurmur      bool   // Presence of breathy voice/murmur
		SchwaDeleted   bool   // Whether schwa deletion is applied
		IsMulakshara   bool   // Whether it's a basic consonant without vowel
	}

	// Morphological features specific to Gujarati
	VerbStructure struct {
		Root          string   // Verb root
		IsTransitive  bool     // Transitivity
		IsCausative   bool     // Causative form
		IsPassive     bool     // Passive voice
		Aspect       string   // Imperfective, perfective, etc.
		Tense        string   // Present, past, future
		Person       int      // 1st, 2nd, or 3rd person
		Number       string   // Singular or plural
		Gender       string   // Masculine, feminine, neuter
	}

	// Nominal features
	NounProperties struct {
		Gender       string   // Grammatical gender
		Number       string   // Grammatical number
		Case         string   // Grammatical case
		IsAnimated   bool     // Animated vs non-animated distinction
		IsDefinite   bool     // Definiteness
		HasHonorific bool     // Honorific usage
	}

	// Compound word analysis
	CompoundType    string    // Type of compound word
	CompoundParts   []string  // Components of compound words
	SandhiRules     []string  // Applied sandhi rules

	// Script and transliteration features
	GujaratiForm    string    // Original form in Gujarati script
	HasDiacritic    bool      // Presence of diacritical marks
	DiacriticTypes  []string  // Types of diacritics used

	// Register and style
	Register struct {
		IsFormal     bool     // Formal vs informal usage
		IsTatsama    bool     // Sanskrit-derived word
		IsDesaja     bool     // Native Gujarati word
		IsPersian    bool     // Persian/Arabic loanword
		IsEnglish    bool     // English loanword
		Style        string   // Literary, colloquial, technical, etc.
	}

	// Dialect and variation
	Dialect struct {
		Region       string   // Geographic region
		SubDialect   string   // Specific sub-dialect
		IsSurati     bool     // Surat dialect
		IsKathiyawadi bool    // Kathiyawad dialect
		IsPatani     bool     // Patan dialect
		Features     []string // Specific dialectal features
	}

	// Additional linguistic features
	ExtraFeatures struct {
		HasAgglutination bool     // Agglutinative features
		IsReduplicated   bool     // Reduplication
		IsEcho          bool     // Echo word formation
		IsOnomatopoeia  bool     // Sound-symbolic word
	}
}

// NewToken creates a new Gujarati token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			IsToken:  true,
			Language: Lang,
			Script:   ScriptGujarati,
		},
	}
}

// IsTrueTatsama checks if the word is a pure Sanskrit loanword
func (t *Tkn) IsTrueTatsama() bool {
	return t.Register.IsTatsama && !t.Register.IsDesaja
}

// GetDialectName returns the full dialect classification
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Dialect.IsSurati:
		return "Surati Gujarati"
	case t.Dialect.IsKathiyawadi:
		return "Kathiyawadi Gujarati"
	case t.Dialect.IsPatani:
		return "Patani Gujarati"
	default:
		return "Standard Gujarati"
	}
}

// HasAgglutinativeFeatures checks for agglutinative morphology
func (t *Tkn) HasAgglutinativeFeatures() bool {
	return t.ExtraFeatures.HasAgglutination
}

// GetGrammaticalInfo returns complete grammatical information for nouns
func (t *Tkn) GetGrammaticalInfo() string {
	return t.NounProperties.Gender + " " + 
		t.NounProperties.Number + " " + 
		t.NounProperties.Case
}

// IsVerbCompound checks if the token is part of a compound verb
func (t *Tkn) IsVerbCompound() bool {
	return t.CompoundType == "verb"
}
