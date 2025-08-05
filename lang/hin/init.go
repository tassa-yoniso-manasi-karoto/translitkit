
package hin

import (
	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	"fmt"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/lang/mul"
)

func init() {	
	defaultProviders := []common.ProviderEntry{
		{
			Provider:     &mul.UnisegProvider{},
			Capabilities: []string{"tokenization"},
		},
		{
			Provider:     mul.NewAksharamukhaProvider(Lang),
			Capabilities: []string{"transliteration"},
		},
	}

	err := common.SetDefault(Lang, defaultProviders)
	if err != nil {
		common.Log.Warn().Err(err).
			Str("pkg", Lang).
			Msg("failed to set default providers")
	}
}


func placeholder3456543() {
	fmt.Println("")
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}