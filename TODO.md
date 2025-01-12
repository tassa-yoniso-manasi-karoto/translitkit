*urgent:*

- implement lang specific checks using static.go
- testing testing testing testing testing
- add basic usage to README

<hr>

- fix func (m *JapaneseModule) KanaParts


*later:*

- fix chunkify, make it default but preserve the non-chunkified way as well (maybe benchmark ideally)
- add robpike/nihongo to force romanization after ichiran process
- for TokenizedStr ideally place space between word-words not word-punctuation or punct-punct
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation