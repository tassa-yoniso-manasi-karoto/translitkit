// Code generated by generator; DO NOT EDIT.

package tam

import (
	"fmt"
	"reflect"

	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const Lang = "tam" // Tamil

type Module struct {
	*common.Module
}

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

type TknSliceWrapper struct {
	common.TknSliceWrapper
	NativeSlice []*Tkn
}

// Tokens returns the token slice wrapper without filtering out non-lexical tokens.
func (m *Module) Tokens(input string) (*TknSliceWrapper, error) {
	tsw, err := m.Module.Tokens(input)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("lang/%s: %w", Lang, err)
	}
	customTsw, ok := tsw.(*TknSliceWrapper)
	if !ok {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of %s.TknSliceWrapper: real type is %s", Lang, reflect.TypeOf(tsw))
	}

	tkns, err := assertLangSpecificTokens(customTsw.Slice)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("failed assertion of []%s.Tkn: %w", Lang, err)
	}
	customTsw.NativeSlice = tkns
	return customTsw, nil
}

// Tokens returns a filtered token slice wrapper containing only tokens with lexical content.
// It calls Tokens() and then applies the Filter() method on its output,
// thereby avoiding re‑processing via additional module methods.
func (m *Module) LexicalTokens(input string) (*TknSliceWrapper, error) {
	raw, err := m.Tokens(input)
	if err != nil {
		return &TknSliceWrapper{}, fmt.Errorf("lang/%s: %w", Lang, err)
	}
	return raw.Filter(), nil
}

// Filter returns a new TknSliceWrapper containing only tokens that have lexical content.
// It processes the Tokens output without invoking further module-level processing.
func (w *TknSliceWrapper) Filter() *TknSliceWrapper {
	filtered := &TknSliceWrapper{
		TknSliceWrapper: common.TknSliceWrapper{},
		NativeSlice: make([]*Tkn, 0, len(w.NativeSlice)),
	}
	// Iterate over the tokens using the common interface's methods.
	for i := 0; i < w.Len(); i++ {
		token := w.GetIdx(i)
		nativeToken := w.NativeSlice[i]
		if token.IsLexicalContent() {
			filtered.Append(token)
			filtered.NativeSlice = append(filtered.NativeSlice, nativeToken)
		}
	}
	return filtered
}


func assertLangSpecificTokens(anyTokens []common.AnyToken) ([]*Tkn, error) {
	tokens := make([]*Tkn, len(anyTokens))
	for i, t := range anyTokens {
		token, ok := t.(*Tkn)
		if !ok {
			return nil, fmt.Errorf("token at index %d is not a %s.Tkn: real type is %s", i, Lang, reflect.TypeOf(t))
		}
		tokens[i] = token
	}
	return tokens, nil
}

