*urgent:*

- **for TokenizedStr ideally place space between word-words not word-punctuation or punct-punct**

- get Aksharamukha docker outside of the client and into a pkg level var (docker client is very large and bloats debugging the GlobalRegister)

- ichiran: implement partial kana transliteration with user-defined upper limit based on Kanji frequency

- TH2EN: Use config to scrape & select desired romanization

- func (tokens TknSliceWrapper) Tokens() []AnyToken


- TH2EN test behavior of rod when BrowserAccessURL is invalid

- testing testing testing testing testing

<hr>

- mesure perf with pprof (impact of asserting every token in customModule notably)

- for dealing with arbitrary languages either github.com/pemistahl/lingua-go or aksharamukha script autodetection and redirect at correct language

*later:*

- Aksharamukha: somehow implement bulk transliteration (maybe with splitter)
- add robpike/nihongo to force romanization after ichiran process
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation