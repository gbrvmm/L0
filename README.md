# L0 | Orders Demo | NATS Streaming + PostgreSQL + Go

[![Go](https://img.shields.io/badge/Go-1.24-blue)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791)](https://www.postgresql.org/)
[![NATS Streaming](https://img.shields.io/badge/NATS-Streaming-1997B5)](https://docs.nats.io/nats-concepts/what-is-nats)
[![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-red)](#)

Демонстрационный сервис заказов: читает события из **NATS Streaming (STAN)**, валидирует и сохраняет в **PostgreSQL** (хранение в **JSONB**), прогревает **in‑memory** кэш и отдает данные через **HTTP API** и простую веб‑страницу.

[![Смотреть видео](https://img.shields.io/badge/-Смотреть_видео-red?style=for-the-badge&logo=youtube)](https://disk.yandex.ru/i/3JA-TTrvnp-8rA)


---

## Оглавление

- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Стек](#стек)
- [Быстрый старт](#быстрый-старт)
  - [Вариант A — Docker Compose](#вариант-a--docker-compose)
  - [Вариант B — Локально](#вариант-b--локально)
- [Переменные окружения](#переменные-окружения)
- [HTTP API](#http-api)
- [Статический UI](#статический-ui)
- [Структура проекта](#структура-проекта)
- [Команды Makefile](#команды-makefile)
- [Надёжность и поведение подписчика](#надёжность-и-поведение-подписчика)
- [Модель хранения в БД](#модель-хранения-в-бд)
- [Полезные ссылки](#полезные-ссылки)

---

## Возможности

- Подписка на канал **STAN** и обработка входящих JSON‑заказов
- **Валидация** модели заказа перед сохранением
- Запись заказа в **PostgreSQL** (JSONB)
- **Прогрев кэша** из БД при запуске; быстрые чтения из памяти
- **HTTP API** для получения заказа по `order_uid` и **веб‑страница** для удобного просмотра
- **Graceful shutdown**

## Архитектура

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│  Publisher  │ ───▶ │     NATS     │ ───▶ │   Service   │
│ (cmd/pub)   │      │  Streaming   │      │  (server)   │
└─────────────┘      └──────────────┘      └──────┬──────┘
                                                  │
                                           ┌──────▼──────┐
                                           │ PostgreSQL  │
                                           │  (JSONB)    │
                                           └──────┬──────┘
                                                  │
                                           ┌──────▼──────┐
                                           │ In-Memory   │
                                           │   Cache     │
                                           └──────┬──────┘
                                                  │
                                           ┌──────▼──────┐
                                           │  HTTP API   │
                                           │   + Web UI  │
                                           └─────────────┘
```

## Стек

| Компонент     | Версия / Примечание                  |
|---------------|--------------------------------------|
| Go            | 1.24 (см. `go.mod`)                  |
| PostgreSQL    | 15                                   |
| NATS Streaming| STAN (файл `internal/stan/stan.go`)  |
| Хранение      | JSONB + кэш (`map` + `RWMutex`)      |
| Web           | Чистый HTTP + статические файлы      |

---

## Быстрый старт

### Вариант A — Docker Compose

1. Поднимите инфраструктуру (**PostgreSQL** и **NATS Streaming**):
   ```bash
   docker compose up -d
   ```

2. Запустите сервер локально (или создайте бинарники через `make build`):
   ```bash
   make run
   ```

3. Отправьте тестовый заказ в канал STAN:
   ```bash
   make publish
   # публикует sample/model.json в канал (см. CHANNEL)
   ```

4. Откройте UI:
   - `http://localhost:8080` — страница поиска заказа по `order_uid`
   - `http://localhost:8080/healthz` — health‑check

### Вариант B — Локально

1. Создайте БД и пользователя (пример для разработки):
   ```sql
   CREATE USER orders WITH PASSWORD 'orders';
   CREATE DATABASE orders OWNER orders;
   GRANT ALL PRIVILEGES ON DATABASE orders TO orders;
   ```

2. Экспортируйте переменные окружения (см. раздел ниже) и запустите сервер:
   ```bash
   go run ./cmd/server
   ```

---

## Переменные окружения

> Значения по умолчанию берутся в коде (`internal/config/config.go`).

**PostgreSQL**

| Ключ     | По умолчанию | Описание                  |
|----------|---------------|---------------------------|
| `DB_HOST`| `localhost`   | Хост БД                   |
| `DB_PORT`| `5432`        | Порт БД                   |
| `DB_NAME`| `orders`      | Имя БД                    |
| `DB_USER`| `orders`      | Пользователь              |
| `DB_PASS`| `orders`      | Пароль                    |
| `DB_SSL` | `disable`     | Режим SSL для драйвера    |

**HTTP**

| Ключ        | По умолчанию | Описание          |
|-------------|--------------|-------------------|
| `HTTP_ADDR` | `:8080`      | Адрес HTTP‑сервера|

**NATS Streaming (STAN)**

| Ключ              | По умолчанию         | Описание                |
|-------------------|----------------------|-------------------------|
| `STAN_CLUSTER_ID` | `test-cluster`       | ID кластера             |
| `STAN_CLIENT_ID`  | `orders-server-1`    | ID клиента              |
| `STAN_URL`        | `nats://localhost:4222` | Адрес STAN           |
| `CHANNEL`         | `orders`             | Канал с заказами        |
| `DURABLE`         | `orders-durable`     | Durable‑имя подписки    |

---

## HTTP API

| Метод | Путь                     | Описание                        | Успех | Ошибка |
|------:|--------------------------|----------------------------------|:-----:|:------:|
| GET   | `/api/orders/{order_uid}`| Вернуть заказ (JSON из кэша/БД)  | 200   | 404    |
| GET   | `/healthz`               | Проверка живости сервиса         | 200   | 500    |

Пример запроса:
```bash
curl -s http://localhost:8080/api/orders/b563feb7b2b84b6test | jq .
```

---

## Статический UI

Файлы в `web/static/`:
- `index.html` — форма поиска по `order_uid`
- `script.js` — запрос к `/api/orders/{id}` и рендер ответа
- `styles.css` — лёгкие стили

В футере есть ссылка на `/healthz` для быстрой проверки.

---

## Структура проекта

```
L0/
├── cmd/
│   ├── publisher/            # CLI: публикует тестовый заказ в STAN
│   └── server/               # HTTP-сервер + подписчик STAN
├── internal/
│   ├── cache/                # in-memory кэш (map + RWMutex)
│   ├── config/               # чтение конфигурации из env
│   ├── db/                   # работа с PostgreSQL (JSONB)
│   ├── model/                # модель заказа + валидация
│   └── stan/                 # init STAN, durable-подписка, ack
├── sample/
│   └── model.json            # пример заказа (для publisher)
├── web/
│   └── static/               # index.html, styles.css, script.js
├── Makefile                  # цели build/run/publish
├── docker-compose.yml        # Postgres + NATS Streaming
├── go.mod / go.sum
└── README.md
```

---

## Команды Makefile

```makefile
tidy       # go mod tidy
build      # сборка bin/orders-server и bin/orders-publisher
run        # запуск сервера с локальными env
publish    # публикация sample/model.json в канал STAN
```

---

## Надёжность и поведение подписчика

<details>
<summary>Подписка, доставляемость и обработка ошибок</summary>

- **Durable‑подписка** + `DeliverAllAvailable` — при рестарте дочитываем накопившиеся сообщения.
- **Ручной ACK** — подтверждаем обработку только после успешной записи в БД/кэш.
- **Idempotency** — `order_uid` уникален; дубликаты не ломают состояние.
- **Bad messages** — некорректные сообщения сохраняются в таблицу `bad_messages` с причиной (например, «invalid json»), чтобы не терять факт получения.
</details>

---

## Модель хранения в БД

Одна основная таблица:
```sql
CREATE TABLE IF NOT EXISTS orders (
  order_uid   text PRIMARY KEY,
  data        jsonb NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now()
);
```
Для эволюции схемы используем JSONB; при необходимости можно добавить денормализованные индексы/таблицы.

---

## Полезные ссылки

- STAN monitoring (если открыт порт `8222`): `http://localhost:8222/`
- `go.mod`: модуль `github.com/gbrvmm/L0`
- Пример `order_uid` для UI: `b563feb7b2b84b6test`

---

