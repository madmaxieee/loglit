package utils

func JoinSlices[T any](slices ...[]T) []T {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)

	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}

func Ptr[T any](v T) *T {
	return &v
}
