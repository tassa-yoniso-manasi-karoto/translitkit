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
	}

	// B) Transliterator: GoPinyin
	gopinyinProv := &GoPinyinProvider{}
	gopinyinEntry := common.ProviderEntry{
		Provider:     gopinyinProv,
		Capabilities: []string{"transliteration"},
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
			Name:        "tone",
			Description: "Pinyin with diacritic tone marks (mā má mǎ mà)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "normal",
			Description: "Pinyin without tone marks",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "tone2",
			Description: "Pinyin with trailing numeric tone (ma1 ma2 ma3 ma4)",
			Providers:   []string{"gojieba", "gopinyin"},
		},
		{
			Name:        "tone3",
			Description: "Pinyin with inline numeric tone",
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
