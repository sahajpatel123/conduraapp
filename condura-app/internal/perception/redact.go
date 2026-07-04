package perception

import "regexp"

// compileAndReplace is a tiny helper that compiles pat, replaces
// all matches in src with repl, and returns the result. Extracted
// so Redact can use a for loop without a closure per pattern.
//
// Implementation note: this lives in its own file so the regex
// package is imported only when Redact is called. Pure-data
// types in perception.go stay regex-free.
func compileAndReplace(pat, src, repl string, _ string) string {
	re, err := regexp.Compile(pat)
	if err != nil {
		// Pattern compile errors are programmer errors; skip
		// rather than panic so a broken pattern can't take down
		// the perception pipeline.
		return src
	}
	return re.ReplaceAllString(src, repl)
}
