*urgent:*

- TH2EN: Use config to scrape & select desired romanization

- func (tokens TknSliceWrapper) Tokens() []AnyToken

- for TokenizedStr ideally place space between word-words not word-punctuation or punct-punct

- TH2EN test behavior of rod when BrowserAccessURL is invalid

- testing testing testing testing testing

<hr>

- mesure perf with pprof (impact of asserting every token in customModule notably)


*later:*

- add robpike/nihongo to force romanization after ichiran process
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation