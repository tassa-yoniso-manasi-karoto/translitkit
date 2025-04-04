package common

import (
	"fmt"
	"unicode"
	iso "github.com/barbashov/iso639-3"
)

var (
	stdLang2Ranges = make(map[string][]*unicode.RangeTable)
	
	// End punctuation (no space before these)
	endPunctuation = map[rune]bool{
		'.': true, ',': true, '!': true, '?': true, ':': true, ';': true, 
		')': true, ']': true, '}': true, '»': true, '…': true, '"': true, '\'': true,
		'」': true, '】': true, '）': true, '］': true, '｝': true, '』': true, '》': true, '〉': true,
		'。': true, '、': true, '：': true, '；': true, '，': true, '．': true, '！': true, '？': true,
	}
	
	// Opening punctuation (no space after these)
	openPunctuation = map[rune]bool{
		'(': true, '[': true, '{': true, '«': true, '"': true, '\'': true, 
		'「': true, '【': true, '（': true, '［': true, '『': true, '《': true, '〈': true,
	}
)

// GetUnicodeRangesFromLang returns the Unicode range tables that represent the primary
// writing scripts for the specified language.
// 
// The function accepts any valid ISO 639 language code (e.g. ISO 639-1, ISO 639-2, or ISO 639-3).
// 
// If the provided language code is not recognized or has no associated Unicode ranges, an error is returned.
func GetUnicodeRangesFromLang(lang string) ([]*unicode.RangeTable, error) {
	// If the map with standardized language codes hasn't been made yet, make it
	if len(stdLang2Ranges) == 0 {
		for origCode, ranges := range rawLang2Ranges {
			lang := iso.FromAnyCode(origCode)
			if lang == nil {
				continue
			}
			stdLang2Ranges[lang.Part3] = ranges
			delete(rawLang2Ranges, origCode)
		}
		
	}
	
	if obj := iso.FromAnyCode(lang); obj != nil {
		ranges, ok := stdLang2Ranges[obj.Part3]
		if !ok {
			return []*unicode.RangeTable{}, fmt.Errorf("'%s' has no range available", lang)
		}
		return ranges, nil
	}
	return []*unicode.RangeTable{}, fmt.Errorf("'%s' is not a valid ISO 639 language", lang)
}


// getScriptCategory determines which writing system a character belongs to
func getScriptCategory(r rune) string {
	switch {
	case unicode.Is(unicode.Han, r):
		return "Han" // Chinese characters (Hanzi, Kanji, Hanja)
	case unicode.Is(unicode.Hiragana, r):
		return "Hiragana"
	case unicode.Is(unicode.Katakana, r):
		return "Katakana"
	case unicode.Is(unicode.Hangul, r):
		return "Hangul" // Korean
	case unicode.Is(unicode.Thai, r):
		return "Thai"
	case unicode.Is(unicode.Lao, r):
		return "Lao"
	case unicode.Is(unicode.Khmer, r):
		return "Khmer"
	case unicode.Is(unicode.Myanmar, r):
		return "Myanmar" // Burmese
	case unicode.Is(unicode.Latin, r):
		return "Latin"
	case unicode.Is(unicode.Cyrillic, r):
		return "Cyrillic"
	case unicode.Is(unicode.Greek, r):
		return "Greek"
	case unicode.Is(unicode.Arabic, r):
		return "Arabic"
	case unicode.Is(unicode.Hebrew, r):
		return "Hebrew"
	case unicode.Is(unicode.Devanagari, r):
		return "Devanagari"
	case unicode.Is(unicode.Bengali, r):
		return "Bengali"
	case unicode.Is(unicode.Tamil, r):
		return "Tamil"
	case unicode.Is(unicode.Telugu, r):
		return "Telugu"
	case unicode.Is(unicode.Kannada, r):
		return "Kannada"
	case unicode.Is(unicode.Malayalam, r):
		return "Malayalam"
	case unicode.Is(unicode.Gujarati, r):
		return "Gujarati"
	case unicode.Is(unicode.Gurmukhi, r):
		return "Gurmukhi" // Punjabi
	default:
		return "Other"
	}
}

