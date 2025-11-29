
package mul

import (
	"fmt"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const Lang = "mul"

var indicLangs = []string{
	"hin", "ben", "fas", "guj", "mar", "pan", "sin", "urd", "tam", "tel",
}

func init() {
	unisegEntry := common.ProviderEntry{
		Provider:     &UnisegProvider{},
		Capabilities: []string{"tokenization"},
	}
	aksharamukhaEntry := common.ProviderEntry{
		Provider:     &AksharamukhaProvider{},
		Capabilities: []string{"transliteration"},
	}
	iuliiaEntry := common.ProviderEntry{
		Provider:     NewIuliiaProvider("rus"),
		Capabilities: []string{"transliteration"},
	}
	

	err := common.Register("mul", unisegEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register uniseg provider: %w", err))
	}
	
	err = common.Register("mul", aksharamukhaEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register aksharamukha provider: %w", err))
	}
	
	err = common.Register("mul", iuliiaEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register iuliia provider: %w", err))
	}
	
	// #### Schemes registration ####

	for _, indicLang := range indicLangs {
		for _, scheme := range indicSchemes {
			scheme.Providers = []string{"aksharamukha"}
			scheme.NeedsDocker = true
			if err := common.RegisterScheme(indicLang, scheme); err != nil {
				common.Log.Warn().
					Str("pkg", Lang).
					Str("lang", indicLang).
					Msg("Failed to register scheme " + scheme.Name)
			}
		}
	}
	
	for _, scheme := range russianSchemes {
		scheme.Providers = []string{"iuliia"}
		if err := common.RegisterScheme("rus", scheme); err != nil {
			common.Log.Warn().
				Str("pkg", Lang).
				Str("lang", "rus").
				Msg("Failed to register scheme " + scheme.Name)
		}
	}

	if err := common.RegisterScheme("uzb", uzbekScheme); err != nil {
		common.Log.Warn().
			Str("pkg", Lang).
			Str("lang", "uzb").
			Msg("Failed to register scheme " + uzbekScheme.Name)
	}
}