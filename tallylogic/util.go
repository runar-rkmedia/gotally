package tallylogic

func unique[T comparable](list []T) []T {
	m := map[T]struct{}{}
	for _, v := range list {
		m[v] = struct{}{}
	}
	uniq := make([]T, len(m))
	var i int
	for k := range m {
		uniq[i] = k
		i++
	}
	return uniq
}

func ReverseSlice[T comparable](s []T) []T {
	var r []T
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}
