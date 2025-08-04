package tha

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// convertToThaiToken converts a common.Tkn to tha.Tkn
// This maps only the essential fields from go-pythainlp results
func convertToThaiToken(token *common.Tkn) *Tkn {
	return &Tkn{
		Tkn: *token, // Embed the common token
		
		// Thai-specific fields are left empty for now
		// These could be populated in future versions when
		// go-pythainlp provides more linguistic annotations
		
		// Example of fields that could be populated later:
		// - InitialConsonant, Vowel, FinalConsonant (from syllable analysis)
		// - ConsonantClass (high, mid, low)
		// - SyllableType
		// - IsRoyal (from dictionary lookup)
		// - RegisterLevel (formal/informal detection)
		// - Etymology (Thai, Pali, Sanskrit detection)
	}
}

// convertPyThaiNLPToken could be used in the future to convert
// go-pythainlp's Token struct to tha.Tkn if go-pythainlp
// provides richer token structures
func convertPyThaiNLPToken(pyToken interface{}) *Tkn {
	// For now, this is a placeholder for future enhancement
	// when go-pythainlp might provide its own token types
	
	tkn := &Tkn{
		Tkn: common.Tkn{
			IsLexical: true,
		},
	}
	
	// Future: Extract additional fields from pyToken
	
	return tkn
}