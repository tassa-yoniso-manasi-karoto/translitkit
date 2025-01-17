package jpn

import (
	"fmt"
	"strings"
	
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

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



// TODO Maybe automatically return Katakana or Hiragana as fit

// Returns a tokenized string of Hiragana readings
func (m *Module) Kana(input string) (string, error) {
	if m.Transliterator == nil && m.ProviderType != common.CombinedType {
		return "", fmt.Errorf("Kana requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return "", err
	}
	return tkns.Kana(), nil
}

// Returns a slice of string of Hiragana readings
func (m *Module) KanaParts(input string) ([]string, error) {
	if m.Transliterator == nil && m.ProviderType != common.CombinedType {
		return []string{}, fmt.Errorf("KanaParts requires either a transliterator or combined provider (got %s)", m.ProviderType)
	}
	tkns, err := m.Tokens(input)
	if err != nil {
		return []string{}, err
	}
	return tkns.KanaParts(), nil
}


func (wrapper TknSliceWrapper) Kana() string {
	return strings.Join(wrapper.KanaParts(), " ")
}

func (wrapper TknSliceWrapper) KanaParts() []string {
	var parts []string
	for _, token := range wrapper.NativeSlice {
		if token.Tkn.IsToken && token.Hiragana != "" {
			parts = append(parts, token.Hiragana)
		} else {
			parts = append(parts, token.Tkn.Surface)
		}
	}
	return parts
}


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
		// Set part of speech from first gloss FIXME
		jt.PartOfSpeech = it.Gloss[0].Pos

		// Convert Ichiran glosses to common glosses
		jt.Glosses = make([]common.Gloss, len(it.Gloss))
		for i, g := range it.Gloss {
			jt.Glosses[i] = common.Gloss{
				PartOfSpeech: g.Pos,
				Definition:   g.Gloss,
				Info:        g.Info,
			}
		}
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

// ToAnyTokenSlice converts all ichiran.JSONTokens to []common.AnyToken with underlying type []jpn.Tkn
//
//	NOTE: Golang limitation: the function's return type must explicitly be set to common.AnyTokenSliceWrapper.
//	It CAN NOT be inferred from jpn.TknSliceWrapper even if it implements the AnyTokenSliceWrapper interface.
func ToAnyTokenSlice(JSONTokens *ichiran.JSONTokens) (tkns []common.AnyToken) {
	for _, token := range *JSONTokens {
		tkns = append(tkns, ToJapaneseToken(token))
	}
	return
}

