package util

func Keys[k comparable, v any](m map[k]v) []k {
	results := make([]k, len(m))
	i := 0
	for k := range m {
		results[i] = k
		i++
	}
	return results
}
