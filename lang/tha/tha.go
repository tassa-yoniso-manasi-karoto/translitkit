package tha

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// Tkn extends common.Tkn with Thai-specific features
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

