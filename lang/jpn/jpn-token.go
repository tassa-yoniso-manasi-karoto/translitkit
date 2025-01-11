package jpn

import (
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	common "github.com/tassa-yoniso-manasi-karoto/translitkit"

	//iso "github.com/barbashov/iso639-3"
)

type JapaneseSliceTkns struct {
	common.Tkns
}

// Tkn extends common Token with Japanese-specific features
type Tkn struct {
	common.Tkn

	// Japanese Writing Systems
	Kanji    string // 漢字 representation
	Hiragana string // ひらがな representation
	Katakana string // カタカナ representation

	// Japanese Linguistic Features
	Okurigana string // Kana suffix following kanji
	Pitch     []int  // Pitch accent pattern
	MoraCount int    // Number of morae

	// Word Formation
	IsKango    bool // 漢語 (Sino-Japanese word)
	IsWago     bool // 和語 (Native Japanese word)
	IsGairaigo bool // 外来語 (Foreign loanword)

	// Conjugation Information
	BaseForm   string // Dictionary form (見出し語)
	Inflection struct {
		Type     string // Conjugation type (五段, 一段, etc.)
		Form     string // Inflected form (て形, た形, etc.)
		Polite   bool   // です/ます style
		Negative bool   // Negative form
	}

	// Additional Features
	IsHonorific bool   // 敬語 (Honorific form)
	IsHumble    bool   // 謙譲語 (Humble form)
	IsKeigo     bool   // General keigo flag
	Register    string // Language register (formal, casual, etc.)
}

func (t Tkn) GetSurface() string {
	return t.Surface
}

func (t Tkn) GetRomanization() string {
	if !t.IsToken || t.Surface == t.Romanization {
		return ""
	}
	return t.Romanization
}

func (t Tkn) IsTokenType() bool {
	return true
}

type JapaneseModule struct {
	language       string
	providerType   common.ProviderType
	tokenizer      common.Provider[common.AnyTokenSliceWrapper, common.AnyTokenSliceWrapper]
	transliterator common.Provider[common.AnyTokenSliceWrapper, common.AnyTokenSliceWrapper]
	combined       common.Provider[common.AnyTokenSliceWrapper, common.AnyTokenSliceWrapper]
	MaxLenQuery    int
}

/*func (m *JapaneseModule) KanaParts(input string) ([]string, error) {
	if m.transliterator == nil && m.providerType != common.CombinedType {
		return nil, fmt.Errorf("katakana requires either a transliterator or combined provider (got %s)", m.providerType)
	}
	m.outputType = common.KanaType
	return m.GetSlice(input) FIXME  m.GetSlice undefined (type *JapaneseModule has no field or method GetSlice)
}*/

// ToJapaneseToken converts an JSONToken to a Tkn
func ToJapaneseToken(it *ichiran.JSONToken) (jt Tkn) {
	// Fill common Tkn fields
	jt.Surface = it.Surface
	jt.IsToken = it.IsToken

	// If this is not a Japanese token, return early with minimal information
	if !it.IsToken {
		return jt
	}
	jt.Metadata = make(map[string]interface{})

	// Continue with Japanese-specific token processing
	jt.Normalized = it.Surface // Could be enhanced with actual normalization
	jt.Position.Start = it.Seq
	jt.Confidence = float64(it.Score)
	jt.Language = "jpn"
	jt.Script = "Jpan"
	jt.Romanization = it.Romaji

	// Set Japanese-specific fields
	jt.Kanji = it.Surface
	jt.Hiragana = it.Kana

	// Process glosses
	if len(it.Gloss) > 0 {
		// Set part of speech from first gloss
		jt.PartOfSpeech = it.Gloss[0].Pos

		// Store all glosses in metadata
		glosses := make([]string, len(it.Gloss))
		for i, g := range it.Gloss {
			glosses[i] = g.Gloss
		}
		jt.Metadata["glosses"] = glosses
	}

	// Process conjugation information
	if len(it.Conj) > 0 {
		conj := it.Conj[0] // Take first conjugation

		jt.BaseForm = conj.Reading

		// Process properties
		for _, prop := range conj.Prop {
			switch {
			case prop.Type == "polite":
				jt.Inflection.Polite = true
			case prop.Neg:
				jt.Inflection.Negative = true
			}

			// Store conjugation type
			if prop.Type != "" {
				jt.Inflection.Type = prop.Type
			}
		}
	}

	// Store original Ichiran data in metadata
	jt.Metadata["ichiran"] = map[string]interface{}{
		"score":       it.Score,
		"alternative": it.Alternative,
		"raw":         string(it.Raw),
	}

	return jt
}

// ToTokenSlice converts all ichiran.JSONTokens to JapaneseSliceTkns
//
//	NOTE: Golang limitation: the function's return type must explicitly be set to common.AnyTokenSliceWrapper.
//	It CAN NOT be inferred from JapaneseSliceTkns even if it implements the AnyTokenSliceWrapper interface.
func ToTokenSlice(JSONTokens *ichiran.JSONTokens) (tkns common.AnyTokenSliceWrapper) {
	tkns = JapaneseSliceTkns{common.Tkns{Slice: make([]common.AnyToken, 0)}}

	for _, token := range *JSONTokens {
		inter := ToJapaneseToken(token)
		tkns = tkns.Append(inter)
	}
	return
}

// ToGeneric converts the Japanese token to a generic token
func (t *Tkn) ToGeneric() common.Tkn {
	// Store Japanese-specific information in metadata
	t.Metadata["japanese"] = map[string]interface{}{
		// Writing Systems
		"kanji":    t.Kanji,
		"hiragana": t.Hiragana,
		"katakana": t.Katakana,

		// Linguistic Features
		"okurigana": t.Okurigana,
		"pitch":     t.Pitch,
		"moraCount": t.MoraCount,

		// Word Formation
		"isKango":    t.IsKango,
		"isWago":     t.IsWago,
		"isGairaigo": t.IsGairaigo,

		// Conjugation
		"baseForm": t.BaseForm,
		"inflection": map[string]interface{}{
			"type":     t.Inflection.Type,
			"form":     t.Inflection.Form,
			"polite":   t.Inflection.Polite,
			"negative": t.Inflection.Negative,
		},

		// Additional Features
		"isHonorific": t.IsHonorific,
		"isHumble":    t.IsHumble,
		"isKeigo":     t.IsKeigo,
		"register":    t.Register,
	}

	return t.Tkn
}
