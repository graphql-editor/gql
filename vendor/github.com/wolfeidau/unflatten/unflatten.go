package unflatten

import "strings"

// SplitByDot split the supplied keys by dot or fullstop character
func SplitByDot(k string) []string { return strings.Split(k, ".") }

// TokenizerFunc This function is used to tokenize the keys in the flattened data structure.
//
// The following example uses strings.Split to tokenize based on .
//  func(k string) []string { return strings.Split(k, ".") }
type TokenizerFunc func(string) []string

// Unflatten This function will unflatten a map with keys which are comprised of multiple tokens which
// are segmented by the tokenizer function.
func Unflatten(m map[string]interface{}, tf TokenizerFunc) map[string]interface{} {
	var tree = make(map[string]interface{})
	for k, v := range m {
		ks := tf(k)
		tr := tree
		for _, tk := range ks[:len(ks)-1] {
			trnew, ok := tr[tk]
			if !ok {
				trnew = make(map[string]interface{})
				tr[tk] = trnew
			}
			tr = trnew.(map[string]interface{})
		}
		tr[ks[len(ks)-1]] = v
	}
	return tree
}
