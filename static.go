package translitkit

var LangsNeedTokenization = []string{
	"cmn", // Chinese (Mandarin) - 920 million
	"yue", // Chinese (Cantonese) - 85 million
	"jpn", // Japanese - 125 million
	// "kor", // Korean - 80 million - modern Korean typically uses spaces
	"wuu", // Chinese (Wu/Shanghainese) - 80 million
	"cjy", // Chinese (Jin) - 63 million
	"hsn", // Chinese (Xiang) - 36 million
	"mya", // Burmese - 33 million
	"tha", // Thai - 30 million
	"hak", // Chinese (Hakka) - 30 million
	"gan", // Chinese (Gan) - 31 million
	"nan", // Chinese (Min Nan/Southern Min) - 30 million
	"cdo", // Chinese (Min Dong/Eastern Min) - 9 million
	"khm", // Khmer - 16 million
	"mnp", // Chinese (Min Bei/Northern Min) - 10 million
	"lao", // Lao - 7 million
	"tib", // Tibetan - 1.2 million
}

var LangsNeedTransliteration = []string{
	"cmn", // Chinese (Mandarin) - 920 million
	"hin", // Hindi (Devanagari script) - 600 million
	"ara", // Arabic (Modern Standard) - 450 million
	"ben", // Bengali - 265 million
	"jpn", // Japanese - 125 million
	"rus", // Russian (Cyrillic script) - 150 million
	"kor", // Korean (Hangul) - 80 million
	"yue", // Chinese (Cantonese) - 85 million
	"wuu", // Chinese (Wu/Shanghainese) - 80 million
	"tel", // Telugu - 83 million
	"mar", // Marathi (Devanagari script) - 83 million
	"tam", // Tamil - 75 million
	"cjy", // Chinese (Jin) - 63 million
	"urd", // Urdu - 70 million
	"guj", // Gujarati - 55 million
	"kan", // Kannada - 45 million
	"ukr", // Ukrainian (Cyrillic script) - 45 million
	"pes", // Persian (Western/Iranian) - 45 million
	"prs", // Persian (Eastern/Dari) - 15 million
	"mal", // Malayalam - 38 million
	"hsn", // Chinese (Xiang) - 36 million
	"mya", // Burmese - 33 million
	"ori", // Oriya (Odia script) - 33 million
	"bho", // Bhojpuri (Devanagari script) - 33 million
	"gan", // Chinese (Gan) - 31 million
	"pan", // Punjabi (Gurmukhi script) - 30 million
	"hak", // Chinese (Hakka) - 30 million
	"nan", // Chinese (Min Nan) - 30 million
	"tha", // Thai - 30 million
	"amh", // Amharic (Ethiopic script) - 22 million
	"asm", // Assamese - 15 million
	"sin", // Sinhala - 15 million
	"khm", // Khmer - 16 million
	"heb", // Hebrew - 9 million
	"nep", // Nepali (Devanagari script) - 16 million
	"kaz", // Kazakh (Cyrillic script) - 13 million
	"cdo", // Chinese (Min Dong) - 9 million
	"bel", // Belarusian (Cyrillic script) - 9 million
	"mnp", // Chinese (Min Bei) - 10 million
	"lao", // Lao - 7 million
	"hye", // Armenian - 6.7 million
	"tgr", // Tigrinya (Ge'ez script) - 9 million
	"kat", // Georgian (Mkhedruli script) - 3.7 million
	"mon", // Mongolian (Traditional Mongolian script) - 5 million
	"uzb", // Uzbek (Cyrillic script) - 27 million
	"kir", // Kirghiz (Cyrillic script) - 4.5 million
	"tib", // Tibetan - 1.2 million
	"dzo", // Dzongkha (Tibetan script) - 640,000
	"san", // Sanskrit (Devanagari script)
	"grc", // Ancient Greek - (historical)
}
