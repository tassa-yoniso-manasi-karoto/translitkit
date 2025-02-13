
package ben

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Script constants for Bengali text
const (
	ScriptBengali = "Beng"  // Bengali/Bangla script
	ScriptLatin   = "Latn"  // Romanized/Latin script
)

// BengaliTkn extends the common Token with Bengali-specific features
type Tkn struct {
	common.Tkn

	// Bengali-specific phonological features
	HasNasalization bool	// Whether the token has nasalization
	HasAspiration  bool	// Whether consonants are aspirated
	
	// Morphological features specific to Bengali
	VerbForm struct {
		IsShadhuBhasha bool	// Classical/literary Bengali form
		IsCholit	   bool	// Colloquial/modern Bengali form
		Honorific	 string   // Honorific level (intimate/familiar/polite/formal)
		Aspect		string   // Progressive, perfective, etc.
		Tense		 string   // Present, past, future
		Person		int	  // 1st, 2nd, or 3rd person
	}
	
	// Nominal features
	NounClass struct {
		IsAnimated	bool	 // Animated vs non-animated distinction
		IsDefinite	bool	 // Definiteness marking
		ClassifierUsed string  // Associated measure word/classifier if applicable
	}
	
	// Compound word analysis specific to Bengali
	CompoundType string		// Type of compound (tatpurusha, dvandva, etc.)
	SandhiRules  []string	 // Applied sandhi (phonological combination) rules
	
	// Script and transliteration
	BengaliForm	string	 // Original form in Bengali script
	IASTPrecise	string	 // Scientific transliteration (IAST)
	Isdialectal	bool	   // Whether this is a dialectal variant
	DialectRegion  string	 // Geographic region for dialectal forms
	
	// Register and style
	Register struct {
		IsFormal	  bool	// Formal vs informal usage
		IsPoetic	  bool	// Poetic/literary usage
		IsTatsama	 bool	// Sanskrit-derived word
		IsDeshi	   bool	// Native Bengali word
		IsPersoArabic bool	// Persian/Arabic loanword
	}
}

// NewToken creates a new Bengali token with default values
func NewToken(surface string) *Tkn {
	return &Tkn{
		Tkn: common.Tkn{
			Surface:   surface,
			Language:  Lang,
			Script:	ScriptBengali,
		},
	}
}

// ToShadhuBhasha converts the token to classical Bengali form if possible
// func (t *Tkn) ToShadhuBhasha() string {
// 	// Implementation would go here
// 	return ""  // Placeholder
// }

// ToCholit converts the token to colloquial Bengali form if possible
// func (t *Tkn) ToCholit() string {
// 	// Implementation would go here
// 	return ""  // Placeholder
// }

// GetHonorificLevel returns the full description of the honorific level
func (t *Tkn) GetHonorificLevel() string {
	switch t.VerbForm.Honorific {
	case "int":
		return "Intimate"
	case "fam":
		return "Familiar"
	case "pol":
		return "Polite"
	case "form":
		return "Formal"
	default:
		return "Unknown"
	}
}