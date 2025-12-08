package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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

func StructToMap(data any) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // unexported field
			continue
		}
		result[field.Name] = val.Field(i).Interface()
	}

	return result
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

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Repeats the given slice for the given times
func RepeatSlice[T any](s []T, times int) []T {
	if times <= 0 || len(s) == 0 {
		return s
	}

	n := len(s)
	result := make([]T, n*times)
	copy(result, s)

	for copied := n; copied < len(result); copied *= 2 {
		copy(result[copied:], result[:copied])
	}

	return result
}

// Repeats each element for given times
func RepeatEach[T any](s []T, times int) []T {
	if times <= 0 || len(s) == 0 {
		return s
	}

	n := len(s)
	result := make([]T, 0, n*times)
	for _, v := range s {
		for i := 0; i < times; i++ {
			result = append(result, v)
		}
	}

	return result
}
