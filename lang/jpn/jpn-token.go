package jpn

import (
	"fmt"
	"strings"
	"reflect"
	
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"

	//iso "github.com/barbashov/iso639-3"
)

type Module struct {
	*common.Module
}


// DefaultModule returns a new Japanese-specific Module configured with the default providers.
func DefaultModule() (*Module, error) {
	m, err := common.DefaultModule("jpn")
	if err != nil {
		return nil, err
	}
	jm := &Module{
		Module: m,
	}
	return jm, nil
}



func (m *Module) Tokens(input string) (*TknSliceWrapper, error) {
	tsw, err := m.Module.Tokens(input)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("lang/jpn: %v", err)
	}
	jtsw, ok := tsw.(*TknSliceWrapper)
	if !ok {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of jpn.TknSliceWrapper: real type is %s", reflect.TypeOf(tsw))
	}
	
	// takes []AnyToken, returns asserted []jpn.Token
	// TODO mesure perf impact
	tkns, err := assertJPNTokens(jtsw.Slice)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of []jpn.Tkn: %v", err)
	}
	jtsw.NativeSlice = tkns
	return jtsw, nil
}

// TODO Maybe automatically return Katakana or Hiragan as fit

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



type TknSliceWrapper struct {
	common.TknSliceWrapper
	NativeSlice []Tkn
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


func assertJPNTokens(anyTokens []common.AnyToken) ([]Tkn, error) {
	tokens := make([]Tkn, len(anyTokens))
	for i, t := range anyTokens {
		token, ok := t.(Tkn)
		if !ok {
			return nil, fmt.Errorf("token at index %d is not a jpn.Tkn", i) // add reflect type
		}
		tokens[i] = token
	}
	return tokens, nil
}

// ToGeneric converts the Japanese token to a generic token
func (t *Tkn) ToCommon() common.Tkn {
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
