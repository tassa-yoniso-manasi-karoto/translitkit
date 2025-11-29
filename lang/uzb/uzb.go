package uzb

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Uzbek-specific features
type Tkn struct {
	common.Tkn

	// Script-related features
	Script        ScriptType     // Latin, Cyrillic, or Arabic (historical)
	HasApostrophe bool           // For Latin script: contains ʻ (for ъ)
	
	// Case and possession
	Case          GramCase       // Nominative, Genitive, Dative, Accusative, Locative, Ablative
	Possession    Possession     // Possessive suffixes
	
	// Number
	Number        Number         // Singular, Plural
	IsCollective  bool          // Collective number (e.g., -lar with collective meaning)
	
	// Verb features
	VerbForm      VerbForm      // Finite, Infinitive, Participle, Converb
	Tense         Tense         // Present-Future, Past (Definite/Indefinite), Future
	Person        Person        // 1st, 2nd (Informal/Formal), 3rd
	Voice         Voice         // Active, Passive, Causative, Cooperative, Reflexive
	Mood          Mood          // Indicative, Imperative, Conditional, Optative
	Ability       bool          // Ability form (-a ol-)
	
	// Adjective features
	Degree        Degree        // Positive, Comparative, Superlative, Intensive
	
	// Derivation and word formation
	Compounds     []string      // Parts of compound words
	Reduplication bool         // Reduplicated form (e.g., katta-katta)
	
	// Vowel harmony
	VowelType     VowelType    // Front or Back vowels
	HasHarmony    bool         // Whether token follows vowel harmony
	
	// Morphological features
	Negation      bool         // Presence of negative suffix -ma
	Question      bool         // Presence of question particle -mi
	Emphasis      bool         // Emphatic particles (e.g., -ku, -da)
	Copula        bool         // Zero copula or expressed copula
}

// Enums for Uzbek linguistic features
type ScriptType string
const (
	Latin    ScriptType = "latin"
	Cyrillic ScriptType = "cyrillic"
	Arabic   ScriptType = "arabic"
)

type GramCase string
const (
	Nominative GramCase = "nom"
	Genitive   GramCase = "gen"
	Dative     GramCase = "dat"
	Accusative GramCase = "acc"
	Locative   GramCase = "loc"
	Ablative   GramCase = "abl"
)

type Possession string
const (
	PossNone   Possession = ""
	Poss1Sg    Possession = "1sg"
	Poss2SgInf Possession = "2sg.inf"  // Informal
	Poss2SgFrm Possession = "2sg.frm"  // Formal
	Poss3Sg    Possession = "3sg"
	Poss1Pl    Possession = "1pl"
	Poss2Pl    Possession = "2pl"
	Poss3Pl    Possession = "3pl"
)

type Number string
const (
	Singular Number = "sg"
	Plural   Number = "pl"
)

type VerbForm string
const (
	Finite     VerbForm = "fin"
	Infinitive VerbForm = "inf"
	Participle VerbForm = "part"
	Converb    VerbForm = "conv"
)

type Tense string
const (
	PresentFuture   Tense = "pres.fut"
	PastDefinite    Tense = "past.def"
	PastIndefinite  Tense = "past.indef"
	PastNarrative   Tense = "past.narr"
	FutureDefinite  Tense = "fut.def"
	FutureIndefinite Tense = "fut.indef"
)

type Person string
const (
	First      Person = "1"
	SecondInf  Person = "2.inf"
	SecondFrm  Person = "2.frm"
	Third      Person = "3"
)

type Voice string
const (
	Active      Voice = "act"
	Passive     Voice = "pass"
	Causative   Voice = "caus"
	Cooperative Voice = "coop"
	Reflexive   Voice = "refl"
)

type Mood string
const (
	Indicative  Mood = "ind"
	Imperative  Mood = "imp"
	Conditional Mood = "cond"
	Optative    Mood = "opt"
)

type Degree string
const (
	Positive   Degree = "pos"
	Comparative Degree = "comp"
	Superlative Degree = "sup"
	Intensive   Degree = "int"    // e.g., qip-qizil
)

type VowelType string
const (
	Front VowelType = "front"
	Back  VowelType = "back"
)

// Helper methods

// IsVerb returns true if the token is a verb
func (t *Tkn) IsVerb() bool {
	return t.PartOfSpeech == "verb"
}

// HasPossession returns true if the token has any possessive suffix
func (t *Tkn) HasPossession() bool {
	return t.Possession != PossNone
}

// IsCompound returns true if the token is a compound word
func (t *Tkn) IsCompound() bool {
	return len(t.Compounds) > 1
}

// NeedsScriptConversion returns true if token might need script conversion
func (t *Tkn) NeedsScriptConversion() bool {
	return t.Script != Latin // Assuming Latin is the target script
}
