// Code generated by generator; DO NOT EDIT.

package hin

import (
	"fmt"
	"reflect"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const Lang = "hin"

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

func assertLangSpecificTokens(anyTokens []common.AnyToken) ([]*Tkn, error) {
	tokens := make([]*Tkn, len(anyTokens))
	for i, t := range anyTokens {
		token, ok := t.(*Tkn)
		if !ok {
			return nil, fmt.Errorf("token at index %d is not a %s.Tkn: real type is %s", i, Lang, reflect.TypeOf(token))
		}
		tokens[i] = token
	}
	return tokens, nil
}