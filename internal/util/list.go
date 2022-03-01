package util

func Map[A any, B any](as []A, f func(A) B) []B {
	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = f(a)
	}
	return bs
}

func Filter[A any](as []A, f func(A) bool) []A {
	filtered := make([]A, 0)
	for _, a := range as {
		if f(a) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func Contains[A comparable](as []A, a0 A) bool {
	for _, a := range as {
		if a == a0 {
			return true
		}
	}
	return false
}
