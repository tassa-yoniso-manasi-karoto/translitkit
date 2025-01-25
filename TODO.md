*urgent:*

- TokenizedStr & TokenizedParts: **currently some tokenizers return spaces and some don't → Find a solution**

- **ichiran: implement PARTIAL KANA TRANSLITERATION** with user-defined upper limit based on Kanji frequency

- TH2EN:
  - Use config to scrape & select desired romanization
  - test behavior of rod when BrowserAccessURL is invalid

- func (tokens TknSliceWrapper) Tokens() []AnyToken

- write tests

<hr>

- mesure perf with pprof (impact of asserting every token in customModule notably)

- for dealing with arbitrary languages either github.com/pemistahl/lingua-go or aksharamukha script autodetection and redirect at correct language

*later:*

- Aksharamukha: somehow implement bulk transliteration (maybe with splitter)
- add robpike/nihongo to force romanization after ichiran process
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation
