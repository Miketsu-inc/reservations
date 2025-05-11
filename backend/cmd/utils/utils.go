package utils

import (
	"fmt"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

// Convert Go []int slice to PostgreSQL array string
func IntSliceToPgArray(ints []int) string {
	if len(ints) == 0 {
		return "{}"
	}

	strInts := make([]string, len(ints))
	for i, num := range ints {
		strInts[i] = fmt.Sprintf("%d", num)
	}

	return "{" + strings.Join(strInts, ",") + "}"
}

// Convert Go []string slice to PostgreSQL array string
func TimeStringToPgArray(strs []string) string {
	if len(strs) == 0 {
		return "{}"
	}

	strValues := make([]string, len(strs))
	for i, str := range strs {
		strValues[i] = fmt.Sprintf("'%s'", str) // Wrap each string in single quotes
	}

	return "{" + strings.Join(strValues, ",") + "}" // Join them with commas and wrap in {}
}

// convert the PostgresSql array into []int  (array format: {1, NULL, 3, NULL, 5} )
func ParsePgArrayToInt(arrayStr string) ([]int, error) {
	if arrayStr == "NULL" || arrayStr == "{}" {
		return []int{}, nil
	}

	trimmed := arrayStr[1 : len(arrayStr)-1]
	elements := strings.Split(trimmed, ",")

	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if elem == "NULL" {
			continue
		}
		val, err := strconv.Atoi(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

func CalculatePercentChange(previous int, current int) int {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return int(float64(current-previous) / float64(previous) * 100)
}