var langsNeedTokenization = []string{
	"zho", "cmn", // Chinese (Mandarin) - 920 million
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

var langsNeedTransliteration = []string{
	"zho", "cmn", // Chinese (Mandarin) - 920 million
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

var rawLang2Ranges = map[string][]*unicode.RangeTable{
	"abq": {unicode.Cyrillic},
	"ab":  {unicode.Cyrillic},
	"abr": {unicode.Latin},
	"ace": {unicode.Latin},
	"ach": {unicode.Latin},
	"ada": {unicode.Latin},
	"ady": {unicode.Cyrillic},
	"aa":  {unicode.Latin},
	"af":  {unicode.Latin},
	"agq": {unicode.Latin},
	"ain": {unicode.Katakana, unicode.Latin},
	"ak":  {unicode.Latin},
	"akk": {unicode.Cuneiform},
	"bss": {unicode.Latin},
	"akz": {unicode.Latin},
	"sq":  {unicode.Latin, unicode.Elbasan},
	"ale": {unicode.Latin},
	"arq": {unicode.Arabic},
	"am":  {unicode.Ethiopic},
	"amo": {unicode.Latin},
	"egy": {unicode.Egyptian_Hieroglyphs},
	"grc": {unicode.Greek, unicode.Cypriot, unicode.Linear_B},
	"xna": {unicode.Old_North_Arabian},
	"anp": {unicode.Devanagari},
	"blo": {unicode.Latin},
	"njo": {unicode.Latin},
	"ar":  {unicode.Arabic, unicode.Syriac},
	"an":  {unicode.Latin},
	"arc": {unicode.Imperial_Aramaic, unicode.Nabataean, unicode.Palmyrene},
	"aro": {unicode.Latin},
	"arp": {unicode.Latin},
	"arw": {unicode.Latin},
	"hy":  {unicode.Armenian},
	"rup": {unicode.Latin},
	"frp": {unicode.Latin},
	"as":  {unicode.Bengali},
	"aii": {unicode.Cyrillic, unicode.Syriac},
	"ast": {unicode.Latin},
	"asa": {unicode.Latin},
	"atj": {unicode.Latin},
	"cch": {unicode.Latin},
	"av":  {unicode.Cyrillic},
	"ae":  {unicode.Avestan},
	"awa": {unicode.Devanagari},
	"ay":  {unicode.Latin},
	"az":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"bfq": {unicode.Tamil},
	"ksf": {unicode.Latin},
	"bfd": {unicode.Latin},
	"bfy": {unicode.Devanagari},
	"bqi": {unicode.Arabic},
	"ban": {unicode.Balinese, unicode.Latin},
	"bgx": {unicode.Greek},
	"bft": {unicode.Arabic, unicode.Tibetan},
	"bal": {unicode.Arabic, unicode.Latin},
	"bm":  {unicode.Latin, unicode.Nko},
	"bax": {unicode.Bamum},
	"bn":  {unicode.Bengali},
	"bjn": {unicode.Latin},
	"bap": {unicode.Devanagari},
	"bci": {unicode.Latin},
	"bas": {unicode.Latin},
	"ba":  {unicode.Cyrillic},
	"eu":  {unicode.Latin},
	"bbc": {unicode.Batak, unicode.Latin},
	"btv": {unicode.Devanagari},
	"bar": {unicode.Latin},
	"bej": {unicode.Arabic},
	"be":  {unicode.Cyrillic},
	"bem": {unicode.Latin},
	"bez": {unicode.Latin},
	"bew": {unicode.Latin},
	"bhi": {unicode.Devanagari},
	"bhb": {unicode.Devanagari},
	"bho": {unicode.Devanagari},
	"bik": {unicode.Latin},
	"bin": {unicode.Latin},
	"bpy": {unicode.Bengali},
	"bi":  {unicode.Latin},
	"byn": {unicode.Ethiopic},
	"brx": {unicode.Devanagari},
	"bmq": {unicode.Latin},
	"bs":  {unicode.Cyrillic, unicode.Latin},
	"brh": {unicode.Arabic, unicode.Latin},
	"bra": {unicode.Devanagari},
	"br":  {unicode.Latin},
	"bvb": {unicode.Latin},
	"bug": {unicode.Latin, unicode.Buginese},
	"bku": {unicode.Latin, unicode.Buhid},
	"bg":  {unicode.Cyrillic},
	"bum": {unicode.Latin},
	"bua": {unicode.Cyrillic},
	"my":  {unicode.Myanmar},
	"buc": {unicode.Latin},
	"cad": {unicode.Latin},
	"frc": {unicode.Latin},
	"yue": {unicode.Han, unicode.Han}, // Simplified and Traditional
	"cps": {unicode.Latin},
	"xcr": {unicode.Carian},
	"car": {unicode.Latin},
	"ca":  {unicode.Latin},
	"cay": {unicode.Latin},
	"sef": {unicode.Latin},
	"ceb": {unicode.Latin},
	"tzm": {unicode.Latin, unicode.Tifinagh},
	"dtp": {unicode.Latin},
	"nch": {unicode.Latin},
	"ckb": {unicode.Arabic},
	"maz": {unicode.Latin},
	"ryu": {unicode.Katakana},
	"esu": {unicode.Latin},
	"fuq": {unicode.Latin},
	"ccp": {unicode.Bengali, unicode.Chakma},
	"ch":  {unicode.Latin},
	"ce":  {unicode.Cyrillic},
	"chr": {unicode.Cherokee},
	"chy": {unicode.Latin},
	"hne": {unicode.Devanagari},
	"cic": {unicode.Latin},
	"cgg": {unicode.Latin},
	"clc": {unicode.Latin},
	"qug": {unicode.Latin},
	"zh":  {unicode.Bopomofo, unicode.Latin, unicode.Han, unicode.Han, unicode.Phags_Pa}, // Simplified and Traditional
	"chn": {unicode.Latin},
	"chp": {unicode.Latin, unicode.Canadian_Aboriginal},
	"cho": {unicode.Latin},
	"ckt": {unicode.Cyrillic},
	"cu":  {unicode.Cyrillic},
	"chk": {unicode.Latin},
	"cv":  {unicode.Cyrillic},
	"myz": {unicode.Mandaic},
	"ksh": {unicode.Latin},
	"swb": {unicode.Arabic, unicode.Latin},
	"cop": {unicode.Arabic, unicode.Greek, unicode.Coptic},
	"kw":  {unicode.Latin},
	"co":  {unicode.Latin},
	"cr":  {unicode.Latin, unicode.Canadian_Aboriginal},
	"crh": {unicode.Cyrillic},
	"hr":  {unicode.Latin},
	"cs":  {unicode.Latin},
	"dak": {unicode.Latin},
	"dnj": {unicode.Latin},
	"thl": {unicode.Devanagari},
	"da":  {unicode.Latin},
	"dar": {unicode.Cyrillic},
	"dcc": {unicode.Arabic},
	"del": {unicode.Latin},
	"din": {unicode.Latin},
	"dv":  {unicode.Thaana},
	"doi": {unicode.Arabic, unicode.Devanagari, unicode.Takri},
	"dgr": {unicode.Latin},
	"rmt": {unicode.Arabic},
	"dty": {unicode.Devanagari},
	"dua": {unicode.Latin},
	"dng": {unicode.Cyrillic},
	"nl":  {unicode.Latin},
	"dyu": {unicode.Latin},
	"dz":  {unicode.Tibetan},
	"fud": {unicode.Latin},
	"cjm": {unicode.Arabic, unicode.Cham},
	"frs": {unicode.Latin},
	"nhe": {unicode.Latin},
	"eky": {unicode.Kayah_Li},
	"lwl": {unicode.Thai},
	"mgp": {unicode.Devanagari},
	"taj": {unicode.Devanagari, unicode.Tibetan},
	"efi": {unicode.Latin},
	"arz": {unicode.Arabic},
	"eka": {unicode.Latin},
	"ebu": {unicode.Latin},
	"egl": {unicode.Latin},
	"en":  {unicode.Latin, unicode.Deseret, unicode.Shavian},
	"myv": {unicode.Cyrillic},
	"eo":  {unicode.Latin},
	"et":  {unicode.Latin},
	"ett": {unicode.Latin, unicode.Old_Italic},
	"evn": {unicode.Cyrillic},
	"ee":  {unicode.Latin},
	"ewo": {unicode.Latin},
	"ext": {unicode.Latin},
	"fan": {unicode.Latin},
	"fo":  {unicode.Latin},
	"hif": {unicode.Devanagari, unicode.Latin},
	"fj":  {unicode.Latin},
	"fil": {unicode.Latin, unicode.Tagalog},
	"fi":  {unicode.Latin},
	"fon": {unicode.Latin},
	"gur": {unicode.Latin},
	"fr":  {unicode.Latin, unicode.Duployan},
	"fur": {unicode.Latin},
	"ff":  {unicode.Adlam, unicode.Latin},
	"fvr": {unicode.Latin},
	"gaa": {unicode.Latin},
	"gag": {unicode.Cyrillic, unicode.Latin},
	"gl":  {unicode.Latin},
	"gan": {unicode.Han}, // Simplified
	"lg":  {unicode.Latin},
	"gbm": {unicode.Devanagari},
	"grt": {unicode.Bengali},
	"gay": {unicode.Latin},
	"gba": {unicode.Latin},
	"gez": {unicode.Ethiopic},
	"ka":  {unicode.Georgian},
	"de":  {unicode.Latin, unicode.Runic},
	"aln": {unicode.Latin},
	"bbj": {unicode.Latin},
	"glk": {unicode.Arabic},
	"gil": {unicode.Latin},
	"gon": {unicode.Devanagari, unicode.Telugu},
	"gor": {unicode.Latin},
	"got": {unicode.Gothic},
	"grb": {unicode.Latin},
	"el":  {unicode.Greek},
	"gos": {unicode.Latin},
	"gub": {unicode.Latin},
	"gn":  {unicode.Latin},
	"gcr": {unicode.Latin},
	"gu":  {unicode.Gujarati},
	"gju": {unicode.Arabic},
	"gvr": {unicode.Devanagari},
	"guz": {unicode.Latin},
	"gwi": {unicode.Latin},
	"hoj": {unicode.Devanagari},
	"hai": {unicode.Latin},
	"ht":  {unicode.Latin},
	"hak": {unicode.Han}, // Simplified
	"hur": {unicode.Latin},
	"hnn": {unicode.Latin, unicode.Hanunoo},
	"bgc": {unicode.Devanagari},
	"ha":  {unicode.Arabic, unicode.Latin},
	"haw": {unicode.Latin},
	"haz": {unicode.Arabic},
	"he":  {unicode.Hebrew},
	"hz":  {unicode.Latin},
	"hil": {unicode.Latin},
	"hi":  {unicode.Devanagari, unicode.Latin, unicode.Mahajani},
	"ho":  {unicode.Latin},
	"hit": {unicode.Cuneiform},
	"hmn": {unicode.Latin, unicode.Pahawh_Hmong},
	"hnj": {unicode.Lao},
	"hoc": {unicode.Devanagari, unicode.Warang_Citi},
	"hop": {unicode.Latin},
	"hu":  {unicode.Latin},
	"hup": {unicode.Latin},
	"iba": {unicode.Latin},
	"ibb": {unicode.Latin},
	"is":  {unicode.Latin},
	"ife": {unicode.Latin},
	"ig":  {unicode.Latin},
	"ilo": {unicode.Latin},
	"smn": {unicode.Latin},
	"id":  {unicode.Arabic, unicode.Latin},
	"mvy": {unicode.Arabic},
	"izh": {unicode.Latin},
	"inh": {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"moe": {unicode.Latin},
	"ia":  {unicode.Latin},
	"ie":  {unicode.Latin},
	"iu":  {unicode.Latin, unicode.Canadian_Aboriginal},
	"ik":  {unicode.Latin},
	"ga":  {unicode.Latin},
	"it":  {unicode.Latin},
	"jam": {unicode.Latin},
	"ja":  {unicode.Hiragana, unicode.Katakana, unicode.Han}, // Japanese combines Hiragana, Katakana, and Han (Kanji)
	"jv":  {unicode.Javanese, unicode.Latin},
	"bze": {unicode.Latin},
	"kaj": {unicode.Latin},
	"dyo": {unicode.Arabic, unicode.Latin},
	"jrb": {unicode.Hebrew},
	"jpr": {unicode.Hebrew},
	"jml": {unicode.Devanagari},
	"jut": {unicode.Latin},
	"kbd": {unicode.Cyrillic},
	"kea": {unicode.Latin},
	"kab": {unicode.Latin},
	"kfr": {unicode.Devanagari},
	"gjk": {unicode.Arabic},
	"kac": {unicode.Latin},
	"kgp": {unicode.Latin},
	"kkj": {unicode.Latin},
	"kl":  {unicode.Latin},
	"kck": {unicode.Latin},
	"kln": {unicode.Latin},
	"xal": {unicode.Cyrillic},
	"rmf": {unicode.Latin},
	"kam": {unicode.Latin},
	"bjj": {unicode.Devanagari},
	"xnr": {unicode.Devanagari},
	"kn":  {unicode.Kannada},
	"kr":  {unicode.Latin},
	"kaa": {unicode.Cyrillic, unicode.Latin},
	"krc": {unicode.Cyrillic},
	"krl": {unicode.Latin},
	"ks":  {unicode.Arabic, unicode.Devanagari},
	"csb": {unicode.Latin},
	"tkt": {unicode.Devanagari},
	"kk":  {unicode.Arabic, unicode.Cyrillic},
	"kvr": {unicode.Latin},
	"bzx": {unicode.Latin},
	"kjh": {unicode.Cyrillic},
	"kht": {unicode.Myanmar},
	"khn": {unicode.Devanagari},
	"kca": {unicode.Cyrillic},
	"kha": {unicode.Bengali, unicode.Latin},
	"km":  {unicode.Khmer},
	"kjg": {unicode.Lao, unicode.Latin},
	"khw": {unicode.Arabic},
	"ki":  {unicode.Latin},
	"kmb": {unicode.Latin},
	"krj": {unicode.Latin},
	"rw":  {unicode.Latin},
	"kiu": {unicode.Latin},
	"mwk": {unicode.Latin},
	"thq": {unicode.Devanagari},
	"bkm": {unicode.Latin},
	"kge": {unicode.Latin},
	"kv":  {unicode.Cyrillic, unicode.Old_Permic},
	"koi": {unicode.Cyrillic},
	"kg":  {unicode.Latin},
	"kok": {unicode.Devanagari, unicode.Latin},
	"knn": {unicode.Devanagari},
	"ko":  {unicode.Hangul}, // Korean uses Hangul primarily
	"kfo": {unicode.Latin},
	"bqv": {unicode.Latin},
	"kpy": {unicode.Cyrillic},
	"kos": {unicode.Latin},
	"avk": {unicode.Latin},
	"khq": {unicode.Latin},
	"ses": {unicode.Latin},
	"kpe": {unicode.Latin},
	"kri": {unicode.Latin},
	"kj":  {unicode.Latin},
	"kfy": {unicode.Devanagari},
	"kum": {unicode.Cyrillic},
	"ku":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"kru": {unicode.Devanagari},
	"kut": {unicode.Latin},
	"kxv": {unicode.Devanagari, unicode.Latin, unicode.Oriya, unicode.Telugu},
	"kdt": {unicode.Thai},
	"kwk": {unicode.Latin},
	"nmg": {unicode.Latin},
	"ky":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"quc": {unicode.Latin},
	"lld": {unicode.Latin},
	"lad": {unicode.Hebrew},
	"lbe": {unicode.Cyrillic},
	"lki": {unicode.Arabic},
	"lkt": {unicode.Latin},
	"lam": {unicode.Latin},
	"lmn": {unicode.Telugu},
	"ljp": {unicode.Latin},
	"lag": {unicode.Latin},
	"laj": {unicode.Latin},
	"lo":  {unicode.Lao},
	"ltg": {unicode.Latin},
	"la":  {unicode.Latin},
	"lv":  {unicode.Latin},
	"lzz": {unicode.Georgian, unicode.Latin},
	"lep": {unicode.Lepcha},
	"lez": {unicode.Cyrillic, unicode.Caucasian_Albanian},
	"lij": {unicode.Latin},
	"lil": {unicode.Latin},
	"lif": {unicode.Devanagari, unicode.Limbu},
	"li":  {unicode.Latin},
	"lab": {unicode.Linear_A},
	"ln":  {unicode.Latin},
	"lfn": {unicode.Cyrillic, unicode.Latin},
	"lis": {unicode.Lisu},
	"lzh": {unicode.Han}, // Simplified
	"lt":  {unicode.Latin},
	"liv": {unicode.Latin},
	"lmo": {unicode.Latin},
	"ngl": {unicode.Latin},
	"nds": {unicode.Latin},
	"sli": {unicode.Latin},
	"dsb": {unicode.Latin},
	"loz": {unicode.Latin},
	"khb": {unicode.Tai_Tham},
	"lu":  {unicode.Latin},
	"lua": {unicode.Latin},
	"lui": {unicode.Latin},
	"smj": {unicode.Latin},
	"lun": {unicode.Latin},
	"luo": {unicode.Latin},
	"lut": {unicode.Latin},
	"lb":  {unicode.Latin},
	"luy": {unicode.Latin},
	"xlc": {unicode.Lycian},
	"xld": {unicode.Lydian},
	"ffm": {unicode.Latin},
	"mk":  {unicode.Cyrillic},
	"jmc": {unicode.Latin},
	"mad": {unicode.Latin},
	"maf": {unicode.Latin},
	"mag": {unicode.Devanagari},
	"mdh": {unicode.Latin},
	"vmf": {unicode.Latin},
	"mai": {unicode.Devanagari, unicode.Tirhuta},
	"mak": {unicode.Latin, unicode.Buginese},
	"vmw": {unicode.Latin},
	"mgh": {unicode.Latin},
	"kde": {unicode.Latin},
	"mg":  {unicode.Latin},
	"ms":  {unicode.Arabic, unicode.Latin},
	"ml":  {unicode.Malayalam},
	"pqm": {unicode.Latin},
	"mt":  {unicode.Latin},
	"mnc": {unicode.Mongolian},
	"mdr": {unicode.Latin, unicode.Buginese},
	"man": {unicode.Latin, unicode.Nko},
	"xmn": {unicode.Manichaean},
	"mni": {unicode.Bengali, unicode.Meetei_Mayek},
	"mns": {unicode.Cyrillic},
	"gv":  {unicode.Latin},
	"mxc": {unicode.Latin},
	"mi":  {unicode.Latin},
	"arn": {unicode.Latin},
	"mr":  {unicode.Devanagari, unicode.Modi},
	"chm": {unicode.Cyrillic},
	"mh":  {unicode.Latin},
	"mwr": {unicode.Devanagari},
	"myx": {unicode.Latin},
	"mas": {unicode.Latin},
	"mls": {unicode.Latin},
	"mzn": {unicode.Arabic},
	"mdt": {unicode.Latin},
	"mgy": {unicode.Latin},
	"byv": {unicode.Latin},
	"men": {unicode.Latin, unicode.Mende_Kikakui},
	"mwv": {unicode.Latin},
	"xmr": {unicode.Meroitic_Cursive},
	"mer": {unicode.Latin},
	"mgo": {unicode.Latin},
	"mtr": {unicode.Devanagari},
	"wtm": {unicode.Devanagari},
	"mic": {unicode.Latin},
	"crg": {unicode.Latin},
	"dum": {unicode.Latin},
	"enm": {unicode.Latin},
	"frm": {unicode.Latin},
	"gmh": {unicode.Latin},
	"nan": {unicode.Han}, // Simplified
	"min": {unicode.Latin},
	"xmf": {unicode.Georgian},
	"mwl": {unicode.Latin},
	"lus": {unicode.Bengali},
	"mhn": {unicode.Latin},
	"moh": {unicode.Latin},
	"mdf": {unicode.Cyrillic},
	"mnw": {unicode.Myanmar},
	"lol": {unicode.Latin},
	"mn":  {unicode.Cyrillic, unicode.Mongolian, unicode.Phags_Pa},
	"crm": {unicode.Canadian_Aboriginal},
	"mfe": {unicode.Latin},
	"ary": {unicode.Arabic},
	"mos": {unicode.Latin},
	"mro": {unicode.Latin},
	"unx": {unicode.Bengali, unicode.Devanagari},
	"mua": {unicode.Latin},
	"unr": {unicode.Bengali, unicode.Devanagari},
	"mus": {unicode.Latin},
	"ttt": {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"nqo": {unicode.Nko},
	"ars": {unicode.Arabic},
	"naq": {unicode.Latin},
	"gld": {unicode.Cyrillic},
	"nsk": {unicode.Latin, unicode.Canadian_Aboriginal},
	"na":  {unicode.Latin},
	"nv":  {unicode.Latin},
	"nxq": {unicode.Latin},
	"ndc": {unicode.Latin},
	"ng":  {unicode.Latin},
	"wni": {unicode.Arabic},
	"nap": {unicode.Latin},
	"zmi": {unicode.Latin},
	"yrk": {unicode.Cyrillic},
	"ne":  {unicode.Devanagari},
	"new": {unicode.Devanagari},
	"nij": {unicode.Latin},
	"zdj": {unicode.Arabic},
	"nnh": {unicode.Latin},
	"jgo": {unicode.Latin},
	"yrl": {unicode.Latin},
	"nia": {unicode.Latin},
	"fuv": {unicode.Latin},
	"pcm": {unicode.Latin},
	"noe": {unicode.Devanagari},
	"niu": {unicode.Latin},
	"fia": {unicode.Arabic},
	"nog": {unicode.Cyrillic},
	"nd":  {unicode.Latin},
	"scs": {unicode.Latin},
	"tts": {unicode.Thai},
	"crl": {unicode.Latin, unicode.Canadian_Aboriginal},
	"frr": {unicode.Latin},
	"hno": {unicode.Arabic},
	"kxm": {unicode.Thai},
	"lrc": {unicode.Arabic},
	"se":  {unicode.Cyrillic, unicode.Latin},
	"nso": {unicode.Latin},
	"nod": {unicode.Tai_Tham},
	"no":  {unicode.Latin},
	"nb":  {unicode.Latin},
	"nn":  {unicode.Latin},
	"nov": {unicode.Latin},
	"nus": {unicode.Latin},
	"nym": {unicode.Latin},
	"ny":  {unicode.Latin},
	"nyn": {unicode.Latin},
	"tog": {unicode.Latin},
	"nyo": {unicode.Latin},
	"nzi": {unicode.Latin},
	"ann": {unicode.Latin},
	"oc":  {unicode.Latin},
	"or":  {unicode.Oriya},
	"ojs": {unicode.Canadian_Aboriginal},
	"oj":  {unicode.Latin, unicode.Canadian_Aboriginal},
	"oka": {unicode.Latin},
	"ang": {unicode.Latin},
	"fro": {unicode.Latin},
	"goh": {unicode.Latin},
	"sga": {unicode.Latin, unicode.Ogham},
	"non": {unicode.Runic},
	"peo": {unicode.Old_Persian},
	"pro": {unicode.Latin},
	"otk": {unicode.Old_Turkic},
	"om":  {unicode.Ethiopic, unicode.Latin},
	"osa": {unicode.Latin, unicode.Osage},
	"osc": {unicode.Latin, unicode.Old_Italic},
	"os":  {unicode.Cyrillic},
	"pal": {unicode.Psalter_Pahlavi},
	"pfl": {unicode.Latin},
	"pau": {unicode.Latin},
	"pi":  {unicode.Devanagari, unicode.Sinhala, unicode.Thai},
	"pam": {unicode.Latin},
	"pag": {unicode.Latin},
	"pap": {unicode.Latin},
	"kvx": {unicode.Arabic},
	"prd": {unicode.Arabic},
	"xpr": {unicode.Inscriptional_Parthian},
	"ps":  {unicode.Arabic},
	"mfa": {unicode.Arabic},
	"pdc": {unicode.Latin},
	"fa":  {unicode.Arabic},
	"phn": {unicode.Phoenician},
	"pcd": {unicode.Latin},
	"pms": {unicode.Latin},
	"pis": {unicode.Latin},
	"crk": {unicode.Canadian_Aboriginal},
	"pdt": {unicode.Latin},
	"pon": {unicode.Latin},
	"pko": {unicode.Latin},
	"pl":  {unicode.Latin},
	"pnt": {unicode.Cyrillic, unicode.Greek, unicode.Latin},
	"pt":  {unicode.Latin},
	"prg": {unicode.Latin},
	"pa":  {unicode.Arabic, unicode.Gurmukhi},
	"puu": {unicode.Latin},
	"qu":  {unicode.Latin},
	"raj": {unicode.Devanagari},
	"rjs": {unicode.Devanagari},
	"thr": {unicode.Devanagari},
	"rkt": {unicode.Bengali},
	"rap": {unicode.Latin},
	"rar": {unicode.Latin},
	"rej": {unicode.Latin, unicode.Rejang},
	"rcf": {unicode.Latin},
	"ria": {unicode.Latin},
	"rif": {unicode.Latin, unicode.Tifinagh},
	"bto": {unicode.Latin},
	"rhg": {unicode.Arabic, unicode.Latin},
	"rgn": {unicode.Latin},
	"ro":  {unicode.Cyrillic, unicode.Latin},
	"rm":  {unicode.Latin},
	"rom": {unicode.Cyrillic, unicode.Latin},
	"rof": {unicode.Latin},
	"rng": {unicode.Latin},
	"rtm": {unicode.Latin},
	"rug": {unicode.Latin},
	"rn":  {unicode.Latin},
	"ru":  {unicode.Cyrillic},
	"rue": {unicode.Cyrillic},
	"rwk": {unicode.Latin},
	"xsa": {unicode.Old_South_Arabian},
	"sck": {unicode.Devanagari},
	"saf": {unicode.Latin},
	"ssy": {unicode.Latin},
	"smp": {unicode.Samaritan},
	"sam": {unicode.Hebrew, unicode.Samaritan},
	"saq": {unicode.Latin},
	"sm":  {unicode.Latin},
	"sgs": {unicode.Latin},
	"sad": {unicode.Latin},
	"sxn": {unicode.Latin},
	"sg":  {unicode.Latin},
	"sbp": {unicode.Latin},
	"sa":  {unicode.Devanagari, unicode.Sinhala, unicode.Grantha, unicode.Sharada, unicode.Siddham},
	"sat": {unicode.Bengali, unicode.Devanagari, unicode.Latin, unicode.Oriya, unicode.Ol_Chiki},
	"skr": {unicode.Arabic},
	"sc":  {unicode.Latin},
	"sas": {unicode.Latin},
	"sdc": {unicode.Latin},
	"stq": {unicode.Latin},
	"saz": {unicode.Saurashtra},
	"sco": {unicode.Latin},
	"gd":  {unicode.Latin},
	"syi": {unicode.Latin},
	"sly": {unicode.Latin},
	"sel": {unicode.Cyrillic},
	"seh": {unicode.Latin},
	"see": {unicode.Latin},
	"sr":  {unicode.Cyrillic, unicode.Latin},
	"srr": {unicode.Latin},
	"sei": {unicode.Latin},
	"crs": {unicode.Latin},
	"ksb": {unicode.Latin},
	"shn": {unicode.Myanmar},
	"swv": {unicode.Devanagari},
	"xsr": {unicode.Devanagari},
	"sn":  {unicode.Latin},
	"cjs": {unicode.Cyrillic},
	"ii":  {unicode.Latin, unicode.Yi},
	"scn": {unicode.Latin},
	"sid": {unicode.Latin},
	"bla": {unicode.Latin},
	"szl": {unicode.Latin},
	"sd":  {unicode.Arabic, unicode.Devanagari, unicode.Khojki, unicode.Khudawadi},
	"si":  {unicode.Sinhala},
	"rmo": {unicode.Latin},
	"srx": {unicode.Devanagari},
	"sms": {unicode.Latin},
	"den": {unicode.Latin, unicode.Canadian_Aboriginal},
	"sk":  {unicode.Latin},
	"sl":  {unicode.Latin},
	"xog": {unicode.Latin},
	"so":  {unicode.Arabic, unicode.Latin, unicode.Osmanya},
	"snk": {unicode.Latin},
	"srb": {unicode.Latin, unicode.Sora_Sompeng},
	"nr":  {unicode.Latin},
	"alt": {unicode.Cyrillic},
	"crj": {unicode.Latin, unicode.Canadian_Aboriginal},
	"hnd": {unicode.Arabic},
	"sdh": {unicode.Arabic},
	"luz": {unicode.Arabic},
	"sma": {unicode.Latin},
	"st":  {unicode.Latin},
	"sou": {unicode.Thai},
	"es":  {unicode.Latin},
	"srn": {unicode.Latin},
	"zgh": {unicode.Tifinagh},
	"suk": {unicode.Latin},
	"su":  {unicode.Latin, unicode.Sundanese},
	"sus": {unicode.Arabic, unicode.Latin},
	"swg": {unicode.Latin},
	"sw":  {unicode.Latin},
	"csw": {unicode.Canadian_Aboriginal},
	"ss":  {unicode.Latin},
	"sv":  {unicode.Latin},
	"gsw": {unicode.Latin},
	"syl": {unicode.Bengali, unicode.Syloti_Nagri},
	"syr": {unicode.Syriac},
	"tab": {unicode.Cyrillic},
	"shi": {unicode.Arabic, unicode.Latin, unicode.Tifinagh},
	"rob": {unicode.Latin},
	"tbw": {unicode.Latin, unicode.Tagbanwa},
	"ty":  {unicode.Latin},
	"blt": {unicode.Tai_Viet},
	"tdd": {unicode.Tai_Le},
	"dav": {unicode.Latin},
	"tg":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"tly": {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"tmh": {unicode.Latin},
	"ta":  {unicode.Tamil},
	"trv": {unicode.Latin},
	"twq": {unicode.Latin},
	"tt":  {unicode.Cyrillic},
	"tsg": {unicode.Latin},
	"rmu": {unicode.Latin},
	"ctd": {unicode.Latin},
	"te":  {unicode.Telugu},
	"ter": {unicode.Latin},
	"teo": {unicode.Latin},
	"tet": {unicode.Latin},
	"th":  {unicode.Thai},
	"tdh": {unicode.Devanagari},
	"bo":  {unicode.Tibetan},
	"tig": {unicode.Ethiopic},
	"ti":  {unicode.Ethiopic},
	"tem": {unicode.Latin},
	"tiv": {unicode.Latin},
	"tli": {unicode.Latin},
	"tpi": {unicode.Latin},
	"tkl": {unicode.Latin},
	"tok": {unicode.Latin},
	"lbw": {unicode.Latin},
	"dtm": {unicode.Latin},
	"to":  {unicode.Latin},
	"ttj": {unicode.Latin},
	"fit": {unicode.Latin},
	"trw": {unicode.Arabic},
	"tkr": {unicode.Cyrillic, unicode.Latin},
	"tsd": {unicode.Greek},
	"tsj": {unicode.Tibetan},
	"tsi": {unicode.Latin},
	"ts":  {unicode.Latin},
	"tn":  {unicode.Latin},
	"tcy": {unicode.Kannada},
	"tum": {unicode.Latin},
	"aeb": {unicode.Arabic},
	"tr":  {unicode.Arabic, unicode.Latin},
	"tk":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"tru": {unicode.Latin, unicode.Syriac},
	"tvl": {unicode.Latin},
	"tyv": {unicode.Cyrillic},
	"kcg": {unicode.Latin},
	"aoz": {unicode.Latin},
	"ude": {unicode.Cyrillic},
	"udm": {unicode.Cyrillic, unicode.Latin},
	"uga": {unicode.Ugaritic},
	"uk":  {unicode.Cyrillic},
	"uli": {unicode.Latin},
	"xum": {unicode.Latin, unicode.Old_Italic},
	"umb": {unicode.Latin},
	"hsb": {unicode.Latin},
	"ur":  {unicode.Arabic},
	"ug":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"uz":  {unicode.Arabic, unicode.Cyrillic, unicode.Latin},
	"vai": {unicode.Latin, unicode.Vai},
	"ve":  {unicode.Latin},
	"vec": {unicode.Latin},
	"vep": {unicode.Latin},
	"vi":  {unicode.Han, unicode.Latin},
	"vic": {unicode.Latin},
	"vo":  {unicode.Latin},
	"vro": {unicode.Latin},
	"vot": {unicode.Latin},
	"vun": {unicode.Latin},
	"wbq": {unicode.Telugu},
	"kxp": {unicode.Arabic},
	"wbr": {unicode.Devanagari},
	"wls": {unicode.Latin},
	"wa":  {unicode.Latin},
	"wae": {unicode.Latin},
	"war": {unicode.Latin},
	"wbp": {unicode.Latin},
	"was": {unicode.Latin},
	"guc": {unicode.Latin},
	"cy":  {unicode.Latin},
	"vls": {unicode.Latin},
	"bgn": {unicode.Arabic},
	"ikt": {unicode.Latin},
	"cja": {unicode.Arabic, unicode.Cham},
	"fy":  {unicode.Latin},
	"nhw": {unicode.Latin},
	"kyu": {unicode.Kayah_Li},
	"lcp": {unicode.Thai},
	"mrd": {unicode.Devanagari},
	"mrj": {unicode.Cyrillic},
	"lah": {unicode.Arabic},
	"tdg": {unicode.Devanagari, unicode.Tibetan},
	"wal": {unicode.Ethiopic},
	"wo":  {unicode.Arabic, unicode.Latin},
	"wuu": {unicode.Han}, // Simplified
	"kao": {unicode.Latin},
	"xav": {unicode.Latin},
	"xh":  {unicode.Latin},
	"hsn": {unicode.Han}, // Simplified
	"sah": {unicode.Cyrillic},
	"yav": {unicode.Latin},
	"yao": {unicode.Latin},
	"yap": {unicode.Latin},
	"ybb": {unicode.Latin},
	"yi":  {unicode.Hebrew},
	"yo":  {unicode.Latin},
	"yua": {unicode.Latin},
	"zag": {unicode.Latin},
	"zap": {unicode.Latin},
	"dje": {unicode.Latin},
	"zza": {unicode.Latin},
	"zea": {unicode.Latin},
	"zen": {unicode.Tifinagh},
	"za":  {unicode.Latin, unicode.Han}, // Simplified
	"gbz": {unicode.Arabic},
	"zu":  {unicode.Latin},
	"zun": {unicode.Latin},
}