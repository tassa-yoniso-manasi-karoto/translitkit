// zho_test.go
package zho_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tassa-yoniso-manasi-karoto/translitkit"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/lang/zho"
)

// Sample texts
var sampleText = "你好吗，世界？"
var longText = "从前，有一个非常勤奋的学生，他每天都早起跑步，并且坚持背诵古诗，成绩始终名列前茅。"
var shortText = "你好"
var mixedText = "Hello 你好 123"

func TestGoJieba_TokenizerBasic(t *testing.T) {
	prov := &zho.GoJiebaProvider{}
	require.NoError(t, prov.Init())

	wrapper := &zho.TknSliceWrapper{
		TknSliceWrapper: common.TknSliceWrapper{
			Raw: []string{sampleText},
		},
	}
	out, err := prov.ProcessFlowController(wrapper)
	require.NoError(t, err)

	var surfaces []string
	for i := 0; i < out.Len(); i++ {
		surfaces = append(surfaces, out.GetIdx(i).GetSurface())
	}
	t.Logf("Tokenizer result surfaces: %#v", surfaces)
	// Possibly => ["你好", "吗", "，", "世界", "？"] or ["你", "好", "吗", "，", ...] – depends on dictionary

	// We at least expect "你好" or "你" "好"
	foundNiHao := false
	for _, s := range surfaces {
		if strings.Contains(s, "你好") {
			foundNiHao = true
			break
		}
	}
	assert.True(t, foundNiHao,
		"Tokenizer should contain either '你好' as a single token or '你' and '好'")
}

func TestGoJieba_EdgeCases(t *testing.T) {
	prov := &zho.GoJiebaProvider{}
	require.NoError(t, prov.Init())

	// 1) Empty input
	w1 := &zho.TknSliceWrapper{
		TknSliceWrapper: common.TknSliceWrapper{Raw: []string{""}},
	}
	out1, err1 := prov.ProcessFlowController(w1)
	require.NoError(t, err1)
	assert.Equal(t, 0, out1.Len())

	// 2) ASCII
	w2 := &zho.TknSliceWrapper{
		TknSliceWrapper: common.TknSliceWrapper{Raw: []string{"Hello world!"}},
	}
	out2, err2 := prov.ProcessFlowController(w2)
	require.NoError(t, err2)
	assert.GreaterOrEqual(t, out2.Len(), 1, "Should produce tokens from ASCII")

	// Just log them
	var surfaces []string
	for i := 0; i < out2.Len(); i++ {
		surfaces = append(surfaces, out2.GetIdx(i).GetSurface())
	}
	t.Logf("ASCII token surfaces: %+v", surfaces)
}

func TestGoPinyinProvider_BasicTone(t *testing.T) {
	pprov := &zho.GoPinyinProvider{}
	pprov.SaveConfig(map[string]interface{}{"scheme": "tone"}) // diacritics
	require.NoError(t, pprov.Init())

	// Suppose we have tokens for "你好"
	wrapper := &zho.TknSliceWrapper{}
	wrapper.Append(
		&zho.Tkn{
			Tkn: common.Tkn{
				Surface:   "你",
				IsLexical: true,
			},
		},
		&zho.Tkn{
			Tkn: common.Tkn{
				Surface:   "好",
				IsLexical: true,
			},
		},
	)
	out, err := pprov.ProcessFlowController(wrapper)
	require.NoError(t, err)
	require.Equal(t, 2, out.Len())

	tkn1 := out.GetIdx(0).(*zho.Tkn)
	tkn2 := out.GetIdx(1).(*zho.Tkn)

	t.Logf("Token1 => %s, Token2 => %s", tkn1.Pinyin, tkn2.Pinyin)
	// We check partial match for "nǐ" / "hǎo"
	assert.Contains(t, tkn1.Pinyin, "nǐ", "Should contain 'nǐ'")
	assert.Contains(t, tkn2.Pinyin, "hǎo", "Should contain 'hǎo'")
}

