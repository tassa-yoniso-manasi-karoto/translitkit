
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
		Type:        common.TokenizerType,
	}
	aksharamukhaEntry := common.ProviderEntry{
		Provider:     &AksharamukhaProvider{},
		Capabilities: []string{"transliteration"},
		Type:        common.TransliteratorType,
	}
	iuliiaEntry := common.ProviderEntry{
		Provider:     NewIuliiaProvider("rus"),
		Capabilities: []string{"transliteration"},
		Type:        common.TransliteratorType,
	}
	

	err := common.Register("mul", common.TokenizerType, "uniseg", unisegEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register uniseg provider: %v", err))
	}
	
	err = common.Register("mul", common.TransliteratorType, "aksharamukha", aksharamukhaEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register aksharamukha provider: %v", err))
	}
	
	err = common.Register("mul", common.TransliteratorType, "iuliia", iuliiaEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register iuliia provider: %v", err))
	}
	
	// #### Schemes registration ####

	for _, indicLang := range indicLangs {
		for _, scheme := range indicSchemes {
			scheme.Provider = "aksharamukha"
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
		scheme.Provider = "iuliia"
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