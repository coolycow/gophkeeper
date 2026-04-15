// Package main предоставляет multichecker — единый статический анализатор кода на Go.
//
// # Запуск multichecker
//
// Из корня проекта:
//
//	go run ./cmd/staticlint/ ./...
//
// Или после сборки:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Проверка только пакетов проекта (без тестов и vendor):
//
//	./staticlint ./cmd/... ./internal/...
//
// Поддерживаются стандартные флаги go/analysis:
//   - -json — вывод в формате JSON;
//   - -fix — применить автоисправления, где возможно;
//   - -NAME — включить анализатор с именем NAME (по умолчанию все включены);
//   - -NAME=false — отключить анализатор NAME.
//
// # Состав multichecker
//
// 1. Стандартные анализаторы пакета golang.org/x/tools/go/analysis/passes:
//   - asmdecl — соответствие объявлений в Go и ассемблере;
//   - assign — бесполезные присваивания;
//   - atomic — типичные ошибки использования sync/atomic;
//   - atomicalign — выравнивание аргументов атомарных функций;
//   - bools — подозрительные конструкции с булевыми операторами;
//   - buildtag — проверка директив сборки (go:build и плюс build) в комментариях;
//   - cgocall — правила передачи указателей в CGO;
//   - composite — композитные литералы без ключей полей (из других пакетов);
//   - copylock — копирование значений, содержащих блокировки;
//   - deepequalerrors — использование reflect.DeepEqual для error;
//   - defers — типичные ошибки в defer;
//   - directive — проверка директив компилятора (//go:);
//   - framepointer — интерпретация фреймов;
//   - httpresponse — использование тела ответа HTTP;
//   - ifaceassert — бессмысленные утверждения типов интерфейсов;
//   - loopclosure — захват переменных в замыканиях циклов;
//   - lostcancel — отмена контекста в цепочке вызовов;
//   - nilfunc — вызов nil-функции;
//   - nilness — анализ nil-указателей;
//   - pkgfact — сбор фактов о пакетах;
//   - printf — соответствие аргументов и формата в Printf-подобных вызовах;
//   - reflectvaluecompare — сравнение reflect.Value;
//   - shadow — затенение переменных;
//   - shift — сдвиги с неверной шириной;
//   - sortslice — корректность использования sort.Slice и т.п.;
//   - stdmethods — сигнатуры стандартных интерфейсов (например, String());
//   - structtag — синтаксис тегов структур;
//   - tests — типичные ошибки в тестах;
//   - unmarshal — передача не-указателей в Unmarshal/Decode;
//   - unreachable — недостижимый код;
//   - unsafeptr — использование unsafe.Pointer;
//   - unusedresult — неиспользуемые возвращаемые значения.
//
// 2. Все анализаторы класса SA пакета staticcheck (honnef.co/go/tools/staticcheck):
//   - SA1xxx — некорректное использование стандартной библиотеки (регексы, time, encoding и т.д.);
//   - SA2xxx — проблемы конкурентности (WaitGroup, Lock/Unlock);
//   - SA3xxx — проблемы в тестах (TestMain, b.N);
//   - SA4xxx — бесполезный или подозрительный код (одинаковые операнды, dead code);
//   - SA5xxx — корректность (nil map, defer, пустые циклы);
//   - SA6xxx — производительность (regexp в цикле, sync.Pool);
//   - SA9xxx — сомнительные конструкции (defer в range, FileMode и т.д.).
//
// 3. Анализаторы других классов staticcheck:
//   - simple (S) — упрощение кода (каналы, copy, strings.Contains, append и т.д.);
//   - stylecheck (ST) — стиль (комментарии к пакету, имена, формат ошибок, порядок return и т.д.).
//
// 4. Сторонние анализаторы:
//   - errcheck — проверка необработанных ошибок (github.com/kisielk/errcheck);
//   - bodyclose — проверка закрытия тела HTTP-ответа (github.com/timakin/bodyclose).
//
// 5. Собственный анализатор exitcheck:
//   - Запрещает прямой вызов os.Exit в функции main пакета main.
//   - Цель: обеспечить выполнение defer и упростить тестирование; рекомендуется возвращать код из main и вызывать os.Exit в одной точке.
package main
