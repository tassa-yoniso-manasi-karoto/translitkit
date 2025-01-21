package pan

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const (
	ScriptGurmukhi  = "Guru" // Gurmukhi script (Eastern Punjabi)
	ScriptShahmukhi = "Arab" // Shahmukhi script (Western Punjabi)
	ScriptLatin     = "Latn" // Romanized/Latin script
)

// Tkn extends the common Token with Punjabi-specific features
type Tkn struct {
	common.Tkn

	// Tonal features specific to Punjabi
	Tonality struct {
		HasTone      bool    // Whether the word has tonal features
		ToneType     string  // High, low, or level tone
		IsTonogenic  bool    // Whether derived from loss of voiced aspirates
		TonePosition int     // Position of tone in word
	}

	// Script-specific features
	Script struct {
		HasAddak      bool    // Gemination marker in Gurmukhi (ੱ)
		HasTippi      bool    // Nasalization marker in Gurmukhi (ੰ)
		HasBindi      bool    // Nasalization marker in Gurmukhi (ਂ)
		HasPairin     bool    // Sub-consonant marker in Gurmukhi
		HasSihari     bool    // Vowel marker in Gurmukhi
		IsGeminative  bool    // Whether consonant is doubled
	}

	// Morphological features
	VerbStructure struct {
		Root         string  // Verb root
		IsTransitive bool    // Transitivity
		Aspect      string  // Perfect, imperfect, progressive, etc.
		Tense       string  // Present, past, future
		Person      int     // 1st, 2nd, or 3rd person
		Number      string  // Singular or plural
		Gender      string  // Masculine or feminine
		IsPeriphrasticTense bool // Complex tense formation
		HasErgative bool    // Ergative construction
	}

	// Nominal features
	NounProperties struct {
		Gender      string  // Grammatical gender
		Number      string  // Grammatical number
		Case        string  // Direct, oblique, vocative, etc.
		IsAnimated  bool    // Animacy distinction
		HasVocative bool    // Vocative case marking
		IsRespect   bool    // Honorific/respect form
	}

	// Dialect and variation
	Dialect struct {
		Region      string   // Geographic region
		IsMajhi     bool     // Majhi dialect (standard)
		IsMalwai    bool     // Malwai dialect
		IsDoabi     bool     // Doabi dialect
		IsPowadhi   bool     // Powadhi dialect
		IsPothwari  bool     // Pothwari dialect
		Features    []string // Dialect-specific features
	}

	// Register and etymology
	Style struct {
		IsFormal    bool     // Formal vs informal
		IsTatsama   bool     // Sanskrit-derived
		IsPersoArabic bool   // Persian/Arabic loanword
		IsEnglish   bool     // English loanword
		Register    string   // Overall register classification
	}

	// Compound word analysis
	Compound struct {
		IsCompound   bool     // Whether it's a compound word
		Parts       []string  // Component parts
		Type        string   // Type of compound
		HasSandhi   bool     // Whether sandhi rules apply
	}

	// Additional linguistic features
	Extra struct {
		IsReduplicated bool    // Reduplication
		IsEcho        bool    // Echo word formation
		HasGender     bool    // Gender inflection
		IsMultiscript bool    // Available in multiple scripts
	}

	// Cross-script representations
	Forms struct {
		GurmukhiForm  string  // Form in Gurmukhi script
		ShahmukhiForm string  // Form in Shahmukhi script
		IASTPrecise   string  // Scientific transliteration
	}
}

// NewToken creates a new Punjabi token with default values
func NewToken(surface string, script string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			IsToken:  true,
			Language: Lang,
			Script:   script,
		},
	}
}

// HasTonalFeatures checks if the word exhibits tonal characteristics
func (t *Tkn) HasTonalFeatures() bool {
	return t.Tonality.HasTone
}

// GetDialectName returns the complete dialect classification
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Dialect.IsMajhi:
		return "Majhi Punjabi"
	case t.Dialect.IsMalwai:
		return "Malwai Punjabi"
	case t.Dialect.IsDoabi:
		return "Doabi Punjabi"
	case t.Dialect.IsPowadhi:
		return "Powadhi Punjabi"
	case t.Dialect.IsPothwari:
		return "Pothwari Punjabi"
	default:
		return "Standard Punjabi"
	}
}

// GetVerbAgreement returns the complete agreement features
func (t *Tkn) GetVerbAgreement() string {
	agreement := t.VerbStructure.Gender + " " +
		t.VerbStructure.Number + " " +
		"Person-" + string(t.VerbStructure.Person)
	if t.VerbStructure.HasErgative {
		agreement += " (Ergative)"
	}
	return agreement
}

// IsMultiScriptAvailable checks if the token has representations in both scripts
func (t *Tkn) IsMultiScriptAvailable() bool {
	return t.Forms.GurmukhiForm != "" && t.Forms.ShahmukhiForm != ""
}

// GetNominalization returns nominal form information if applicable
func (t *Tkn) GetNominalization() string {
	if t.NounProperties.Case == "" {
		return ""
	}
	return t.NounProperties.Gender + " " +
		t.NounProperties.Number + " " +
		t.NounProperties.Case
}