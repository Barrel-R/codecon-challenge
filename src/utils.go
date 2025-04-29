package main

func Has[S ~[]T, T comparable](lst S, target T) bool {
	for _, val := range lst {
		if val == target {
			return true
		}
	}

	return false
}

func Filter[S ~[]T, T any](lst S, callback func(val T) bool) []T {
	var filteredArr []T

	for _, val := range lst {
		if callback(val) {
			filteredArr = append(filteredArr, val)
		}
	}

	return filteredArr
}
