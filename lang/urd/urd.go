package urd

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)


const (
	ScriptNastaliq = "Aran" // Nastaliq style Arabic script
	ScriptNaskh    = "Arab" // Naskh style Arabic script
	ScriptLatin    = "Latn" // Romanized/Latin script
	ScriptDeva     = "Deva" // Devanagari script (for Hindustani forms)
)

// Tkn extends the common Token with Urdu-specific features
type Tkn struct {
	common.Tkn

	// Orthographic features
	OrthographicForm struct {
		HasTashkeel    bool      // Diacritical marks present
		HasTanween     bool      // Nunation marks
		HasTashdeed    bool      // Gemination marker
		HasHamza       bool      // Hamza marker
		HasMadd        bool      // Madd sign
		FinalForm      string    // Final form of the letter
		MedialForm     string    // Medial form of the letter
		InitialForm    string    // Initial form of the letter
		IsolatedForm   string    // Isolated form of the letter
	}

	// Morphological features
	VerbStructure struct {
		Root          string    // Trilateral/quadrilateral root
		Pattern       string    // Verb pattern/form (باب)
		Aspect        string    // Perfect, imperfect, etc.
		Tense         string    // Present, past, future
		Person        int       // 1st, 2nd, or 3rd person
		Number        string    // Singular or plural
		Gender        string    // Masculine or feminine
		Mood          string    // Indicative, subjunctive, etc.
		Voice         string    // Active or passive
		IsLight       bool      // Light verb construction
		LightVerb     string    // Associated light verb if any
	}

	// Nominal features
	NominalProperties struct {
		Gender        string    // Grammatical gender
		Number        string    // Grammatical number
		Case          string    // Nominative, oblique, etc.
		IsEzafe       bool      // Ezafe construction
		HasIzafat     bool      // Izafat marker
		Definiteness  string    // Definite, indefinite
		HasTanwin     bool      // Nunation marking
	}

	// Etymology and register
	Etymology struct {
		Origin        string    // Language of origin
		IsPersian     bool      // Persian origin
		IsArabic      bool      // Arabic origin
		IsTurkic      bool      // Turkic origin
		IsIndic       bool      // Indic origin
		IsEnglish     bool      // English loanword
		RegisterLevel string    // High/formal vs. common usage
	}

	// Style and register
	Style struct {
		IsFormal      bool      // Formal usage
		IsPoetic      bool      // Poetic/literary usage
		IsColloquial  bool      // Colloquial usage
		Register      string    // Overall register
		Formality     string    // Level of formality
	}

	// Regional variation
	Dialect struct {
		Region        string    // Geographic region
		IsDeccani     bool      // Deccani Urdu
		IsLucknowi    bool      // Lucknow dialect
		IsDelhiStyle  bool      // Delhi style
		IsRekhta      bool      // Rekhta style
		Features      []string  // Dialect-specific features
	}

	// Cross-script equivalence
	Equivalents struct {
		DevanagariForm string   // Devanagari representation
		RomanForm      string   // Roman Urdu form
		IAST          string   // Scientific transliteration
	}

	// Compound analysis
	CompoundStructure struct {
		IsCompound    bool      // Is it a compound word
		Components    []string  // Component parts
		Type          string   // Type of compound
		HeadPosition  string   // Position of head word
	}

	// Additional features
	Extra struct {
		HasHonorific  bool      // Honorific form
		IsPlusCentric bool      // Plus-centric form
		IsVocative    bool      // Vocative form
		HasEnclitic   bool      // Has enclitic particles
	}
}

// NewToken creates a new Urdu token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			Language: Lang,
			Script:   ScriptNastaliq,
		},
	}
}

// IsLightVerb checks if the token is part of a light verb construction
func (t *Tkn) IsLightVerb() bool {
	return t.VerbStructure.IsLight && t.VerbStructure.LightVerb != ""
}

// GetDialectName returns the complete dialect classification
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Dialect.IsDeccani:
		return "Deccani Urdu"
	case t.Dialect.IsLucknowi:
		return "Lucknowi Urdu"
	case t.Dialect.IsDelhiStyle:
		return "Delhi Urdu"
	case t.Dialect.IsRekhta:
		return "Rekhta"
	default:
		return "Standard Urdu"
	}
}

// HasIzafatConstruction checks if the token is part of an izafat construction
func (t *Tkn) HasIzafatConstruction() bool {
	return t.NominalProperties.IsEzafe || t.NominalProperties.HasIzafat
}

// GetEtymology returns the complete etymology classification
func (t *Tkn) GetEtymology() string {
	switch {
	case t.Etymology.IsArabic:
		return "Arabic origin"
	case t.Etymology.IsPersian:
		return "Persian origin"
	case t.Etymology.IsTurkic:
		return "Turkic origin"
	case t.Etymology.IsIndic:
		return "Indic origin"
	case t.Etymology.IsEnglish:
		return "English loanword"
	default:
		return "Unknown origin"
	}
}

// GetScriptForms returns all available script representations
func (t *Tkn) GetScriptForms() map[string]string {
	forms := make(map[string]string)
	if t.Surface != "" {
		forms["Nastaliq"] = t.Surface
	}
	if t.Equivalents.DevanagariForm != "" {
		forms["Devanagari"] = t.Equivalents.DevanagariForm
	}
	if t.Equivalents.RomanForm != "" {
		forms["Roman"] = t.Equivalents.RomanForm
	}
	return forms
}