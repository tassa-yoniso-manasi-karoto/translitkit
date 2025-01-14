*urgent:*

- for TokenizedStr ideally place space between word-words not word-punctuation or punct-punct

- testing testing testing testing testing

<hr>


*later:*

- maybe implement jpn.NewModule
- fix chunkify, make it default but preserve the non-chunkified way as well (maybe benchmark ideally)
- add robpike/nihongo to force romanization after ichiran process
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation