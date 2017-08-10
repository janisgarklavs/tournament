package main

import "strconv"

func splitEvenly(amount, parts int) []int {
	var result []int
	evenValues := amount / parts
	for i := 0; i < parts; i++ {
		result = append(result, evenValues)
	}
	remainder := amount % parts
	for i := 0; i < remainder; i++ {
		result[i]++
	}
	return result
}

func getPointsFromString(input string) (int, error) {
	points, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	points *= 100
	return int(points), nil
}
