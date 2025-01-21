package rus

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Russian-specific features
type Tkn struct {
	common.Tkn

	// Morphological features
	Case          GramCase        // 6 cases: Nominative, Genitive, Dative, Accusative, Instrumental, Prepositional
	Number        Number          // Singular, Plural
	Gender        Gender          // Masculine, Feminine, Neuter
	Animacy       Animacy        // Animate vs Inanimate (affects accusative case)
	
	// Verb-specific features
	Aspect        Aspect         // Perfective vs Imperfective
	Tense         Tense          // Past, Present, Future
	Person        Person         // 1st, 2nd, 3rd person
	VerbForm      VerbForm       // Infinitive, Finite, Participle, Gerund
	Voice         Voice          // Active, Passive
	Mood         Mood           // Indicative, Imperative, Subjunctive/Conditional
	
	// Adjective features
	AdjForm       AdjForm        // Long form vs Short form
	Degree        Degree         // Positive, Comparative, Superlative
	
	// Stress and phonetic features
	StressPos     int            // Position of stressed syllable (important for correct pronunciation)
	HasYo         bool           // Contains ё (often written as е but pronounced differently)
	YoPositions   []int          // Positions of ё in the token
	
	// Word formation
	Prefix        string         // Derivational prefix if any
	Root          string         // Root morpheme
	Suffix        string         // Derivational/inflectional suffix
	Ending        string         // Grammatical ending
	
	// Spelling rules
	VowelAlternation bool        // Whether token exhibits vowel alternation
	Palatalization   bool        // Whether consonant palatalization occurs
}

// Enums for Russian linguistic features
type GramCase string
const (
	Nominative    GramCase = "nom"
	Genitive      GramCase = "gen"
	Dative        GramCase = "dat"
	Accusative    GramCase = "acc"
	Instrumental  GramCase = "ins"
	Prepositional GramCase = "prep"
)

type Number string
const (
	Singular Number = "sg"
	Plural   Number = "pl"
)

type Gender string
const (
	Masculine Gender = "m"
	Feminine  Gender = "f"
	Neuter    Gender = "n"
)

type Animacy string
const (
	Animate   Animacy = "anim"
	Inanimate Animacy = "inan"
)

type Aspect string
const (
	Perfective   Aspect = "perf"
	Imperfective Aspect = "imperf"
)

type Tense string
const (
	Past    Tense = "past"
	Present Tense = "pres"
	Future  Tense = "fut"
)

type Person string
const (
	First  Person = "1"
	Second Person = "2"
	Third  Person = "3"
)

type VerbForm string
const (
	Infinitive VerbForm = "inf"
	Finite     VerbForm = "fin"
	Participle VerbForm = "part"
	Gerund     VerbForm = "ger"
)

type Voice string
const (
	Active  Voice = "act"
	Passive Voice = "pass"
)

type Mood string
const (
	Indicative  Mood = "ind"
	Imperative  Mood = "imp"
	Conditional Mood = "cond"
)

type AdjForm string
const (
	Long  AdjForm = "long"
	Short AdjForm = "short"
)

type Degree string
const (
	Positive    Degree = "pos"
	Comparative Degree = "comp"
	Superlative Degree = "sup"
)

// Helper methods

// IsVerb returns true if the token is a verb
func (t *Tkn) IsVerb() bool {
	return t.PartOfSpeech == "verb"
}

// IsAdjective returns true if the token is an adjective
func (t *Tkn) IsAdjective() bool {
	return t.PartOfSpeech == "adj"
}

// HasStress returns true if stress position is marked
func (t *Tkn) HasStress() bool {
	return t.StressPos >= 0
}

// NeedsYoResolution returns true if token potentially contains ё written as е
func (t *Tkn) NeedsYoResolution() bool {
	return t.HasYo && len(t.YoPositions) > 0
}
