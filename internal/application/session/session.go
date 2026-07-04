package session

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/analyzer"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/parser"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/constants"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/ui"
)

func Run() error {
	// получаем аргументы.
	args, err := ui.GetArgs()
	if err != nil {
		return Error{"ошибка с переданными аргументами", err}
	}

	// Получаем тип хранилища логов (сервер/локально), получаем все пути для поиска логов, т.к. мог быть указан шаблон.
	storageType, parsedPaths, err := parser.ParsePathTemplate(args.SourceTemplate)
	if err != nil {
		return Error{"ошибка парсера", err}
	}

	newAnalyzer := analyzer.Analyzer{
		StorageType: storageType,
		From:        args.From,
		To:          args.To,
		FilterType:  args.FilterType,
		FilterValue: args.FilterValue,
	}

	stats, err := newAnalyzer.Analyze(parsedPaths)
	if err != nil {
		return Error{"ошибка анализатора логов", err}
	}

	// Если --format не был указан, то статистика строится сразу в двух форматах.
	if args.Format == constants.Md || args.Format == "" {
		err = ui.WriteMD(parsedPaths, "info.md", stats)
		if err != nil {
			return Error{"ошибка записи в формат Markdown", err}
		}
	}

	if args.Format == constants.Adoc || args.Format == "" {
		err = ui.WriteADOC(parsedPaths, "info.adoc", stats)
		if err != nil {
			return Error{"ошибка записи в формат Adoc", err}
		}
	}

	return nil
}
