package tha

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	"reflect"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const Lang = "tha"

type Module struct {
	*common.Module
}


// DefaultModule returns a new Module specific of this language configured with the default providers.
func DefaultModule() (*Module, error) {
	m, err := common.DefaultModule(Lang)
	if err != nil {
		return nil, err
	}
	customModule := &Module{
		Module: m,
	}
	return customModule, nil
}

func (m *Module) Tokens(input string) (*TknSliceWrapper, error) {
	tsw, err := m.Module.Tokens(input)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("lang/%s: %v", Lang, err)
	}
	customTsw, ok := tsw.(*TknSliceWrapper)
	if !ok {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of %s.TknSliceWrapper: real type is %s", Lang, reflect.TypeOf(tsw))
	}
	
	tkns, err := assertLangSpecificTokens(customTsw.Slice)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of []%s.Tkn: %v", Lang, err)
	}
	customTsw.NativeSlice = tkns
	return customTsw, nil
}


type TknSliceWrapper struct {
	common.TknSliceWrapper
	NativeSlice []Tkn
}


func assertLangSpecificTokens(anyTokens []common.AnyToken) ([]Tkn, error) {
	tokens := make([]Tkn, len(anyTokens))
	for i, t := range anyTokens {
		token, ok := t.(Tkn)
		if !ok {
			return nil, fmt.Errorf("token at index %d is not a %s.Tkn: real type is %s", Lang, i, reflect.TypeOf(token))
		}
		tokens[i] = token
	}
	return tokens, nil
}


// tha.Tkn extends common.Tkn with Thai-specific features
type Tkn struct {
	common.Tkn

	// Thai Syllable Structure
	InitialConsonant string // พยัญชนะต้น
	FirstConsonant   string // อักษรนำ (leading consonant)
	Vowel            string // สระ
	FinalConsonant   string // ตัวสะกด
	Tone             int    // วรรณยุกต์ (0-4)

	// Thai-specific Classifications
	ConsonantClass string // อักษรสูง, อักษรกลาง, อักษรต่ำ (high, mid, low class)
	SyllableType   string // แม่ ก กา, แม่ กง, etc.

	// Thai Word Formation
	IsPrefixWord   bool   // คำหน้า
	IsSuffixWord   bool   // คำหลัง
	IsCompoundPart bool   // ส่วนประกอบของคำประสม
	CompoundRole   string // บทประกอบของคำประสม (head, modifier)

	// Thai-specific Features
	IsKaranWord      bool // คำการันต์ (words with special ending marks)
	HasSpecialMarker bool // Contains special markers (ฯ, ๆ, etc.)
	IsAbbreviation   bool // คำย่อ
	IsRoyal          bool // ราชาศัพท์ (royal vocabulary)

	// Thai Word Categories
	IsFunction bool // คำไวยากรณ์ (grammatical word)
	IsContent  bool // คำศัพท์ (content word)

	// Additional Thai Analysis
	RegisterLevel string // ระดับภาษา (formal, informal, etc.)
	Etymology     string // ที่มาของคำ (Thai, Pali, Sanskrit, etc.)

	// Alternative Analyses
	PossibleReadings []string // Alternative pronunciations
	AlternativeTones []int    // Possible tone variations
}


// ToCommon converts ThaiToken to common Token
func (t *Tkn) ToCommon() common.Tkn {
	// Store Thai-specific information in metadata
	t.Metadata["thai"] = map[string]interface{}{
		"tone":           t.Tone,
		"consonantClass": t.ConsonantClass,
		"syllableType":   t.SyllableType,
		"isRoyal":        t.IsRoyal,
		"registerLevel":  t.RegisterLevel,
		"etymology":      t.Etymology,
	}

	// Store syllable structure
	t.Metadata["syllable"] = map[string]interface{}{
		"initial": t.InitialConsonant,
		"first":   t.FirstConsonant,
		"vowel":   t.Vowel,
		"final":   t.FinalConsonant,
	}

	return t.Tkn
}





func placeholder() {
	fmt.Print("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
