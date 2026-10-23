package utils

func Map[T any, Z any](array []T, mapFunc func(T) Z) []Z {
	result := make([]Z, 0, len(array))
	for _, item := range array {
		result = append(result, mapFunc(item))
	}
	return result
}
