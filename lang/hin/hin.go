
package hin

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Hindi-specific features
type Tkn struct {
	common.Tkn
	
	// Schwa deletion information
	HasSchwa      bool    // Whether token contains schwa sound
	SchwaPosition int     // Position of schwa in token
	SchwaDeleted  bool    // Whether schwa should be deleted in pronunciation
	
	// Sandhi (sound change) information
	HasSandhi     bool              // Whether token undergoes sandhi
	SandhiType    SandhiType        // Type of sandhi transformation
	SandhiParts   []string          // Original parts before sandhi
	
	// Gender grammatical features
	Gender        GrammaticalGender // Grammatical gender of the word
	
	// Number-specific features
	NumberMarker  NumberMarking     // Type of plural/singular marking
	
	// Case marking
	Case          CaseMarking       // Grammatical case
	PostPosition  string            // Associated postposition if any
	
	// Verb-specific features
	VerbForm      VerbForm         // Specific verb form/conjugation
	Aspect        AspectType       // Perfective, imperfective, etc.
	Tense         TenseType        // Present, past, future
	
	// Honorific level
	Honorific     HonorificLevel   // Formal, informal, familiar
	
	// Urdu cognate (if exists)
	UrduCognate   string           // Equivalent word in Urdu script
}

// Enums for Hindi linguistic features
type SandhiType string
const (
	NoSandhi    SandhiType = ""
	VowelSandhi SandhiType = "vowel"    // e.g., विद्या + आलय → विद्यालय
	VisargaSandhi SandhiType = "visarga" // e.g., नमः + करोति → नमस्करोति
)

type GrammaticalGender string
const (
	Masculine GrammaticalGender = "m"
	Feminine  GrammaticalGender = "f"
)

type NumberMarking string
const (
	Singular NumberMarking = "sg"
	Plural   NumberMarking = "pl"
)

type CaseMarking string
const (
	Direct    CaseMarking = "direct"
	Oblique   CaseMarking = "oblique"
	Vocative  CaseMarking = "vocative"
)

type VerbForm string
const (
	Stem      VerbForm = "stem"
	Infinitive VerbForm = "inf"
	Participle VerbForm = "part"
	Imperative VerbForm = "imp"
)

type AspectType string
const (
	Perfective    AspectType = "perf"
	Imperfective  AspectType = "imperf"
	Habitual      AspectType = "hab"
	Progressive   AspectType = "prog"
)

type TenseType string
const (
	Present TenseType = "pres"
	Past    TenseType = "past"
	Future  TenseType = "fut"
)

type HonorificLevel string
const (
	Informal HonorificLevel = "informal" // तू
	Familiar HonorificLevel = "familiar" // तुम
	Formal   HonorificLevel = "formal"   // आप
)

// Helper methods

// IsVerb returns true if the token is a verb
func (t *Tkn) IsVerb() bool {
	return t.PartOfSpeech == "verb"
}

// IsPronoun returns true if the token is a pronoun
func (t *Tkn) IsPronoun() bool {
	return t.PartOfSpeech == "pronoun"
}

// HasPostPosition returns true if the token has an associated postposition
func (t *Tkn) HasPostPosition() bool {
	return t.PostPosition != ""
}

// NeedsSchwaDelete determines if schwa deletion should be applied
func (t *Tkn) NeedsSchwaDelete() bool {
	return t.HasSchwa && !t.SchwaDeleted
}

// GetStemForm returns the stem form of a verb
func (t *Tkn) GetStemForm() string {
	if !t.IsVerb() {
		return ""
	}
	return t.Lemma
}
