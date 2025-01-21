package tam

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Tamil-specific features
type Tkn struct {
	common.Tkn

	// Script features
	HasPulli      bool           // புள்ளி - Marks pure consonants
	HasAaytham    bool           // ஃ presence
	
	// Case system (வேற்றுமை)
	Case          GramCase       // 8 cases in Tamil
	
	// Gender-Number-Person features (பால்-எண்-இடம்)
	Gender        Gender         // Masculine, Feminine, Neuter (with rational/irrational distinction)
	Number        Number         // Singular, Plural (with inclusive/exclusive distinction)
	Person        Person         // 1st, 2nd, 3rd person
	Rationality   Rationality    // Rational vs Irrational distinction
	
	// Verb features
	VerbForm      VerbForm      // Finite, Infinite, Participle types
	Tense         Tense         // Past, Present, Future
	Voice         Voice         // Active, Passive, Middle, Causative
	Mood          Mood          // Indicative, Imperative, Optative, etc.
	Polarity      Polarity      // Positive, Negative
	//Aspect        Aspect        // Perfect, Progressive, Habitual
	
	// Sandhi (புணர்ச்சி) features
	SandhiType    SandhiType    // Type of morphophonemic change
	PreSandhi     string        // Form before sandhi
	PostSandhi    string        // Form after sandhi
	
	// Word formation
	Compounds     []string      // Parts of compound words
	IsCompound    bool          // Whether it's a compound word
	CompoundType  CompoundType  // Type of compound formation
	
	// Honorific features
	Honorific     HonorificLevel // Different levels of respect
	
	// Phonological features
	KurilNedil    []Length      // Short/Long vowel markers for each syllable
	
	// Literary classification
	LitCategory   LitCategory   // செந்தமிழ் (formal) vs கொடுந்தமிழ் (colloquial)
	RegisterLevel RegisterLevel // Formal, Informal, Literary, etc.
}

// Enums for Tamil linguistic features
type GramCase string
const (
	Nominative    GramCase = "nom"    // எழுவாய்
	Accusative    GramCase = "acc"    // செயப்படுபொருள்
	Instrumental  GramCase = "ins"    // கருவி
	Sociative    GramCase = "soc"    // உடனிகழ்ச்சி
	Dative       GramCase = "dat"    // கொடை
	Genitive     GramCase = "gen"    // உடைமை
	Locative     GramCase = "loc"    // இடம்
	Ablative     GramCase = "abl"    // நீங்கல்
)

type Gender string
const (
	MasculineRational   Gender = "m.rat"    // ஆண்பால் (உயர்திணை)
	FeminineRational    Gender = "f.rat"    // பெண்பால் (உயர்திணை)
	NeuterRational      Gender = "n.rat"    // ஒன்றன்பால் (உயர்திணை)
	NeuterIrrational    Gender = "n.irr"    // ஒன்றன்பால் (அஃறிணை)
)

type Number string
const (
	Singular        Number = "sg"
	PluralInclusive Number = "pl.incl"  // நாம்
	PluralExclusive Number = "pl.excl"  // நாங்கள்
)

type Person string
const (
	First  Person = "1"
	Second Person = "2"
	Third  Person = "3"
)

type Rationality string
const (
	Rational   Rationality = "உயர்திணை"
	Irrational Rationality = "அஃறிணை"
)

type VerbForm string
const (
	Finite           VerbForm = "fin"
	Infinitive       VerbForm = "inf"
	VerbalParticiple VerbForm = "v.part"
	AdverbParticiple VerbForm = "adv.part"
	RelativeParticiple VerbForm = "rel.part"
	ConditionalParticiple VerbForm = "cond.part"
)

type Tense string
const (
	Past    Tense = "past"
	Present Tense = "pres"
	Future  Tense = "fut"
)

type Voice string
const (
	Active    Voice = "act"
	Passive   Voice = "pass"
	Middle    Voice = "mid"
	Causative Voice = "caus"
)

type Mood string
const (
	Indicative  Mood = "ind"
	Imperative  Mood = "imp"
	Optative    Mood = "opt"
	Subjunctive Mood = "subj"
)

type Polarity string
const (
	Positive Polarity = "pos"
	Negative Polarity = "neg"
)

type SandhiType string
const (
	NoSandhi     SandhiType = "none"
	IdaiyinamSandhi SandhiType = "idaiyinam"  // இடையினப் புணர்ச்சி
	VallinamSandhi  SandhiType = "vallinam"   // வல்லினப் புணர்ச்சி
	MellinamSandhi  SandhiType = "mellinam"   // மெல்லினப் புணர்ச்சி
)

type CompoundType string
const (
	Vetrrumai   CompoundType = "vetrrumai"    // வேற்றுமைத் தொகை
	Vinai       CompoundType = "vinai"        // வினைத் தொகை
	Panbu       CompoundType = "panbu"        // பண்புத் தொகை
	Uvama       CompoundType = "uvama"        // உவமைத் தொகை
	Ummai       CompoundType = "ummai"        // உம்மைத் தொகை
	Anmozhi     CompoundType = "anmozhi"      // அன்மொழித் தொகை
)

type HonorificLevel string
const (
	Respectful    HonorificLevel = "resp"     // மரியாதை
	Familiar      HonorificLevel = "fam"      // இயல்பு
	Intimate      HonorificLevel = "int"      // இழிவு
)

type Length string
const (
	Kuril  Length = "kuril"    // குறில்
	Nedil  Length = "nedil"    // நெடில்
)

type LitCategory string
const (
	Sentamil    LitCategory = "sentamil"     // செந்தமிழ்
	Koduntamil  LitCategory = "koduntamil"   // கொடுந்தமிழ்
)

type RegisterLevel string
const (
	Formal    RegisterLevel = "formal"
	Informal  RegisterLevel = "informal"
	Literary  RegisterLevel = "literary"
	Colloquial RegisterLevel = "colloquial"
)

// Helper methods

// IsVerb returns true if the token is a verb
func (t *Tkn) IsVerb() bool {
	return t.PartOfSpeech == "verb"
}

// IsRational returns true if the token refers to a rational being
func (t *Tkn) IsRational() bool {
	return t.Rationality == Rational
}

// HasSandhi returns true if the token undergoes sandhi changes
func (t *Tkn) HasSandhi() bool {
	return t.SandhiType != NoSandhi
}

