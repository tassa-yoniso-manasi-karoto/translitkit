### Status: pre-release [![Go Reference](https://pkg.go.dev/badge/github.com/tassa-yoniso-manasi-karoto/translitkit.svg)](https://pkg.go.dev/github.com/tassa-yoniso-manasi-karoto/translitkit) [![Go Report Card](https://goreportcard.com/badge/github.com/tassa-yoniso-manasi-karoto/translitkit)](https://goreportcard.com/report/github.com/tassa-yoniso-manasi-karoto/translitkit)

This library primarily aims at **providing linguistic tokenization** and **phonetically-accurate transliteration**.

Currently, neural network-oriented NLP tasks such as text classification, named entity recognition... are not supported because I don't have a need for it but I am open to iterating on the current implementation if people want to contribute. Hence, I will keep this as "pre-release" but the tokenization/romanization APIs should be pretty stable. 

I am not trying to reinvent the wheel therefore this library will leverage reputable implementations of (tokenizers+) romanizers for each language: either a go library or a ***dockerized component*** using the Docker Compose API (especially in the case of tokenizers or neural network based libraries).

**This library does not perform any kind of natural language processing by itself**, it only serves as a gateway to the underlying tokenizers/transliterators/NLP libs.

Thus there are linguistic annotations available following analysis such as part-of-speech tagging, lemmatization... **only if the underlying provider offers it and to the extent it provides it**.

> [!IMPORTANT]
> This library is ***not meant***:
> 	- ***to slugifiy or transform a string into ASCII*** (i.e. to generate URLs):<br>You may use [gosimple/slug](https://github.com/gosimple/slug) or [mozillazg/go-unidecode](https://github.com/mozillazg/go-unidecode) for these. go-unicode already ships a chinese tokenizer but other languages may benefit from using one of the tokenizer provided here though.
> 	- ***for Model-Ready tokenization for a transformer neural network***: you may use [sugarme/tokenizer](https://github.com/sugarme/tokenizer) for that.

## tldr

```go
import (
	"fmt"
	"github.com/tassa-yoniso-manasi-karoto/translitkit"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/lang/jpn"
)

const text = "日本語の例文です"

func main() {
	// Provides basic features for an arbitrary language of the ISO-639 set
	m, err := translitkit.DefaultModule("jpn")
	check(err)

	m.MustInit()
	defer m.Close()
	
	roman, err := m.Roman(text)
	check(err)

	// To access language-specific methods import the module of the dedicated pkg
	jm := jpn.DefaultModule()
	// No need to init/close again since it's the same underlying module

	kana, err := jm.KanaParts(text)
	check(err)

	fmt.Println(roman)
	fmt.Printf("%#v\n", kana)
}
```
### Output

```
nihongo no reibun desu
[]string{"にほんご", "の", "れいぶん", "です"}
```
See docs of sub package "common" for the basic methods set available across languages.


## Currently implemented tokenizers / transliterators

> [!NOTE]
> Traditional NLP providers such as morphological analyzers are prioritized over neural network-based libraries as they can be run on any hardware. See note below.

"combined" means the provider implement both.

### Japanese

- Ichiran [combined]

### Thai

 - thai2english.com scraper [combined] *(may be obsolete in the future but currently is the best option since the docker img of pythainlp is broken)*
 
## AI Doomer note
LLMs are perfectly suited for NLP.

Therefore all traditional NLP providers are on their way to becoming obsolete, included here under “traditional” are even the neural network libraries based on BERT and even those based on ELECTRA.

However, there is still a case for using traditional providers such as morphological analyzers:
1) They are resource-effective, affordable, anyone can run it without any expense or expensive hardware.
2) They have maturity and thorough testing/fine-tuning by humans.

Because I am writing this library to provide bulk/mass transliteration of subtitle files on my project langkit, LLMs are not suitable. However they already outperform traditional providers in some cases.


Besides that, this library was extensively written by Claude Sonnet 24.10, which authored most of the code except for the most elaborate parts of the design. For the ichiran.go bindings, which involves just JSON parsing, it has authored probably >90% of the code under minimum supervision, most of it being zero-shot generation.

The CEO of Anthropic said in a [very interesting interview](https://www.youtube.com/watch?v=ugvHCXCOmm4) back in November 24 there is no true roadblocks in sight and he speculates from the current curve AIs smarter than humans will exist within one or two years.

Dear professional programmers, you have a year or two before the job market becomes a sequel to Hunger Games. 🫡

