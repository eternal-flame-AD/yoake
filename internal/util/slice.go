package util

// Contain checks if an element is in a slice.
func Contain[T comparable](a []T, x T) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Unique returns a slice with all duplicate elements removed.
func Unique[T comparable](a []T) []T {
	var result []T
	for _, e := range a {
		if !Contain(result, e) {
			result = append(result, e)
		}
	}
	return result
}

// AntiJOIN returns a slice with all elements in a that are not in b.
func AntiJoin[T comparable](a []T, b []T) []T {
	var result []T
	for _, e := range a {
		if !Contain(b, e) {
			result = append(result, e)
		}
	}
	return result
}

func Join[T any](vals ...[]T) []T {
	result := make([]T, 0)
	for _, val := range vals {
		result = append(result, val...)
	}
	return result
}

func Reverse[T any](a []T) []T {
	var result []T
	for i := len(a) - 1; i >= 0; i-- {
		result = append(result, a[i])
	}
	return result
}
