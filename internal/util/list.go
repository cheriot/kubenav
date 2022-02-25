package util

func Map[A any, B any](as []A, f func(A) B) []B {
	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = f(a)
	}
	return bs
}