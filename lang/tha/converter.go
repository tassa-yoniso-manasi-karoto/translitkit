package tha

import (
	"unicode"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// convertToThaiToken converts a common.Tkn to tha.Tkn
// This maps only the essential fields from go-pythainlp results
func convertToThaiToken(token *common.Tkn) *Tkn {
	thaiToken := &Tkn{
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
	
	// Override IsLexical to properly detect non-lexical tokens
	// PyThaiNLP includes punctuation as tokens, but they should not be lexical
	thaiToken.IsLexical = isLexicalContent(token.Surface)
	
	return thaiToken
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

// isLexicalContent determines if a token contains actual lexical content
// (not just punctuation or symbols)
func isLexicalContent(text string) bool {
	if text == "" {
		return false
	}
	
	// Check if it contains Thai characters (0x0E00 to 0x0E7F)
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	
	// Check if it contains letters or digits (any script)
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	
	// If it's only punctuation, symbols, or spaces, it's not lexical
	return false
}