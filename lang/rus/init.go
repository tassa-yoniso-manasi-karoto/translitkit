
package rus

import (
	"fmt"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/lang/mul"
)

func init() {
	defaultProviders := []common.ProviderEntry{
		{
			Provider:     &mul.UnisegProvider{},
			Capabilities: []string{"tokenization"},
			Type:        common.TokenizerType,
		},
		{
			Provider:     mul.NewIuliiaProvider(Lang),
			Capabilities: []string{"transliteration"},
			Type:        common.TransliteratorType,
		},
	}

	err := common.SetDefault(Lang, defaultProviders)
	if err != nil {
		panic(fmt.Sprintf("failed to set default providers: %v", err))
	}
}