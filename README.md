# notifyer

HTTP-сервіс сповіщень на Go: зберігає сповіщення в PostgreSQL і відправляє їх
через канали `console`, `email`, `telegram`.

## Запуск

```bash
cp .env.example .env
docker compose up -d        # PostgreSQL
air                         # або: go run ./cmd/notifyer
```

## Міграції

Схема бази описана SQL-файлами в `migrations/`. Інструмент для них —
[goose](https://github.com/pressly/goose).

### Як це працює

- Файли `migrations/*.sql` вбудовуються в бінарник через `embed` (`migrations/embed.go`).
- **На старті сервіс сам застосовує всі нові міграції** (`internal/platform/postgres/migrate.go`).
  Для звичайної роботи нічого запускати вручну не треба.
- goose зберігає, які міграції вже виконані, у таблиці `goose_db_version`.

### Встановити CLI

CLI потрібен лише для ручних команд нижче:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.29.0   # та сама версія, що в go.mod
```

### Команди

Команди беруть `DATABASE_URL` з `.env`, тому файл `.env` має існувати.

```bash
make migrate-create name=add_status   # створити migrations/0000N_add_status.sql
make migrate-status                   # показати, які міграції застосовані
make migrate-down                     # відкотити останню міграцію
```

### Як додати міграцію

1. `make migrate-create name=<що_робить>`. Назва — дієслово + об'єкт, наприклад `create_users`, `add_status_to_notifications`.
2. Заповнити обидві секції файлу:

   ```sql
   -- +goose Up
   ALTER TABLE notifications ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';

   -- +goose Down
   ALTER TABLE notifications DROP COLUMN status;
   ```

   `Down` має точно скасовувати `Up`, інакше `make migrate-down` залишить базу в незрозумілому стані.
3. Перезапустити сервіс, і міграція застосується. `air` не стежить за `.sql`-файлами,
   тому після зміни міграції перезапустіть його вручну або змініть будь-який `.go`-файл.
4. Перевірити відкат: `make migrate-down`, потім знову запустити сервіс.

### Правила

- **Не змінюйте міграцію, яку вже застосовано** (у себе, у колег чи на сервері).
  goose не виконає її вдруге. Щоб щось виправити, створіть нову міграцію.
- Одна міграція — одна логічна зміна.
- Нумерація послідовна (`00001`, `00002`…). Це забезпечує прапорець `-s` у `make migrate-create`.

## API

| Метод | Шлях | Опис |
|---|---|---|
| GET | `/health` | стан сервісу і версія |
| POST | `/api/notifications` | створити й відправити сповіщення |
| GET | `/api/notifications` | список |
| GET | `/api/notifications/{id}` | одне сповіщення |
| PATCH | `/api/notifications/{id}` | оновити |
| GET | `/api/notifications/export` | завантажити всі сповіщення текстовим файлом |

Приклади запитів — у `notifyer.postman_collection.json`.

## Тести

```bash
go test ./...
```

## Архітектура

Див. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
