package ui

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/statistics"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/constants"
)

// Хранит информацию о расшифрованных аргументах командной строки.
type ParsedArgs struct {
	SourceTemplate string    // шаблон (glob) путь к файлам.
	From           time.Time // время начала логов для анализа.
	To             time.Time // время конца логов для анализа.
	Format         string    // формат файла-результата со статистикой.
	FilterType     string    // тип фильтра (по какому полю нужно проводить фильтрацию.
	FilterValue    string    // значение фильтра.
}

// При печати статистики в выходной файл названия файлов помечаются, как код. В Adoc и Md для этого используется один и
// тот же синтаксис. Эта функция вовзращает строку с путями, которая печатается и в Adoc, и в Md.
func getPathString(parsedPaths []string) string {
	result := "`"
	for i := 0; i < len(parsedPaths); i++ {
		result += parsedPaths[i]
		if i < len(parsedPaths)-1 {
			result += "`, `"
		} else {
			result += "`"
		}
	}

	return result
}

// Полчучает элементы из командной строки.
func GetArgs() (res ParsedArgs, err error) {
	ptrPath := flag.String("path", "", "Путь к логам")
	ptrFrom := flag.String("from", "", "С какого момента показывать логи (включительно)")
	ptrTo := flag.String("to", "", "До какого момента показывать логи (включительно)")
	ptrFormat := flag.String("format", "", fmt.Sprintf("Укажите формат вывода: %s или %s", constants.Md,
		constants.Adoc))
	ptrFilterType := flag.String("filter-type", "", fmt.Sprintf(`Фильтр для значений
Возможные типы фильтра:
"remote_addr" - ip с которого был сделан запрос;
"remote_user" - пользователь, аутентифицированный через HTTP аутентификацию;
"time_local" - время посещения в формате: %s;
"method" - тип HTTP-запроса;
"request_url" - адрес запрашиваемого ресурса;
"http_version" - версия HTTP;
"status" - статус ответа;
"body_bytes_sent" - размер ответа сервера в байтах;
"http_referer" - реферал;
"http_user_agent" - юзер-агент.`, constants.Layout))
	ptrFilterValue := flag.String("filter-value", "", "Значение фильтра")
	flag.Parse()

	// Проверка на обязательный параметр.
	if *ptrPath == "" {
		return res, ArgError{msg: "аргумент path не задан или пустой"}
	}

	res.SourceTemplate = *ptrPath

	res.From, err = time.Parse(constants.Layout, *ptrFrom)
	if err != nil && *ptrFrom != "" {
		return res, ArgWrapError{"вводимый формат времени должен соответствовать " + constants.Layout, err}
	}

	res.To, err = time.Parse(constants.Layout, *ptrTo)
	if err != nil && *ptrTo != "" {
		return res, ArgWrapError{"вводимый формат времени должен соответствовать " + constants.Layout, err}
	}

	res.Format = *ptrFormat
	if *ptrFormat != "" && *ptrFormat != constants.Md && *ptrFormat != constants.Adoc {
		return res, ArgError{msg: fmt.Sprintf("некорреткное значение --ptrFormat != %s или %s", constants.Md,
			constants.Adoc)}
	}

	res.FilterType, res.FilterValue = *ptrFilterType, *ptrFilterValue

	// filter-type, filter-value должны всегда быть оба указаны в аргументах или оба не указаны. Нельзя ввести
	// только один из этих аргументов - они связаны.
	if (*ptrFilterType != "" && *ptrFilterValue == "") || (*ptrFilterType == "" && *ptrFilterValue != "") {
		return res, ArgError{msg: "filter-type, filter-value можно использовать только вместе"}
	}

	return res, nil
}

const dirName = "results"

