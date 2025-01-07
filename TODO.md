*urgent:*

- testing testing testing testing testing
- add basic usage to README

<hr>

- fix init() of pkg jpn which register ichiran provider that somehow get voided in main.go (maybe there is 2 separated translitkit instance due to the replace directive in go.mod which each a globalRegister)
- fix func (m *JapaneseModule) KanaParts
- somehow separate IchiranProvider's Docker init and and config init in Init()
- implement lang specific checks using data.go
- godocs


*later:*

- fix chunkify, make it default but preserve the non-chunkified way as well (maybe benchmark ideally)
- add robpike/nihongo to force romanization after ichiran process
- for TokenizedStr ideally place space between word-words not word-punctuation or punct-punct
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation