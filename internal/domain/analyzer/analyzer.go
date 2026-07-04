package analyzer

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/parser"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/statistics"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/constants"
)

// Analyzer занимается анализом логов из выбранного источника.
type Analyzer struct {
	StorageType string    // тип хранилища локально, сервер
	From        time.Time // значение аргумента --From
	To          time.Time // значение аргумента --To
	FilterType  string    // значение аргумента --filter-type
	FilterValue string    // значение аргумента --filter-value
}

// Функция анализа логов, которые хрянятся по пути parsedPath.
func (analyzer *Analyzer) analyzeLog(parsedPath string, stat *statistics.Statistics) (err error) {
	// Открываем хранилище логов на чтение.
	reader, err := analyzer.getReader(parsedPath, analyzer.StorageType)
	if err != nil {
		return WrapError{"ошибка чтения лога", err}
	}

	defer func() {
		// Проверяем, если Close() сработало с ошибкой, то перезаписываем err. Иначе err остается как был.
		if CloseErr := reader.Close(); CloseErr != nil {
			err = WrapError{"ошибка работы с файлом", CloseErr}
		}
	}()

	// Получаем один из путей для логов и начинаем читать от туда построчно.
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		logLine := parser.ParseLogString(scanner.Text())
		// Проверка на фильтры, если они были указаны.
		if analyzer.FilterType != "" {
			analyzeLog, err := analyzer.checkFilter(analyzer.FilterType, analyzer.FilterValue, logLine)
			if err != nil {
				return WrapError{"ошибка при работе с фильтром --filter-type", err}
			}

			// Если текущий лог не проходит фильтры, переходим к следующему.
			if !analyzeLog {
				continue
			}
		}

		parsedTime, err := time.Parse(constants.Layout, logLine.TimeLocal)
		if err != nil {
			return WrapError{"ошибка парсинга времени из лога", err}
		}

		// Если были введены фильтры по времени, текущее значение не попадает в отрезок времени.
		if (!analyzer.From.IsZero() && parsedTime.Before(analyzer.From)) ||
			(!analyzer.To.IsZero() && parsedTime.After(analyzer.To)) {
			continue
		}

		stat.RequestCounter++
		stat.ResponsesCodes[logLine.Status]++
		stat.Resources[logLine.RequestURL]++

		respSize, err := strconv.Atoi(logLine.BodyBytesSent)
		if err != nil {
			return WrapError{"ошибка при обработке размера ответа из лога", err}
		}

		stat.ResponseSizes = append(stat.ResponseSizes, respSize)
	}

	return err
}

// Получает reader, из которого будут читаться логи.
func (*Analyzer) getReader(parsedPath, storageType string) (io.ReadCloser, error) {
	// Получаем ответ от сервера и отправляем body ответа как reader.
	// Если возникла какая-то ошибка, то закрываем body, отправялем error.
	if storageType == constants.Remote {
		request, err := http.NewRequest("GET", parsedPath, http.NoBody)
		if err != nil {
			return nil, WrapError{"ошибка получения логов с сервера", err}
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, WrapError{"ошибка получения логов с сервера", err}
		}

		if err == nil && response.StatusCode == http.StatusOK {
			return response.Body, nil
		}

		CloseErr := response.Body.Close()
		if CloseErr != nil {
			err = WrapError{"ошибка определения типа хранилища логов", CloseErr}
		}

		return nil, err
	}

	if storageType == constants.Local {
		result, err := os.Open(parsedPath)
		if err == nil {
			return result, nil
		}

		CloseErr := result.Close()
		if CloseErr != nil {
			err = WrapError{"ошибка определения типа хранилища логов", CloseErr}
		}

		return nil, err
	}

	return nil, Error{"неподдерживаемый тип хранилища логов"}
}

// Проверяет соответствие заданного фильтра и значения текущего лога. Возвращает true, если значение соответствует
// запрашиваемому в фильтре.
func (*Analyzer) checkFilter(fieldName, value string, logLine *parser.LogInfo) (bool, error) {
	switch fieldName {
	case "remote_addr":
		return logLine.RemoteAddr == value, nil
	case "remote_user":
		return logLine.RemoteUser == value, nil
	case "time_local":
		return logLine.TimeLocal == value, nil
	case "method":
		return logLine.Method == value, nil
	case "request_url":
		return logLine.RequestURL == value, nil
	case "http_version":
		return logLine.HTTPVersion == value, nil
	case "status":
		return logLine.Status == value, nil
	case "body_bytes_sent":
		return logLine.BodyBytesSent == value, nil
	case "http_referer":
		return logLine.HTTPReferer == value, nil
	case "http_user_agent":
		return logLine.HTTPUserAgent == value, nil
	default:
		return false, Error{"неверный тип фильтра"}
	}
}

// Функция анализа логов из всех путей.
func (analyzer *Analyzer) Analyze(parsedPaths []string) (*statistics.Statistics, error) {
	stat := statistics.NewStatistics()

	for _, i := range parsedPaths {
		// На каждой итерации обрабатывается соответствующий путь к логам.
		err := analyzer.analyzeLog(i, stat)
		if err != nil {
			return nil, err
		}
	}

	return stat, nil
}
