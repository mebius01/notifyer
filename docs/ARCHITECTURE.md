# Архітектура notifyer

Структура побудована за конвенціями Go-ком'юніті:

- [Organizing a Go module](https://go.dev/doc/modules/layout) — офіційна рекомендація команди Go (`cmd/`, `internal/`);
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) і [Google Go Style Guide](https://google.github.io/styleguide/go/) — назви, інтерфейси, помилки;
- [Standard Package Layout](https://www.gobeyond.dev/standard-package-layout/) (Ben Johnson) — домен у центрі, адаптери як підпакети;
- [How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/) (Mat Ryer) — `run()`, роутер, middleware.

`golang-standards/project-layout` **не** є офіційним стандартом; з нього взято лише `cmd/` та `internal/`.

## Структура

```
cmd/notifyer/main.go         точка входу: збирає залежності, запускає сервер, graceful shutdown
internal/                    весь код застосунку (ззовні модуля імпортувати не можна)
  apperr/                    загальні помилки: Kind (NotFound, Invalid…), Code, Message; без HTTP/gRPC
  config/                    читання env один раз на старті
  platform/postgres/         спільна інфраструктура: пул з'єднань, міграції
  httpapi/                   транспорт (HTTP)
    server.go                Server, NewHandler, routes() — весь API в одному місці
    middleware.go            логування запитів
    response.go              writeJSON, writeError, decodeJSON, parseID
    errors.go                apperr.Kind → HTTP-статус (writeServiceError)
    notification.go          хендлери + DTO сповіщень
  notification/              фіча "сповіщення"
    notification.go          доменний тип + Validate()
    errors.go                доменні помилки як *apperr.Error (ErrNotFound, ErrInvalid…)
    repository.go            інтерфейси, які потрібні сервісу: Repository, Sender
    service.go               бізнес-сценарії: Create, Get, List, Update, Export
    postgres/                Repository для PostgreSQL
    sender/                  реалізації Sender: console, email, sms, telegram, logging (файл на канал)
migrations/                  SQL-міграції goose, вбудовані в бінарник (embed.go)
docs/                        документація
```

## Напрям залежностей

```
httpapi ──► notification ◄── notification/postgres, notification/sender
                 ▲
cmd/notifyer ────┘   (main знає всіх і з'єднує їх)
```

- Пакет фічі (`notification`) не імпортує HTTP, базу чи інші фічі.
- Адаптери (postgres, sender, httpapi) імпортують домен, а не навпаки.
- Тільки `main.go` знає про всі пакети і передає залежності через конструктори.

## Правила

1. **Пакети за фічами**, а не за типами файлів. Без `models/`, `services/`, `utils/`, `common/`.
2. **Назва пакета — частина імені**: `notification.Service`, `sender.NewEmail`, а не `NotificationService`.
3. **Інтерфейси оголошує споживач**, і вони маленькі. Реалізації повертають конкретні типи.
4. **Залежності явні** — через конструктор. Без глобальних змінних, `init()` і DI-фреймворків.
5. **`main` тонкий**, уся логіка запуску — у `run() error`.
6. **`context.Context` першим параметром** у кожному методі з I/O.
7. **Помилки** — три шари:
   - `apperr` — категорія (`Kind`), стабільний `Code` для клієнта, безпечний `Message`, внутрішня причина `Err`.
   - Домен оголошує свої помилки як `*apperr.Error` у `errors.go`. Деталі для клієнта — `ErrX.WithDetail(...)`, внутрішня причина — `ErrX.Wrap(err)`, контекст — `fmt.Errorf("op: %w", err)`. `errors.Is` порівнює за `Code`. Домен не імпортує `net/http` чи gRPC.
   - Транспорт мапить `Kind` на свій код в одному місці (`httpapi/errors.go`). Невідомі помилки → 500 без деталей, повна помилка — в лог. Особливі випадки (як 502 з `id` для `ErrSendFailed`) — у хендлері домену.
   - Тіло помилки: `{"error": "<message>", "code": "<code>"}`.
8. **Доменні типи без JSON-тегів.** Формат API описують DTO в `httpapi`.
9. **Тести поруч із кодом**, табличні, з фейками замість mock-бібліотек; HTTP — через `httptest`.
   Фейки (`fakeRepo`, `fakeSender`) живуть у `_test.go`-файлах, а не окремими пакетами:
   вони не потрапляють у бінарник і не виглядають як частина застосунку.
10. **Папку створювати, коли вона потрібна**, а не "на виріст".

## Як додати нову фічу (на прикладі `user`)

1. `internal/user/user.go` — тип `User` + `Validate()` + тест.
2. `internal/user/errors.go` — `ErrNotFound`, `ErrInvalid`, `ErrEmailTaken`… як `*apperr.Error` з `Kind` і `Code` (`user.not_found`).
3. `internal/user/repository.go` + `service.go` — інтерфейс сховища і сценарії; тести з фейком `fakeRepo` у `service_test.go`.
4. `internal/user/postgres/` — реалізація сховища.
5. `migrations/0000N_create_users.sql` — `make migrate-create name=create_users`.
6. `internal/httpapi/user.go` — DTO і хендлери. Секретні поля (`PasswordHash`) у response DTO не потрапляють.
7. `internal/httpapi/server.go` — поле `users *user.Service`, параметр у `NewHandler`, маршрути в `routes()`.
8. `cmd/notifyer/main.go` — створити репозиторій і сервіс, передати в `NewHandler`.

Якщо одній фічі потрібні дані іншої (наприклад, `notification` має знати email користувача),
не імпортуйте чужий сервіс. Оголосіть у споживачі маленький інтерфейс
(`type UserLookup interface { Email(ctx context.Context, id int) (string, error) }`)
і передайте реалізацію з `main.go`. Так не буде циклічних імпортів.
