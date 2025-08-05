package zho

import (
	"fmt"

	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// init runs automatically when this package is imported, registering
// and configuring providers & schemes for the Chinese language "zho".
func init() {
	///////////////////////////////////
	// 1) Create the provider entries
	///////////////////////////////////

	// A) Tokenizer: GoJieba
	gojiebaProv := &GoJiebaProvider{}
	gojiebaEntry := common.ProviderEntry{
		Provider:     gojiebaProv,
		Capabilities: []string{"tokenization"},
		Mode:         common.TokenizerMode,
	}

	// B) Transliterator: GoPinyin
	gopinyinProv := &GoPinyinProvider{}
	gopinyinEntry := common.ProviderEntry{
		Provider:     gopinyinProv,
		Capabilities: []string{"transliteration"},
		Mode:         common.TransliteratorMode,
	}

	///////////////////////////////////
	// 2) Register the providers
	///////////////////////////////////

	// Register gojieba as the tokenizer
	if err := common.Register("zho", gojiebaEntry); err != nil {
		panic(fmt.Sprintf("failed to register gojieba: %v", err))
	}

	// Register gopinyin as the transliterator
	if err := common.Register("zho", gopinyinEntry); err != nil {
		panic(fmt.Sprintf("failed to register gopinyin: %v", err))
	}

	///////////////////////////////////
	// 3) Set them as default providers
	///////////////////////////////////

	// The first is the tokenizer, the second is the transliterator.
	defaultChain := []common.ProviderEntry{
		gojiebaEntry,
		gopinyinEntry,
	}
	if err := common.SetDefault("zho", defaultChain); err != nil {
		panic(fmt.Sprintf("failed to set default providers for zho: %v", err))
	}

	///////////////////////////////////
	// 4) Register transliteration schemes for "zho"
	///////////////////////////////////

	// The following "scheme" names map to the GoPinyinProvider. 
	// They match the keys in PinyinSchemes from gopinyin_provider.go,
	// e.g. "tone", "tone2", "tone3", "initials", "firstletter", etc.
	// This lets you do:
	//   mod, err := common.GetSchemeModule("zho", "tone")
	// and get a "gopinyin" provider with that scheme set.

	zhoSchemes := []common.TranslitScheme{
		{
			Name:        "normal",
			Description: "Chinese pinyin without tone marks (pinyin.Normal)",
			Providers:   []string{"gojieba", "gopinyin"}, 
		},
		{
			Name:        "tone",
			Description: "Chinese pinyin with diacritic tone marks (pinyin.Tone)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "tone2",
			Description: "Chinese pinyin with numeric tone (pinyin.Tone2)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "tone3",
			Description: "Chinese pinyin with numeric tone variant (pinyin.Tone3)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "initials",
			Description: "Chinese pinyin initials only (pinyin.Initials)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "firstletter",
			Description: "Chinese pinyin first letter only (pinyin.FirstLetter)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "finals",
			Description: "Chinese pinyin finals only (pinyin.Finals)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "finalstone",
			Description: "Chinese pinyin finals with tone marks (pinyin.FinalsTone)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "finalstone2",
			Description: "Chinese pinyin finals with numeric tone (pinyin.FinalsTone2)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "finalstone3",
			Description: "Chinese pinyin finals with numeric tone variant (pinyin.FinalsTone3)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
	}

	for _, scheme := range zhoSchemes {
		if err := common.RegisterScheme("zho", scheme); err != nil {
			// It's normal for re-registration attempts to fail if name is duplicated
			// or to fail if "zho" not recognized, so handle or ignore as needed.
			// We'll panic for clarity here:
			panic(fmt.Sprintf("failed to register scheme %s for zho: %v", scheme.Name, err))
		}
	}

	// Now "zho" has a set of recognized transliteration scheme names
	// that map to "gopinyin" in the registry.
	///////////////////////////////////

	// That’s it! We have:
	//   - zho default providers: [gojieba -> gopinyin]
	//   - zho transliteration schemes registered: "normal", "tone", "tone2", ...
}
