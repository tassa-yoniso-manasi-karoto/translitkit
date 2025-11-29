package tel

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Telugu-specific features
type Tkn struct {
	common.Tkn

	// Script features
	HasVattu      bool           // Sub-script consonant marker
	HasPollu      bool           // Virama/Pollu marker
	HasAvagraha   bool           // Avagraha (rare, used in Sanskrit loanwords)
	
	// Case system (విభక్తి)
	Case          GramCase       // 7/8 cases in Telugu
	
	// Gender-Number-Person features
	Gender        Gender         // Masculine, Feminine, Neuter
	Number        Number         // Singular, Plural
	Person        Person         // 1st, 2nd, 3rd person
	Rationality   Rationality    // Rational vs Non-rational distinction (మహదాది/అమహదాది)
	
	// Verb features
	VerbForm      VerbForm      // Finite, Infinite, Participle types
	Tense         Tense         // Past, Present, Future
	Aspect        Aspect        // Perfect, Imperfect, Progressive, Habitual
	Voice         Voice         // Active, Passive, Causative
	Mood          Mood          // Indicative, Imperative, Conditional, etc.
	Transitivity  Transitivity  // Transitive, Intransitive
	
	// Sandhi (సంధి) features
	SandhiType    SandhiType    // Type of morphophonemic change
	PreSandhi     string        // Form before sandhi
	PostSandhi    string        // Form after sandhi
	
	// Word formation
	Compounds     []string      // Parts of compound words
	CompoundType  CompoundType  // Type of compound (సమాస)
	
	// Honorific and respect levels
	Honorific     HonorificLevel // Different levels of respect
	
	// Derivational features
	RootWord      string        // Root/stem form
	Derivation    DerivationType // Type of derivation
	
	// Register and style
	Register      Register      // Formal, Literary, Colloquial
	Etymology     Etymology     // Native Telugu, Sanskrit, Persian/Urdu, etc.
	
	// Additional linguistic features
	Reduplication bool         // Reduplicated form
	Emphasis      bool         // Emphatic particle usage
	Question      bool         // Interrogative form
}

// Enums for Telugu linguistic features
type GramCase string
const (
	Nominative    GramCase = "nom"    // ప్రథమ
	Accusative    GramCase = "acc"    // ద్వితీయ
	Instrumental  GramCase = "ins"    // తృతీయ
	Dative        GramCase = "dat"    // చతుర్థి
	Ablative      GramCase = "abl"    // పంచమి
	Genitive      GramCase = "gen"    // షష్ఠి
	Locative      GramCase = "loc"    // సప్తమి
	Vocative      GramCase = "voc"    // సంబోధన
)

type Gender string
const (
	MasculineRational   Gender = "m.rat"    // పుంలింగం (మహదాది)
	FeminineRational    Gender = "f.rat"    // స్త్రీలింగం (మహదాది)
	NeuterNonRational   Gender = "n.nrat"   // నపుంసకలింగం (అమహదాది)
)

type Number string
const (
	Singular Number = "sg"    // ఏకవచనం
	Plural   Number = "pl"    // బహువచనం
)

type Person string
const (
	First  Person = "1"      // ఉత్తమ పురుష
	Second Person = "2"      // మధ్యమ పురుష
	Third  Person = "3"      // ప్రథమ పురుష
)

type Rationality string
const (
	Rational    Rationality = "మహదాది"
	NonRational Rationality = "అమహదాది"
)

type VerbForm string
const (
	Finite           VerbForm = "fin"
	Infinitive       VerbForm = "inf"
	VerbalParticiple VerbForm = "v.part"
	RelativeParticiple VerbForm = "rel.part"
	ConditionalParticiple VerbForm = "cond.part"
	AdverbialParticiple VerbForm = "adv.part"
)

type Tense string
const (
	Past    Tense = "past"
	Present Tense = "pres"
	Future  Tense = "fut"
)

type Aspect string
const (
	Perfect    Aspect = "perf"
	Imperfect  Aspect = "imperf"
	Progressive Aspect = "prog"
	Habitual   Aspect = "hab"
)

type Voice string
const (
	Active    Voice = "act"
	Passive   Voice = "pass"
	Causative Voice = "caus"
)

type Mood string
const (
	Indicative  Mood = "ind"
	Imperative  Mood = "imp"
	Conditional Mood = "cond"
	Potential   Mood = "pot"
	Hortative   Mood = "hort"
)

type Transitivity string
const (
	Transitive   Transitivity = "trans"
	Intransitive Transitivity = "intrans"
)

type SandhiType string
const (
	NoSandhi      SandhiType = "none"
	AkaaraSandhi  SandhiType = "akaara"
	SavarnaSandhi SandhiType = "savarna"
	GunaSandhi    SandhiType = "guna"
	VruddhiSandhi SandhiType = "vruddhi"
)

type CompoundType string
const (
	Tatpurusha  CompoundType = "tatpurusha"
	Dvandva     CompoundType = "dvandva"
	Bahuvrihi   CompoundType = "bahuvrihi"
	Karmadharaya CompoundType = "karmadharaya"
)

type HonorificLevel string
const (
	Respectful HonorificLevel = "resp"    // గౌరవార్థక
	Familiar   HonorificLevel = "fam"     // సామాన్య
	Intimate   HonorificLevel = "int"     // సన్నిహిత
)

type DerivationType string
const (
	Nominal     DerivationType = "nom"
	Verbal      DerivationType = "verb"
	Adjectival  DerivationType = "adj"
	Adverbial   DerivationType = "adv"
)

type Register string
const (
	Formal     Register = "formal"     // గ్రాంథిక
	Literary   Register = "literary"   // సాహిత్య
	Colloquial Register = "colloquial" // వ్యావహారిక
)

type Etymology string
const (
	Native   Etymology = "native"   // తెలుగు
	Sanskrit Etymology = "sanskrit" // సంస్కృత
	PersoUrdu Etymology = "perurdu" // పర్షియన్/ఉర్దూ
	English  Etymology = "english"  // ఆంగ్ల
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

// IsCompound returns true if the token is a compound word
func (t *Tkn) IsCompound() bool {
	return len(t.Compounds) > 1
}
