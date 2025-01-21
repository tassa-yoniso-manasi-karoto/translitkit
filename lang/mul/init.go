
package mul

import (
	"fmt"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const Lang = "mul"

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
}