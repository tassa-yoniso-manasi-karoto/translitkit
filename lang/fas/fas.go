
package fas

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)


const (
	ScriptPersoArabic = "Arab" // Persian-modified Arabic script
	ScriptLatin      = "Latn" // Romanized/Latin script
)

// Tkn extends the common Token with Persian-specific features
type Tkn struct {
	common.Tkn

	// Persian-specific orthographic features
	HasTashdid    bool   // Gemination marker (شدّت)
	HasTanvin     bool   // Nunation marker
	IzafetForm    string // Ezāfe construction form
	
	// Morphological features specific to Persian
	VerbStructure struct {
		Root        string   // Trilateral/quadrilateral root
		Prefix      string   // Verbal prefixes (می, ب, etc.)
		IsCausative bool     // Causative form
		IsPassive   bool     // Passive construction
		Aspect      string   // Perfect, imperfect, etc.
		Mood        string   // Indicative, subjunctive, imperative
		Tense       string   // Present, past, future
		Person      int      // 1st, 2nd, or 3rd person
		Number      string   // Singular or plural
	}

	// Nominal features
	NounProperties struct {
		IsAnimate    bool     // Animacy distinction
		IsCount      bool     // Count vs. mass noun
		HasDefinite  bool     // Definite marking (را)
		IsPossessed  bool     // Has possessive suffix
		IsVocative   bool     // Vocative case (ای)
	}

	// Compound word analysis
	CompoundType   string    // Type of compound word
	LightVerbBase  string    // For compound verbs, the nominal component
	LightVerb      string    // For compound verbs, the light verb component

	// Script and orthography
	PersoArabicForm string   // Original form in Persian script
	Finalized      string   // Form with proper character joining
	DiacriticsForm string   // Full diacritical marking
	
	// Register and etymology
	Register struct {
		IsLiterary    bool   // Literary/poetic usage
		IsColloquial  bool   // Colloquial usage
		IsArabic      bool   // Arabic loanword
		IsPersian     bool   // Native Persian word
		IsFormal      bool   // Formal register
		Register      string // Overall register classification
	}

	// Dialect and variation
	Dialect struct {
		Region     string   // Geographic dialect region
		IsIranian  bool     // Iranian Persian
		IsDari     bool     // Afghan Persian/Dari
		IsTajik    bool     // Tajik Persian
		Features   []string // Specific dialectal features
	}
}

// NewToken creates a new Persian token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			IsToken:  true,
			Language: Lang,
			Script:   ScriptPersoArabic,
		},
	}
}

// IsLightVerb checks if the token is part of a light verb construction
func (t *Tkn) IsLightVerb() bool {
	return t.LightVerb != "" && t.LightVerbBase != ""
}

// GetDialectName returns the full name of the dialect variety
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Dialect.IsIranian:
		return "Iranian Persian"
	case t.Dialect.IsDari:
		return "Dari"
	case t.Dialect.IsTajik:
		return "Tajik"
	default:
		return "Standard Persian"
	}
}

// HasEzafe checks if the token has an ezāfe construction
func (t *Tkn) HasEzafe() bool {
	return t.IzafetForm != ""
}

// GetVerbTense returns the complete tense-aspect-mood combination
func (t *Tkn) GetVerbTense() string {
	if t.VerbStructure.Tense == "" {
		return ""
	}
	return t.VerbStructure.Mood + " " + t.VerbStructure.Aspect + " " + t.VerbStructure.Tense
}