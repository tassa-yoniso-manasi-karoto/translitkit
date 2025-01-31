
package mul

import (
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

var indicSchemes = []common.TranslitScheme{
	{
		Name:        "Harvard-Kyoto",
		Description: "Harvard-Kyoto romanization system",
	},
	{
		Name:        "IAST",
		Description: "International Alphabet of Sanskrit Transliteration",
	},
	{
		Name:        "ITRANS",
		Description: "Indian languages TRANSliteration",
	},
	{
		Name:        "Velthuis",
		Description: "Velthuis transliteration system",
	},
	{
		Name:        "ISO",
		Description: "ISO 15919 transliteration standard",
	},
	{
		Name:        "Titus",
		Description: "TITUS transliteration system",
	},
	{
		Name:        "SLP1",
		Description: "Sanskrit Library Protocol 1",
	},
	{
		Name:        "WX",
		Description: "WX notation system",
	},
	{
		Name:        "Roman-Readable",
		Description: "Simplified readable romanization",
	},
	{
		Name:        "Roman-Colloquial",
		Description: "Colloquial romanization style",
	},
}