func TestGoPinyinProvider_SchemeTone2(t *testing.T) {
	pprov := &zho.GoPinyinProvider{}
	pprov.SaveConfig(map[string]interface{}{"scheme": "tone2"}) // numeric
	require.NoError(t, pprov.Init())

	wrapper := &zho.TknSliceWrapper{}
	wrapper.Append(
		&zho.Tkn{
			Tkn: common.Tkn{Surface: "你", IsLexical: true},
		},
		&zho.Tkn{
			Tkn: common.Tkn{Surface: "好", IsLexical: true},
		},
	)

	out, err := pprov.ProcessFlowController(wrapper)
	require.NoError(t, err)
	require.Equal(t, 2, out.Len())

	tkn1 := out.GetIdx(0).(*zho.Tkn)
	tkn2 := out.GetIdx(1).(*zho.Tkn)

	t.Logf("Tone2 => Tkn1:%s, Tkn2:%s", tkn1.Pinyin, tkn2.Pinyin)
	// We see "ni3" or "ha3o" or "hao3"? 
	// Some dictionaries produce "ha3o" but let's do a partial check to ensure numeric + "3"
	assert.Contains(t, tkn1.Pinyin, "3", "Should contain numeric tone")
	assert.Contains(t, tkn2.Pinyin, "3", "Should contain numeric tone")
}

func TestZhoModule_DefaultPipeline(t *testing.T) {
	m, err := translitkit.DefaultModule("zho")
	require.NoError(t, err)
	m.MustInit()
	defer m.Close()

	// Check short text
	roman, err := m.Roman(shortText) // "你好"
	require.NoError(t, err)
	t.Logf("Roman(你好) => %s", roman)
	// Might be "nǐ hǎo" or "ni3 hao3"
	assert.True(t, strings.Contains(roman, "nǐ") || strings.Contains(roman, "ni3"))

	// Tokenize sample text
	tok, err := m.Tokenized(sampleText)
	require.NoError(t, err)
	t.Logf("Tokenized sample => %s", tok)

	roman2, err := m.Roman(sampleText)
	require.NoError(t, err)
	t.Logf("Roman sample => %s", roman2)

	// Now the longer text
	romanLong, err := m.Roman(longText)
	require.NoError(t, err)
	t.Logf("Roman (long text) => %s", romanLong)

	// We'll check partial keys
	assert.Contains(t, romanLong, "cóng", "Should see pinyin for '从'")
	assert.Contains(t, romanLong, "qián", "Should see pinyin for '前'")
	assert.Contains(t, romanLong, "xué", "Should see pinyin for '学'")
	assert.Contains(t, romanLong, "shēng", "Should see pinyin for '生'")
}

func TestZhoModule_EdgeCases(t *testing.T) {
	m, err := translitkit.DefaultModule("zho")
	require.NoError(t, err)
	m.MustInit()
	defer m.Close()

	// 1) ASCII
	asciiText := "Hello 123"
	roman, err := m.Roman(asciiText)
	require.NoError(t, err)

	// Because chunkification might add spaces, we do a "whitespace-insensitive" check:
	romanNoSpaces := strings.ReplaceAll(roman, " ", "")
	asciiExpected := strings.ReplaceAll(asciiText, " ", "")
	assert.Equal(t, asciiExpected, romanNoSpaces, "ASCII text might remain the same ignoring extra spaces")

	// 2) Mixed text
	romanMix, err := m.Roman(mixedText)
	require.NoError(t, err)
	t.Logf("Mixed text => %s", romanMix)
	// e.g. "Hello   ni3 hao3   123" or "Hello nǐ hǎo  123"
	// We'll check partial
	assert.Contains(t, romanMix, "Hello")
	assert.True(t, strings.Contains(romanMix, "ni3") || strings.Contains(romanMix, "nǐ"),
		"Should contain the pinyin for 你")
}
