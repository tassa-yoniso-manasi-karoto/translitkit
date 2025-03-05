package zho

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Chinese-specific features
type Tkn struct {
	common.Tkn

	// Character features
	Simplified    string         // Simplified form
	Traditional   string         // Traditional form
	Variants      []string      // Character variants (异体字)
	NumStrokes    int           // Number of strokes
	Radical       string        // Character radical (部首)
	Components    []string      // Character components
	
	// Phonological features
	Pinyin       string         // Standard Pinyin
	// PinyinAll holds the multi-pronunciation data from go-pinyin for diacritic forms.
	// If the token has N characters, PinyinAll[i] is an array of possible readings
	// for the i-th character.
	PinyinAll [][]string
	
	PinyinNum    string         // Pinyin with tone numbers
	// PinyinNumAll does the same for numeric pinyin.
	PinyinNumAll [][]string

	Zhuyin       string         // Bopomofo/Zhuyin
	Tone         Tone           // Tone value
	OriginalTone Tone           // Original tone before sandhi
	HasToneSandhi bool         // Whether tone sandhi applies
	
	// Word formation
	Morphemes    []Morpheme    // Individual morpheme analysis
	MorphType    MorphType     // Type of morphological construction
	IsCompound   bool          // Whether token is a compound word
	CompoundType CompoundType  // Type of compound construction
	
	// Semantic features
	Measure      string        // Measure word (量词) if applicable
	ClassifierType ClassifierType // Type of classifier if applicable
	Register     Register      // Literary/formal/informal/etc.
	Style        Style         // Written/spoken style
	Etymology    Etymology     // Word origin
	
	// Grammatical features
	AspectMarker Aspect       // Aspect marking (了,过,着)
	Resultative  string       // Resultative complement
	Directional  string       // Directional complement
	IsStative    bool         // Whether verb/adjective is stative
	IsSeparable  bool         // Separable verb (离合词)
	
	// Idiomatic features
	Chengyu      bool         // Whether token is a chengyu (成语)
	Xiehouyu     bool         // Whether token is a xiehouyu (歇后语)
	Idiom        bool         // Other types of idioms
	
	// Modern/Classical features
	IsClassical  bool         // Whether token is Classical Chinese
	ModernUsage  bool         // Whether used in Modern Chinese
}

// Morpheme represents a single Chinese morpheme
type Morpheme struct {
	Character    string
	Simplified   string
	Traditional  string
	Pinyin       string
	Tone         Tone
	Meaning      string
	Function     MorphFunction
}

// Enums for Chinese linguistic features
type Tone int
const (
	First   Tone = 1  // 阴平
	Second  Tone = 2  // 阳平
	Third   Tone = 3  // 上声
	Fourth  Tone = 4  // 去声
	Neutral Tone = 5  // 轻声
)

type MorphType string
const (
	Single       MorphType = "single"      // 单纯词
	Compound     MorphType = "compound"    // 合成词
	Derived      MorphType = "derived"     // 派生词
	Reduplicated MorphType = "redup"       // 重叠词
)

type CompoundType string
const (
	Coordination CompoundType = "coord"     // 联合式
	Modification CompoundType = "mod"       // 偏正式
	Subjective   CompoundType = "subj"      // 主谓式
	Objective    CompoundType = "obj"       // 动宾式
	Complement   CompoundType = "comp"      // 补充式
	Predicate    CompoundType = "pred"      // 谓补式
)

type ClassifierType string
const (
	Individual  ClassifierType = "indiv"    // 个体量词
	Collective  ClassifierType = "coll"     // 集体量词
	Measure     ClassifierType = "meas"     // 度量量词
	Temporary   ClassifierType = "temp"     // 临时量词
)

type Register string
const (
	Literary    Register = "lit"     // 书面语
	Formal      Register = "formal"  // 正式
	Informal    Register = "inf"     // 口语
	Colloquial  Register = "colloq"  // 俗语
)

type Style string
const (
	Written     Style = "written"   // 书面
	Spoken      Style = "spoken"    // 口头
	Internet    Style = "net"       // 网络用语
)

type Etymology string
const (
	Native      Etymology = "native"    // 固有词
	SinoJapanese Etymology = "sijp"     // 日制汉语
	Western     Etymology = "western"   // 西源词
	Modern      Etymology = "modern"    // 现代词
)

type Aspect string
const (
	Perfective   Aspect = "perf"    // 了
	Experiential Aspect = "exp"     // 过
	Progressive  Aspect = "prog"    // 着
	Durative     Aspect = "dur"     // 起来
)

type MorphFunction string
const (
	Nominal      MorphFunction = "nom"
	Verbal       MorphFunction = "verb"
	Adjectival   MorphFunction = "adj"
	Adverbial    MorphFunction = "adv"
	Grammatical  MorphFunction = "gram"
)

// Helper methods

// IsChinese returns true if the character is a Chinese character
// TODO I am not sure whether this is reliable or not
func (t *Tkn) IsChinese() bool {
	for _, r := range t.Surface {
		if r < 0x4E00 || r > 0x9FFF {
			return false
		}
	}
	return true
}

// HasSimplifiedVariant returns true if the token has a different simplified form
func (t *Tkn) HasSimplifiedVariant() bool {
	return t.Simplified != "" && t.Simplified != t.Surface
}

// IsClassifier returns true if the token is a classifier
func (t *Tkn) IsClassifier() bool {
	return t.ClassifierType != ""
}
