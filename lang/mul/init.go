
package mul

import (
	"fmt"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)


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

	err := common.Register("mul", common.TokenizerType, "uniseg", unisegEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register uniseg provider: %v", err))
	}
	err = common.Register("mul", common.TransliteratorType, "aksharamukha", aksharamukhaEntry)
	if err != nil {
		panic(fmt.Sprintf("failed to register aksharamukha provider: %v", err))
	}

	// Set as default transliterator along with uniseg tokenizer
	// TODO for this leverage aksharamukha script autodetection
	// defaultProviders := []common.ProviderEntry{
	// 	unisegEntry,
	// 	aksharamukhaEntry,
	// }

	// err = common.SetDefault("mul", defaultProviders)
	// if err != nil {
	// 	panic(fmt.Sprintf("failed to set default providers: %v", err))
	// }
}