
package mul

import (
	"github.com/tassa-yoniso-manasi-karoto/go-aksharamukha"
	iuliia "github.com/mehanizm/iuliia-go"
	
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

var indicSchemes = []common.TranslitScheme{
	{ Name: "Harvard-Kyoto", Description: "Harvard-Kyoto romanization system"},
	{ Name: "IAST", Description: "International Alphabet of Sanskrit Transliteration"},
	{ Name: "ITRANS", Description: "Indian languages TRANSliteration"},
	{ Name: "Velthuis", Description: "Velthuis transliteration system"},
	{ Name: "ISO", Description: "ISO 15919 transliteration standard"},
	{ Name: "Titus", Description: "TITUS transliteration system"},
	{ Name: "SLP1", Description: "Sanskrit Library Protocol 1"},
	{ Name: "WX", Description: "WX notation system"},
	{ Name: "Roman-Readable", Description: "Simplified readable romanization"},
	{ Name: "Roman-Colloquial", Description: "Colloquial romanization style"},
}

var indicSchemesToScript = map[string]aksharamukha.Script{
	"Harvard-Kyoto":    aksharamukha.HK,
	"IAST":             aksharamukha.IAST,
	"ITRANS":           aksharamukha.Itrans,
	"Velthuis":         aksharamukha.Velthuis,
	"ISO":              aksharamukha.ISO,
	"Titus":            aksharamukha.Titus,
	"SLP1":             aksharamukha.SLP1,
	"WX":               aksharamukha.WX,
	"Roman-Readable":   aksharamukha.RomanReadable,
	"Roman-Colloquial": aksharamukha.RomanColloquial,
}

var russianSchemes = []common.TranslitScheme{
	{ Name: "ala_lc", Description: "American Library Association - Library of Congress"},
	{ Name: "ala_lc_alt", Description: "American Library Association - Library of Congress (Alternative)"},
	{ Name: "bgn_pcgn", Description: "Board on Geographic Names - Permanent Committee on Geographical Names"},
	{ Name: "bgn_pcgn_alt", Description: "Board on Geographic Names - Permanent Committee on Geographical Names (Alternative)"},
	{ Name: "bs_2979", Description: "British Standard 2979:1958 - Romanization of Cyrillic and Greek Scripts"},
	{ Name: "bs_2979_alt", Description: "British Standard 2979:1958 - Romanization of Cyrillic and Greek Scripts (Alternative)"},
	{ Name: "gost_16876", Description: "GOST 16876-71 - Russian National Standard for Transliteration of Cyrillic Characters"},
	{ Name: "gost_16876_alt", Description: "GOST 16876-71 - Russian National Standard for Transliteration of Cyrillic Characters (Alternative)"},
	{ Name: "gost_52290", Description: "GOST R 52290-2004 - Russian National Standard for Transliteration of Cyrillic Characters"},
	{ Name: "gost_52535", Description: "GOST R 52535.1-2006 - Russian National Standard for Transliteration of Cyrillic Characters"},
	{ Name: "gost_7034", Description: "GOST R 7.0.34-2014 - Russian National Standard for Transliteration of Cyrillic Characters"},
	{ Name: "gost_779", Description: "GOST 7.79-2000 - Russian National Standard for Transliteration of Cyrillic Characters (ISO 9:1995 equivalent)"},
	{ Name: "gost_779_alt", Description: "GOST 7.79-2000 - Russian National Standard for Transliteration of Cyrillic Characters (ISO 9:1995 equivalent, Alternative)"},
	{ Name: "icao_doc_9303", Description: "International Civil Aviation Organization Document 9303 - Machine Readable Travel Documents"},
	{ Name: "iso_9_1954", Description: "ISO/R 9:1954 - International Standard for Transliteration of Cyrillic Characters"},
	{ Name: "iso_9_1968", Description: "ISO/R 9:1968 - International Standard for Transliteration of Cyrillic Characters"},
	{ Name: "iso_9_1968_alt", Description: "ISO/R 9:1968 - International Standard for Transliteration of Cyrillic Characters (Alternative)"},
	{ Name: "mosmetro", Description: "Moscow Metro Map Transliteration Scheme"},
	{ Name: "mvd_310", Description: "MVD 310-1997 - Russian Ministry of Internal Affairs Transliteration Standard"},
	{ Name: "mvd_310_fr", Description: "MVD 310-1997 - Russian Ministry of Internal Affairs Transliteration Standard (French variant)"},
	{ Name: "mvd_782", Description: "MVD 782-2000 - Russian Ministry of Internal Affairs Transliteration Standard"},
	{ Name: "scientific", Description: "Scientific Transliteration Scheme (International System of Transliteration)"},
	{ Name: "telegram", Description: "Telegram Transliteration Scheme"},
	{ Name: "ungegn_1987", Description: "United Nations Group of Experts on Geographical Names 1987 - Romanization System"},
	{ Name: "wikipedia", Description: "Wikipedia Transliteration Scheme"},
	{ Name: "yandex_maps", Description: "Yandex Maps Transliteration Scheme"},
	{ Name: "yandex_money", Description: "Yandex Money Transliteration Scheme"},
}


var russianSchemesToScript = map[string]*iuliia.Schema{
	"ala_lc":         iuliia.Ala_lc,
	"ala_lc_alt":     iuliia.Ala_lc_alt,
	"bgn_pcgn":       iuliia.Bgn_pcgn,
	"bgn_pcgn_alt":   iuliia.Bgn_pcgn_alt,
	"bs_2979":        iuliia.Bs_2979,
	"bs_2979_alt":    iuliia.Bs_2979_alt,
	"gost_16876":     iuliia.Gost_16876,
	"gost_16876_alt": iuliia.Gost_16876_alt,
	"gost_52290":     iuliia.Gost_52290,
	"gost_52535":     iuliia.Gost_52535,
	"gost_7034":      iuliia.Gost_7034,
	"gost_779":       iuliia.Gost_779,
	"gost_779_alt":   iuliia.Gost_779_alt,
	"icao_doc_9303":  iuliia.Icao_doc_9303,
	"iso_9_1954":     iuliia.Iso_9_1954,
	"iso_9_1968":     iuliia.Iso_9_1968,
	"iso_9_1968_alt": iuliia.Iso_9_1968_alt,
	"mosmetro":       iuliia.Mosmetro,
	"mvd_310":        iuliia.Mvd_310,
	"mvd_310_fr":     iuliia.Mvd_310_fr,
	"mvd_782":        iuliia.Mvd_782,
	"scientific":     iuliia.Scientific,
	"telegram":       iuliia.Telegram,
	"ungegn_1987":    iuliia.Ungegn_1987,
	"wikipedia":      iuliia.Wikipedia,
	"yandex_maps":    iuliia.Yandex_maps,
	"yandex_money":   iuliia.Yandex_money,
}


var uzbekScheme = common.TranslitScheme{ Name: "uz", Description: "Uzbekistan cyr-lat transliteration schema", Provider: "iuliia" }

