package mar

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const (
	ScriptDevanagari = "Deva" // Devanagari script
	ScriptLatin     = "Latn" // Romanized/Latin script
	ScriptModi      = "Modi" // Historical Modi script
)

// Tkn extends the common Token with Marathi-specific features
type Tkn struct {
	common.Tkn

	// Phonological features
	Phonology struct {
		HasAnusvara   bool     // Presence of anusvara (ं)
		HasVisarga    bool     // Presence of visarga (ः)
		HasChandrabindu bool   // Presence of chandrabindu (ॅ)
		IsSibilant    bool     // Whether the token contains sibilants
		HasNuktaSign  bool     // Modified consonant with nukta
		SchwaRetention bool    // Whether schwa is retained
	}

	// Morphological features
	VerbStructure struct {
		Root         string    // Verb root
		Transitivity string    // Transitive, intransitive, ditransitive
		Causative    string    // Single or double causative
		Aspect       string    // Perfect, imperfect, progressive
		Tense        string    // Present, past, future
		Person       int       // 1st, 2nd, or 3rd person
		Gender       string    // Masculine, feminine, neuter
		Number       string    // Singular, dual, plural
		IsErgativity bool      // Ergative construction
		HasConcord   bool      // Subject-verb agreement
	}

	// Nominal features
	NominalProperties struct {
		Gender        string   // Grammatical gender
		Number        string   // Grammatical number
		Case          string   // Direct, oblique, etc.
		IsAnimated    bool     // Animacy distinction
		Declension    string   // Declension pattern
		HasPostposition bool   // Associated postposition
	}

	// Derivational morphology
	Derivation struct {
		IsDiminutive  bool     // Diminutive form
		IsHonorific   bool     // Honorific form
		IsVerbalNoun  bool     // Derived from verb
		IsCausative   bool     // Causative derivation
		Suffixes      []string // Derivational suffixes
	}

	// Regional and dialectal variation
	Dialect struct {
		Region        string   // Geographic region
		IsAhirani     bool     // Ahirani dialect
		IsVarhadi     bool     // Varhadi dialect
		IsKhandeshi   bool     // Khandeshi dialect
		IsDangi       bool     // Dangi dialect
		Features      []string // Dialect-specific features
	}

	// Register and style
	Style struct {
		IsFormal      bool     // Formal register
		IsTatsama     bool     // Sanskrit-derived
		IsDesaja      bool     // Native Marathi
		IsTadbhava    bool     // Sanskrit-derived with changes
		IsPersoArabic bool     // Persian/Arabic loanword
		Register      string   // Formal, informal, literary
	}

	// Historical forms
	Historical struct {
		HasModiForm   bool     // Available in Modi script
		OldMarathi    string   // Old Marathi form if known
		Etymology     string   // Etymology information
	}

	// Additional linguistic features
	Extra struct {
		IsReduplicated bool     // Reduplication
		IsCompound     bool     // Compound formation
		HasUpapadVibhakti bool // Special case marking
		IsPerfective   bool     // Perfective aspect marking
	}
}

// NewToken creates a new Marathi token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:  surface,
			Language: Lang,
			Script:   ScriptDevanagari,
		},
	}
}

// IsErgative checks if the verb form uses ergative construction
func (t *Tkn) IsErgative() bool {
	return t.VerbStructure.IsErgativity
}

// GetDialectName returns the complete dialect classification
func (t *Tkn) GetDialectName() string {
	switch {
	case t.Dialect.IsAhirani:
		return "Ahirani Marathi"
	case t.Dialect.IsVarhadi:
		return "Varhadi Marathi"
	case t.Dialect.IsKhandeshi:
		return "Khandeshi Marathi"
	case t.Dialect.IsDangi:
		return "Dangi Marathi"
	default:
		return "Standard Marathi"
	}
}

// HasPostposition checks if the token has an associated postposition
func (t *Tkn) HasPostposition() bool {
	return t.NominalProperties.HasPostposition
}

// GetVerbAgreement returns complete agreement features
func (t *Tkn) GetVerbAgreement() string {
	return t.VerbStructure.Gender + " " + 
		t.VerbStructure.Number + " " + 
		"Person-" + string(t.VerbStructure.Person)
}

// GetMorphologicalType returns the word's morphological classification
func (t *Tkn) GetMorphologicalType() string {
	switch {
	case t.Style.IsTatsama:
		return "Tatsama"
	case t.Style.IsDesaja:
		return "Desaja"
	case t.Style.IsTadbhava:
		return "Tadbhava"
	case t.Style.IsPersoArabic:
		return "Perso-Arabic"
	default:
		return "Unclassified"
	}
}