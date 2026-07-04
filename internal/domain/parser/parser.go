package parser

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/constants"
)

// Представление расшифрованного лога.
type LogInfo struct {
	RemoteAddr    string // ip с которого был сделан запрос.
	RemoteUser    string // пользователь, аутентифицированный через HTTP аутентификацию.
	TimeLocal     string // время посещения.
	Method        string // тип http запроса.
	RequestURL    string // запрошенный путь.
	HTTPVersion   string // версия HTTP.
	Status        string // статус ответа.
	BodyBytesSent string // размер ответа от сервера в байтах.
	HTTPReferer   string // реферал.
	HTTPUserAgent string // юзер-агент.
}

// Обрезает строку с индекса до первого пробела либо до конца.
func splitSpaces(logRunes []rune, index *int) string {
	result := ""

	for ; *index < len(logRunes); *index++ {
		if logRunes[*index] == ' ' {
			*index++
			break
		}

		result += string(logRunes[*index])
	}

	return result
}

// Обрезает строку начиная с индекса и до певого символа == first, до первого символа == second.
// Т.е. находит символы first, second, вырезает значение между ними.
func splitSymbols(logRunes []rune, index *int, first, second rune) string {
	firstIndex := 0
	secondIndex := 0

	for ; *index < len(logRunes); *index++ {
		if logRunes[*index] == second && firstIndex != 0 {
			secondIndex = *index
			*index += 2

			return string(logRunes[firstIndex+1 : secondIndex])
		}

		if logRunes[*index] == first {
			firstIndex = *index
		}
	}

	return ""
}

// Разбиваем request на 3 строки: метод, название ресурса, версия http.
func parseRequest(request string) []string {
	return strings.Split(request, " ")
}

// Парсит шаблон из --path. Определяет тип хранилища логов: локально, удаленно, все пути к логам в этом хранилище.
func ParsePathTemplate(path string) (storageType string, logPaths []string, err error) {
	// Попытка получить ответ от URL
	request, err := http.NewRequest("GET", path, http.NoBody)
	if err != nil {
		return "", nil, Error{"ошибка проверки типа хранилища"}
	}

	// Пробуем пингануть сервер. Если не получилось, то проверяем является ли --path локальным файлом.
	response, err := http.DefaultClient.Do(request)
	if err == nil {
		closeError := response.Body.Close()
		if closeError != nil {
			return "", nil,
				WrapError{"при проверке является ли заданный в аргументах путь корректным URL произошла ошибка", closeError}
		}
	}

	if err == nil && response.StatusCode == http.StatusOK {
		return constants.Remote, []string{path}, nil
	}

	// Попытка получить локальные логи.
	logPaths, err = filepath.Glob(path)
	if err != nil {
		return "", nil, WrapError{"не удалось распарсить шаблон-путь к логам", err}
	}

	// Если локальные файлы существует, то пробуем открыть первый, если он открывается, то считаем, что логи хранятся локально.
	if len(logPaths) > 0 {
		if _, err = os.Open(logPaths[0]); err == nil {
			return constants.Local, logPaths, nil
		}

		return "", nil, err
	}

	return "", nil, Error{"при парсинге не удалось определить тип хранилища логов"}
}

// Парсит строку с логом.
func ParseLogString(logString string) *LogInfo {
	var info LogInfo

	logRunes := []rune(logString)
	index := 0

	info.RemoteAddr = splitSpaces(logRunes, &index)

	// Пропускаем символ "-"
	index += 2

	info.RemoteUser = splitSpaces(logRunes, &index)

	info.TimeLocal = splitSymbols(logRunes, &index, '[', ']')

	// Внутри request содержится слайс вида [тип запроса, адрес запроса, версия http].
	request := parseRequest(splitSymbols(logRunes, &index, '"', '"'))

	info.Method, info.RequestURL, info.HTTPVersion = request[0], request[1], request[2]

	info.Status = splitSpaces(logRunes, &index)

	info.BodyBytesSent = splitSpaces(logRunes, &index)

	info.HTTPReferer = splitSymbols(logRunes, &index, '"', '"')

	info.HTTPUserAgent = splitSymbols(logRunes, &index, '"', '"')

	return &info
}
