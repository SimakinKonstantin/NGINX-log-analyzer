package statistics_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/statistics"
	"github.com/stretchr/testify/assert"
)

func Test_GetAvgResponseSize(t *testing.T) {
	type TestCase struct {
		name     string
		input    statistics.Statistics
		expected float64
	}

	TestCases := []TestCase{
		{name: "Среднее значение - целое число", input: statistics.Statistics{ResponseSizes: []int{4, 6, 10, 4}}, expected: 6.0},
		{name: "Среднее значение - нецелое число", input: statistics.Statistics{ResponseSizes: []int{5, 6}}, expected: 5.5},
		{name: "Среднее значение при одном элементе", input: statistics.Statistics{ResponseSizes: []int{19}}, expected: 19.0},
		{name: "Среднее значение при пустом множестве размеров ответов", input: statistics.Statistics{ResponseSizes: []int{}}, expected: 0.0},
	}

	for _, tc := range TestCases {
		assert.Equal(t, tc.expected, tc.input.GetAvgResponseSize(), tc.name)
	}
}

// Здесь тестируется получение определенного кол-ва популярных элементов на примере кодов ответа http.
// Т.к. для получения самых популярных кодов ответа и ресурсов используется одна и та же функция, тетсируется только с кодами.
func Test_GetPopularElements(t *testing.T) {
	type TestCase struct {
		name        string
		stats       statistics.Statistics
		codesCount  int // Кол-во запрашиваемых максимальных элементов.
		expectedLen int // Размер результирующей последовательности
	}

	TestCases := []TestCase{
		{name: "кол-во запрашиваемых кодов < всего в собранной статистике",
			stats:       statistics.Statistics{ResponsesCodes: map[string]int{"500": 1, "200": 3, "404": 7, "401": 5}},
			codesCount:  3,
			expectedLen: 3,
		},
		{name: "кол-во запрашиваемых кодов == всего в собранной статистике 1",
			stats:       statistics.Statistics{ResponsesCodes: map[string]int{"500": 1, "200": 7, "404": 1, "401": 5}},
			codesCount:  4,
			expectedLen: 4,
		},
		{name: "кол-во запрашиваемых кодов == всего в собранной статистике 2",
			stats:       statistics.Statistics{ResponsesCodes: map[string]int{"500": 11, "200": 11, "401": 11}},
			codesCount:  3,
			expectedLen: 3,
		},
		{name: "кол-во запрашиваемых кодов > всего в собранной статистике (в ответе не должно быть лишних пустых значений)",
			stats:       statistics.Statistics{ResponsesCodes: map[string]int{"500": 1, "200": 3, "404": 7, "401": 5}},
			codesCount:  5,
			expectedLen: 4,
		},
		{name: "все значения равны",
			stats:       statistics.Statistics{ResponsesCodes: map[string]int{"500": 5, "200": 5, "404": 5, "401": 5}},
			codesCount:  4,
			expectedLen: 4,
		},
	}

	// Функция проверки упорядоченности в результирующей последовательности.
	checkOrder := func(arr []statistics.Pair) bool {
		for i := 1; i < len(arr); i++ {
			if arr[i].Value > arr[i-1].Value {
				return false
			}
		}

		return true
	}

	for _, tc := range TestCases {
		result, _ := tc.stats.GetPopularElements(tc.codesCount, tc.stats.ResponsesCodes)

		// Проверка на правильное кол-во максимальных элементов.
		assert.Equal(t, tc.expectedLen, len(result), "Проверка длины: "+tc.name)

		// Проверка, что итоговый результат будет отсортирован в обратном порядке.
		assert.Equal(t, true, checkOrder(result), "Проверка упорядоченности: "+tc.name)
	}
}

func TestStatistics_GetPercentile(t *testing.T) {
	type TestCase struct {
		name     string
		input    statistics.Statistics
		expected int
		percent  int
	}

	TestCases := []TestCase{
		{name: "95% Процентиль",
			input: statistics.Statistics{ResponseSizes: []int{4, 6, 10, 4}}, expected: 10, percent: 95},
		{name: "95% Процентиль когда есть 0 значение",
			input: statistics.Statistics{ResponseSizes: []int{5, 6, 0, 3, 11, 8}}, expected: 11, percent: 95},
		{name: "50% Процентиль",
			input: statistics.Statistics{ResponseSizes: []int{19, 88, 123, 99, 6723, 22}}, expected: 88, percent: 50},
		{name: "1% Процентиль",
			input: statistics.Statistics{ResponseSizes: []int{8, 32, 54, 6, 76, 44}}, expected: 6, percent: 1},
		{name: "10% Процентиль",
			input: statistics.Statistics{ResponseSizes: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}}, expected: 2, percent: 10},
		{name: "0% Процентиль",
			input: statistics.Statistics{ResponseSizes: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}, expected: 0, percent: 0},
		{name: "Процентиль 30% при пустом слайсе",
			input: statistics.Statistics{ResponseSizes: []int{}}, expected: 0, percent: 30},
	}

	for _, tc := range TestCases {
		percentile, _ := tc.input.GetPercentile(tc.percent)
		assert.Equal(t, tc.expected, percentile, tc.name)
	}
}
