- RENABLE LANG SPECIFIC METHODS (KANA PARTS)



- func (tokens TknSliceWrapper) Tokens() []AnyToken

- write tests

<hr>

- for dealing with arbitrary languages either github.com/pemistahl/lingua-go or aksharamukha script autodetection and redirect at correct language

*later:*

- add robpike/nihongo to force romanization after ichiran process
- Slice []AnyToken →→→ Sentences [][]AnyToken using uniseg.FirstSentenceInString
- Ideally: write a TokenWriter upon which multiple providers may write to in a row to enrich the linguistic annotation