// Записывает статистику в формате md. parsedPaths - пути к анализируемым лог файлам, resName - имя выходного файла,
// info - записываемая статистика.
func WriteMD(parsedPaths []string, resName string, info *statistics.Statistics) error {
	// Формирование строки со всеми путями.
	sources := getPathString(parsedPaths)

	percentileValue95, err := info.GetPercentile(95)
	if err != nil {
		return FileWriteError{"ошибка при вычислени процентиля 95%", err}
	}

	percentileValue99, err := info.GetPercentile(99)
	if err != nil {
		return FileWriteError{"ошибка при вычислени процентиля 99%", err}
	}

	text := fmt.Sprintf(`### Статистика
|        Метрика        |   Значение   |
|:---------------------|:------------|
|        Источник       | %s    |
|     Кол-во запросов   | %d      |
| Средний размер ответа | %.3f     |
|     95%% Перцентиль размера в байтах   | %d   |
|     99%% Перцентиль размера в байтах   | %d |
| Максимальный размер ответа в байтах |%d |
`, sources, info.GetRequestCounter(), info.GetAvgResponseSize(), percentileValue95, percentileValue99, info.GetMaxResponseSize())

	text += "### Часто запрашиваемые ресурсы\n|        Ресурс         |  Количество  |\n|:---------------------|:------------|\n"

	popularResources, err := info.GetPopularElements(3, info.Resources)
	if err != nil {
		return FileWriteError{"ошибка при вычислении наиболее популярных ресурсов", err}
	}

	for i := range popularResources {
		text += fmt.Sprintf("|`%s`|%d|\n", popularResources[i].Object, popularResources[i].Value)
	}

	text += "### Частые коды ответов\n| Код |		Название	 |  Количество  |\n|:---|:----------------:|:------------|\n"

	popularCodes, err := info.GetPopularElements(3, info.ResponsesCodes)
	if err != nil {
		return FileWriteError{"ошибка при вычислении наиболее популярных кодов ответа", err}
	}

	for i := range popularCodes {
		intCode, err := strconv.Atoi(popularCodes[i].Object)
		if err != nil {
			return FileWriteError{"ошибка при разборе кода ответа", err}
		}

		codeDescr := http.StatusText(intCode)

		// Проверка, что код валидный.
		if codeDescr == "" {
			return FileWriteError{"ошибка - из статистики получен несуществующий код ответа", err}
		}

		text += fmt.Sprintf("|%d|%s|%d|\n", intCode, codeDescr, popularCodes[i].Value)
	}

	if err = os.Mkdir(dirName, 0o600); err != nil && !os.IsExist(err) {
		return FileWriteError{"ошибка создания директории для файла", err}
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", dirName, resName), []byte(text), 0o600)
	if err != nil {
		return FileWriteError{"ошибка записи файла", err}
	}

	return nil
}

// Записывает статистику в формате md. parsedPaths - пути к анализируемым лог файлам, resName - имя выходного файла,
// info - записываемая статистика.
func WriteADOC(parsedPaths []string, resName string, info *statistics.Statistics) error {
	// Формирование строки со всеми путями.
	sources := getPathString(parsedPaths)

	percentileValue95, err := info.GetPercentile(95)
	if err != nil {
		return FileWriteError{"ошибка при вычислени процентиля 95%", err}
	}

	percentileValue99, err := info.GetPercentile(99)
	if err != nil {
		return FileWriteError{"ошибка при вычислени процентиля 99%", err}
	}

	text := fmt.Sprintf(`=== Статистика
|===
|        Метрика        |   Значение
|        Источник       | %s    
|     Кол-во запросов   | %d
| Средний размер ответа | %.3f
|     95%% Перцентиль размера в байтах  | %d
|     99%% Перцентиль размера в байтах  | %d
| Максимальный размер ответа в байтах | %d
|===
`, sources, info.GetRequestCounter(), info.GetAvgResponseSize(), percentileValue95, percentileValue99, info.GetMaxResponseSize())

	text += "=== Часто запрашиваемые ресурсы\n|===\n|        Ресурс         |  Количество\n"

	popularResources, err := info.GetPopularElements(3, info.Resources)
	if err != nil {
		return FileWriteError{"ошибка при вычислении наиболее популярных ресурсов", err}
	}

	for i := range popularResources {
		text += fmt.Sprintf("|`%s`|%d\n", popularResources[i].Object, popularResources[i].Value)
	}

	text += "|===\n=== Частые коды ответов\n|===\n| Код |		Название	 |  Количество\n"

	popularCodes, err := info.GetPopularElements(3, info.ResponsesCodes)
	if err != nil {
		return FileWriteError{"ошибка при вычислении наиболее популярных кодов ответа", err}
	}

	for i := range popularCodes {
		intCode, err := strconv.Atoi(popularCodes[i].Object)
		if err != nil {
			return FileWriteError{"ошибка при разборе кода ответа", err}
		}

		codeDescr := http.StatusText(intCode)

		// Проверка, что код валидный.
		if codeDescr == "" {
			return FileWriteError{"ошибка - из статистики получен несуществующий код ответа", err}
		}

		text += fmt.Sprintf("|%d|%s|%d\n", intCode, codeDescr, popularCodes[i].Value)
	}

	text += "|==="

	if err = os.Mkdir(dirName, 0o600); err != nil && !os.IsExist(err) {
		return FileWriteError{"ошибка создания директории для файла", err}
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", dirName, resName), []byte(text), 0o600)
	if err != nil {
		return FileWriteError{"ошибка записи файла", err}
	}

	return nil
}
