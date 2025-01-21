package sin

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)


const (
	ScriptSinhala = "Sinh" // Sinhala script
	ScriptLatin   = "Latn" // Romanized/Latin script
)

// Tkn extends the common Token with Sinhala-specific features
type Tkn struct {
	common.Tkn

	// Phonological features
	Phonology struct {
		HasPrenasalization bool    // Prenasalized stops
		HasAspiration     bool    // Aspirated consonants
		HasHalKirima      bool    // Hal marker (්)
		IsPrenasalized    bool    // Prenasalized consonant
		HasRakaransaya    bool    // Special රැ form
		HasYansaya        bool    // Special ය form
		IsConjunct        bool    // Touching letters
	}

	// Diglossia features
	Diglossia struct {
		IsLiterary     bool      // Literary Sinhala
		IsSpoken       bool      // Spoken Sinhala
		RegisterLevel  string    // Formal, informal, etc.
		HasHonorific   bool      // Honorific forms
		Style          string    // Written vs. colloquial style
	}

	// Verb features specific to Sinhala
	VerbStructure struct {
		Root           string    // Verb root
		Volition       string    // Voluntary vs. involuntary action
		Voice          string    // Active, passive, causative
		Aspect         string    // Perfective, imperfective
		Tense          string    // Past, non-past
		Person         int       // 1st, 2nd, or 3rd person
		Number         string    // Singular or plural
		Definiteness   string    // Definite or indefinite
		IsInvolitional bool      // Involitional verb form
		HasCausative   bool      // Causative construction
	}

	// Nominal features
	NominalProperties struct {
		Case          string    // Nominative, accusative, etc.
		Number        string    // Singular, plural
		Definiteness  string    // Definite, indefinite
		Animacy       string    // Animate vs. inanimate
		HasHonorific  bool      // Honorific form
	}

	// Script-specific features
	ScriptFeatures struct {
		HasBindu       bool      // Bindu diacritic
		HasSagngnaka   bool      // Special න්‍ය form
		IsComposed     bool      // Composed character cluster
		Components     []string  // Component characters
	}

	// Word formation
	Formation struct {
		IsSanskritized bool      // Sanskrit-influenced form
		IsPali         bool      // Pali-derived word
		IsNativized    bool      // Nativized loanword
		IsCompound     bool      // Compound word
		Etymology      string    // Word origin
	}

	// Regional and social variation
	Variation struct {
		Region        string    // Geographic region
		IsColombo     bool      // Colombo dialect
		IsKandy       bool      // Kandy dialect
		IsUpcountry   bool      // Upcountry dialect
		IsSouthern    bool      // Southern dialect
		SocialClass   string    // Social class association
		CasteMarker   bool      // Whether indicates caste
	}

	// Additional features
	Extra struct {
		IsReduplicated  bool     // Reduplication
		HasClitics      bool     // Clitic particles
		IsFocusMarked   bool     // Focus marking
		HasTopicMarker  bool     // Topic marker
	}
}

// NewToken creates a new Sinhala token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			IsToken:  true,
			Language: Lang,
			Script:   ScriptSinhala,
		},
	}
}

// GetRegisterType returns the complete register classification
func (t *Tkn) GetRegisterType() string {
	if t.Diglossia.IsLiterary {
		return "Literary Sinhala - " + t.Diglossia.RegisterLevel
	}
	return "Spoken Sinhala - " + t.Diglossia.RegisterLevel
}

// GetDialectName returns the complete dialect classification
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Variation.IsColombo:
		return "Colombo Sinhala"
	case t.Variation.IsKandy:
		return "Kandyan Sinhala"
	case t.Variation.IsUpcountry:
		return "Upcountry Sinhala"
	case t.Variation.IsSouthern:
		return "Southern Sinhala"
	default:
		return "Standard Sinhala"
	}
}

// IsVoluntaryAction checks if the verb denotes a voluntary action
func (t *Tkn) IsVoluntaryAction() bool {
	return t.VerbStructure.Volition == "voluntary"
}

// GetMorphologicalType returns the word's morphological classification
func (t *Tkn) GetMorphologicalType() string {
	switch {
	case t.Formation.IsSanskritized:
		return "Sanskritized"
	case t.Formation.IsPali:
		return "Pali-derived"
	case t.Formation.IsNativized:
		return "Nativized"
	default:
		return "Native Sinhala"
	}
}

// HasSpecialCharacterForm checks for special character formations
func (t *Tkn) HasSpecialCharacterForm() bool {
	return t.Phonology.HasRakaransaya || 
		t.Phonology.HasYansaya || 
		t.ScriptFeatures.HasSagngnaka
}
