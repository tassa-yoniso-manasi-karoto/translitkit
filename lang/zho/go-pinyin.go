package zho

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// GoPinyinProvider now chooses the "most frequent" reading for Tkn.Pinyin
// while also storing all alternatives in Tkn.PinyinAll & Tkn.PinyinNumAll.
type GoPinyinProvider struct {
	ctx    context.Context
	config map[string]interface{}

	chosenScheme string
	mainStyle    int
	numStyle     int

	mainArgs pinyin.Args
	numArgs  pinyin.Args
}

// WithContext attaches a context if desired.
func (p *GoPinyinProvider) WithContext(ctx context.Context) {
	if ctx != nil {
		p.ctx = ctx
	}
}

func (p *GoPinyinProvider) WithProgressCallback(callback common.ProgressCallback) {
}

// SaveConfig obtains user config, e.g. {"scheme":"tone2"}.
func (p *GoPinyinProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}

// Init sets pinyin styles. Now we do heteronym = true to gather all readings.
func (p *GoPinyinProvider) Init() error {
	// If mainArgs.Style != 0, we've done it once already
	if p.mainArgs.Style != 0 {
		return nil
	}

	schemeName, _ := p.config["scheme"].(string)
	if schemeName == "" {
		schemeName = "tone" // default diacritic
	}
	p.chosenScheme = schemeName

	style, ok := PinyinSchemes[strings.ToLower(schemeName)]
	if !ok {
		style = pinyin.Tone
	}
	p.mainStyle = style
	p.numStyle = pinyin.Tone2

	// Prepare mainArgs
	p.mainArgs = pinyin.NewArgs()
	p.mainArgs.Style = p.mainStyle
	p.mainArgs.Heteronym = true // gather multiple possible pronunciations

	// Prepare numArgs
	p.numArgs = pinyin.NewArgs()
	p.numArgs.Style = p.numStyle
	p.numArgs.Heteronym = true // also gather multiple numeric variants

	return nil
}

// InitRecreate re-initializes from scratch.
func (p *GoPinyinProvider) InitRecreate(noCache bool) error {
	p.mainArgs = pinyin.Args{}
	p.numArgs = pinyin.Args{}
	p.mainStyle = 0
	p.numStyle = 0
	return p.Init()
}

// ProcessFlowController: if token is Chinese and lexical, we compute multiple
// pinyin readings, store them, and pick the first (most frequent) for Tkn.Pinyin.
func (p *GoPinyinProvider) ProcessFlowController(
	input common.AnyTokenSliceWrapper,
) (common.AnyTokenSliceWrapper, error) {

	if err := p.Init(); err != nil {
		return nil, fmt.Errorf("gopinyin init failed: %w", err)
	}

	for i := 0; i < input.Len(); i++ {
		anyTkn := input.GetIdx(i)
		if !anyTkn.IsLexicalContent() {
			continue
		}

		zhoTkn, ok := anyTkn.(*Tkn)
		if !ok {
			// Not our specialized token => fallback
			anyTkn.SetRoman(anyTkn.GetSurface())
			continue
		}

		if !zhoTkn.IsChinese() {
			zhoTkn.SetRoman(zhoTkn.Surface)
			continue
		}

		// 1) Retrieve diacritic multi-pronunciation data
		allSyllables := pinyin.Pinyin(zhoTkn.Surface, p.mainArgs) // 2D slice
		zhoTkn.PinyinAll = allSyllables

		// 2) Retrieve numeric multi-pronunciation data
		allNumSyllables := pinyin.Pinyin(zhoTkn.Surface, p.numArgs)
		zhoTkn.PinyinNumAll = allNumSyllables

		// 3) The "most frequent" reading is the *first* in each sub-slice.
		// We'll build Tkn.Pinyin from that.
		var chosenDiacritic []string
		var chosenNumeric []string

		for idxChar, arr := range allSyllables {
			if len(arr) > 0 {
				chosenDiacritic = append(chosenDiacritic, arr[0])
			} else {
				// fallback if no reading
				chosenDiacritic = append(chosenDiacritic, "")
			}

			numArr := allNumSyllables[idxChar]
			if len(numArr) > 0 {
				chosenNumeric = append(chosenNumeric, numArr[0])
			} else {
				chosenNumeric = append(chosenNumeric, "")
			}
		}

		zhoTkn.Pinyin = strings.Join(chosenDiacritic, " ")
		zhoTkn.PinyinNum = strings.Join(chosenNumeric, " ")

		// 4) If single-syllable, parse numeric tone
		if len(chosenNumeric) == 1 {
			toneVal := parseToneNumber(chosenNumeric[0])
			if toneVal > 0 {
				zhoTkn.Tone = Tone(toneVal)
				zhoTkn.OriginalTone = zhoTkn.Tone
				zhoTkn.HasToneSandhi = false
			}
		}

		// 5) Put the final reading in Tkn.Romanization
		zhoTkn.SetRoman(zhoTkn.Pinyin)
	}

	return input, nil
}

// Name identifies this provider as "gopinyin".
func (p *GoPinyinProvider) Name() string {
	return "gopinyin"
}

// GetType returns Transliterator.
func (p *GoPinyinProvider) GetType() common.ProviderType {
	return common.TransliteratorType
}

// GetMaxQueryLen is large for go-pinyin in memory usage.
func (p *GoPinyinProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

// Close is a no-op for go-pinyin.
func (p *GoPinyinProvider) Close() error {
	return nil
}

// PinyinSchemes maps user-friendly scheme names to pinyin int constants.
var PinyinSchemes = map[string]int{
	"normal":       pinyin.Normal,
	"tone":         pinyin.Tone,
	"tone2":        pinyin.Tone2,
	"tone3":        pinyin.Tone3,
	"initials":     pinyin.Initials,
	"firstletter":  pinyin.FirstLetter,
	"finals":       pinyin.Finals,
	"finalstone":   pinyin.FinalsTone,
	"finalstone2":  pinyin.FinalsTone2,
	"finalstone3":  pinyin.FinalsTone3,
}

// parseToneNumber picks the last digit [1..5] from a tone2 syllable like "hao3".
func parseToneNumber(s string) int {
	re := regexp.MustCompile(`(\d)$`)
	match := re.FindStringSubmatch(s)
	if len(match) < 2 {
		return 0
	}
	num, _ := strconv.Atoi(match[1])
	return num
}
