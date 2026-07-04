package statistics

import (
	"fmt"
	"math"
	"slices"
	"sort"
)

// Собранная статистика из логов.
type Statistics struct {
	RequestCounter int            // кол-во запросов.
	Resources      map[string]int // ключ - ресурс, значение - кол-во обращений.
	ResponsesCodes map[string]int // ключ - код ответа, значение - кол-во.
	ResponseSizes  []int          // размер ответов в байтах. Используется слайс для дальнейшего вычисления процентиля.
}

// Пара элементов.
type Pair struct {
	Object string
	Value  int
}

// Делает сдвиг элементо вправо. При сдвиге пропадает последний элемент. index - индекс с которого будет начинаться сдвиг.
func slide(values []Pair, index int) {
	cur := values[index]

	for i := index + 1; i < len(values); i++ {
		values[i], cur = cur, values[i]
	}
}

func NewStatistics() *Statistics {
	return &Statistics{
		RequestCounter: 0,
		Resources:      make(map[string]int),
		ResponsesCodes: make(map[string]int),
		ResponseSizes:  make([]int, 0),
	}
}

// Возвращает кол-во запросов.
func (s *Statistics) GetRequestCounter() int {
	return s.RequestCounter
}

// Возвращает кол-во популярных элементов. maxCount - кол-во запрашиваемых элементов. elements - мапа, откуда будут
// выбираться самые популярные элементы. Возвращает слайс элементов вида: {элемент : кол-во}. Этот слайс упорядочен по "кол-во".
func (s *Statistics) GetPopularElements(maxCount int, elements map[string]int) ([]Pair, error) {
	if maxCount < 0 {
		return nil, Error{fmt.Sprintf("введено некорректное значение maxCount: %d", maxCount)}
	}

	// Чтобы не обрзовывались пустые лишние элементы
	// массива.
	var maxValues []Pair
	if maxCount < len(elements) {
		maxValues = make([]Pair, maxCount)
	} else {
		maxValues = make([]Pair, len(elements))
	}

	for key, value := range elements {
		for i := 0; i < maxCount; i++ {
			if value > maxValues[i].Value {
				slide(maxValues, i)
				maxValues[i].Object = key
				maxValues[i].Value = value

				break
			}
		}
	}

	return maxValues, nil
}

// Возвращает средний размер ответа в байтах.
func (s *Statistics) GetAvgResponseSize() float64 {
	if len(s.ResponseSizes) == 0 {
		return 0.0
	}

	totalSize := 0
	for _, i := range s.ResponseSizes {
		totalSize += i
	}

	return float64(totalSize) / float64(len(s.ResponseSizes))
}

// Возвращает заданный процентиль.
func (s *Statistics) GetPercentile(percent int) (int, error) {
	if percent < 0 || percent > 100 {
		return -1, Error{"значение процента должно быть от 0 до 100"}
	}

	if len(s.ResponseSizes) == 0 || percent == 0 {
		return 0, nil
	}

	numbersCount := int(math.Ceil(float64(percent)*0.01*float64(len(s.ResponseSizes)))) - 1

	sort.Slice(s.ResponseSizes, func(i, j int) bool {
		return s.ResponseSizes[i] < s.ResponseSizes[j]
	})

	return s.ResponseSizes[numbersCount], nil
}

// Возвращает максимальный размер ответа. Можно было бы использовать 100% процентиль, но тут лучше ассимптотика.
func (s *Statistics) GetMaxResponseSize() int {
	if len(s.ResponseSizes) == 0 {
		return 0
	}

	return slices.Max(s.ResponseSizes)
}
