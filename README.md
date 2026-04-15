# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование (pprof)
### Что бы сделано
В репозитории `double_maps` вместо использования `var result []model.ShortURL` теперь используется единовременная аллокация `make([]model.ShortURL, 0, size)`.

В сервисе `url_service`:
* Вместо `s.cfg.BaseURL + "/" + resultKey` используется функция `buildFullURL`, которая позволяет избежать лишних аллокаций при конкатенации строк;
* Вместо `var newURLs []model.ShortURL` как и в репозитории используется единовременная аллокация `newURLs := make([]model.ShortURL, 0, len(URLs))`;
* Дополнительно переделан поиск уже существуюших ключей. Добавлена мапа urlIndexByOriginal, которую можно собрать один раз (затраты) и потом быстро находить то что нужно по ключу (выгода).

### Нагрузка
```shell
for ($i = 0; $i -lt 80000; $i++) {
  Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:8080/" -Body "https://example.com/path/$i" -ContentType "text/plain" -UseBasicParsing | Out-Null
  if ($i % 1000 -eq 0) { Write-Host "Sent $i requests" }
}
```

### Снятие профиля в момент нагрузки
```shell
curl "http://127.0.0.1:8080/debug/pprof/allocs?seconds=30" -o profiles/base_allocs.pprof // ДО ОПТИМИЗАЦИИ
curl "http://127.0.0.1:8080/debug/pprof/allocs?seconds=30" -o profiles/base_allocs.pprof // ПОСЛЕ ОПТИМИЗАЦИИ

```

### Результат
```log
PS I:\practicum\shortener> go tool pprof -top -diff_base=.\profiles\base_allocs.pprof .\profiles\result_allocs.pprof 
File: shortener.exe
Build ID: C:\Users\user\go\tmp\go-build4262431577\b001\exe\shortener.exe2026-02-24 18:06:26.3996441 +0300 MSK        
Type: inuse_space
Time: 2026-02-24 18:07:03 MSK
Duration: 60.01s, Total samples = 6039.17kB
Showing nodes accounting for 3942.63kB, 65.28% of 6039.17kB total
      flat  flat%   sum%        cum   cum%
 1805.17kB 29.89% 29.89%  2349.84kB 38.91%  compress/flate.NewWriter (inline)
 1596.78kB 26.44% 56.33%  1596.78kB 26.44%  github.com/coolycow/shortener/internal/repository.(*DoubleMapsRepository).SaveURL
  544.67kB  9.02% 65.35%   544.67kB  9.02%  compress/flate.(*compressor).initDeflate (inline)
 -516.01kB  8.54% 56.81%  -516.01kB  8.54%  io.init.func1
  512.02kB  8.48% 65.28%   512.02kB  8.48%  strings.(*Builder).grow
         0     0% 65.28%  -516.01kB  8.54%  bufio.(*Writer).Flush
         0     0% 65.28%   544.67kB  9.02%  compress/flate.(*compressor).init
         0     0% 65.28%  2349.84kB 38.91%  compress/gzip.(*Writer).Write
         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.NewRouter.ErrorHandler.func3
         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.NewRouter.Gzip.func1       
         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.NewRouter.RequestGzip.func4         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.NewRouter.RequestLogger.func2
         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.setupURLRoutes.OptionalAuthMiddleware.func2
         0     0% 65.28%  4458.64kB 73.83%  github.com/coolycow/shortener/internal/router.setupURLRoutes.PostHandler.func3
         0     0% 65.28%  1596.78kB 26.44%  github.com/coolycow/shortener/internal/service.(*urlService).CreateShortURL
         0     0% 65.28%  2349.84kB 38.91%  github.com/gin-gonic/contrib/gzip.(*gzipWriter).Write
         0     0% 65.28%  4458.64kB 73.83%  github.com/gin-gonic/gin.(*Context).Next
         0     0% 65.28%  2349.84kB 38.91%  github.com/gin-gonic/gin.(*Context).Render
         0     0% 65.28%  2349.84kB 38.91%  github.com/gin-gonic/gin.(*Context).String
         0     0% 65.28%  4458.64kB 73.83%  github.com/gin-gonic/gin.(*Engine).ServeHTTP
         0     0% 65.28%  4458.64kB 73.83%  github.com/gin-gonic/gin.(*Engine).handleHTTPRequest
         0     0% 65.28%  4458.64kB 73.83%  github.com/gin-gonic/gin.CustomRecoveryWithWriter.func1
         0     0% 65.28%  4458.64kB 73.83%  github.com/gin-gonic/gin.LoggerWithConfig.func1
         0     0% 65.28%  2349.84kB 38.91%  github.com/gin-gonic/gin/render.String.Render
         0     0% 65.28%  2349.84kB 38.91%  github.com/gin-gonic/gin/render.WriteString
         0     0% 65.28%  -516.01kB  8.54%  io.Copy (inline)
         0     0% 65.28%  -516.01kB  8.54%  io.CopyN
         0     0% 65.28%  -516.01kB  8.54%  io.copyBuffer
         0     0% 65.28%  -516.01kB  8.54%  io.discard.ReadFrom
         0     0% 65.28%  -516.01kB  8.54%  net/http.(*chunkWriter).Write
         0     0% 65.28%  -516.01kB  8.54%  net/http.(*chunkWriter).writeHeader
         0     0% 65.28%  3942.63kB 65.28%  net/http.(*conn).serve
         0     0% 65.28%  -516.01kB  8.54%  net/http.(*response).finishRequest
         0     0% 65.28%  4458.64kB 73.83%  net/http.serverHandler.ServeHTTP
         0     0% 65.28%   512.02kB  8.48%  net/url.(*URL).String
         0     0% 65.28%   512.02kB  8.48%  strings.(*Builder).Grow
         0     0% 65.28%  -516.01kB  8.54%  sync.(*Pool).Get
PS I:\practicum\shortener> 
```