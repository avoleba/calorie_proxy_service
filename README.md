# Calorie Proxy Service

Прокси-сервис для поиска продуктов и учёта калорий. Агрегирует данные внешних food API (по умолчанию Open Food Facts), кэширует ответы в Redis, хранит пользователей и корзину в PostgreSQL.

## Возможности

- Поиск продуктов и lookup по штрихкоду
- JWT-авторизация (регистрация / логин)
- Корзина с граммовкой и пересчётом КБЖУ
- Кэш Redis
- Метрики Prometheus (`/metrics`)
- CORS, graceful shutdown

## Стек

| Компонент | Технология |
|-----------|------------|
| Язык | Go 1.23 |
| БД | PostgreSQL 16 |
| Кэш | Redis 7 |
| Auth | JWT (HS256) |
| Метрики | Prometheus |

## Быстрый старт

### Docker Compose

```bash
cp .env.example .env
# при необходимости задайте JWT_SECRET и API-ключи

make docker-up
# или: docker compose up --build
```

Сервис: `http://localhost:8080`

Остановка:

```bash
make docker-down
```

### Локально

Нужны Go 1.23+, PostgreSQL и Redis.

```bash
cp .env.example .env
# DATABASE_URL и REDIS_ADDR должны указывать на локальные инстансы

make run
# или: go run cmd/proxy/main.go
```

Сборка бинарника:

```bash
make build   # → bin/proxy
```

## Конфигурация

Переменные окружения (см. `.env.example`):

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `SERVER_PORT` | `8080` | Порт HTTP |
| `DATABASE_URL` | `postgres://localhost:5432/calorie?sslmode=disable` | PostgreSQL |
| `REDIS_ADDR` | `localhost:6379` | Redis |
| `CACHE_TTL` | `24h` | TTL кэша |
| `JWT_SECRET` | `dev-secret-change-in-production` | Секрет JWT |
| `FOOD_PROVIDER` | `openfoodfacts` | Провайдер данных |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,...` | Разрешённые origins |
| `USDA_API_KEY` / `EDAMAM_*` | — | Опциональные ключи внешних API |

В Docker Compose Postgres: `postgres://calorie:calorie@postgres:5432/calorie?sslmode=disable`.

## API

Базовый URL: `http://localhost:8080`

Защищённые эндпоинты требуют заголовок:

```http
Authorization: Bearer <token>
```

### Health & metrics

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| `GET` | `/health` | нет | Статус сервиса |
| `GET` | `/metrics` | нет | Prometheus metrics |

### Auth

#### `POST /api/v1/auth/register`

```json
{ "email": "user@example.com", "password": "secret1" }
```

Ответ — JWT и пользователь. Пароль минимум 6 символов.

#### `POST /api/v1/auth/login`

```json
{ "email": "user@example.com", "password": "secret1" }
```

### Продукты

#### `GET /api/v1/foods/search?q=milk&page=1&page_size=10`

Поиск. Query-параметры: `q` (обязательный), `page`, `page_size` (≤ 100).

#### `GET /api/v1/foods/barcode?barcode=3017620422003`

Поиск по штрихкоду.

Ответы кэшируются; заголовок `X-Cache: HIT|MISS`.

### Корзина

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/cart` | Добавить продукт (`product` + `grams`) |
| `GET` | `/api/v1/cart` | Список с итогами КБЖУ |
| `PATCH` / `PUT` | `/api/v1/cart?id=<item_id>` | Изменить граммовку |
| `DELETE` | `/api/v1/cart?id=<item_id>` | Удалить позицию |

Пример добавления:

```json
{
  "product": {
    "id": "12345",
    "name": "Молоко 3.2%",
    "brand": "Простоквашино",
    "source": "openfoodfacts",
    "nutrition": {
      "calories": 60,
      "protein": 2.8,
      "fat": 3.2,
      "carbohydrates": 4.7
    }
  },
  "grams": 200
}
```

Подробнее: [`docs/api-cart-post.md`](docs/api-cart-post.md).

## Структура проекта

```
cmd/proxy/           # точка входа
internal/
  auth/              # JWT, register/login
  cache/             # Redis
  cart/              # корзина
  clients/           # Open Food Facts и др.
  config/            # env-конфиг
  middleware/        # CORS, logging, metrics
  models/            # DTO
  proxy/             # поиск продуктов
  store/             # PostgreSQL
docs/                # API-заметки
```

## Make-команды

| Команда | Действие |
|---------|----------|
| `make build` | Сборка `bin/proxy` |
| `make run` | Запуск локально |
| `make test` | Тесты |
| `make docker-up` | Поднять стек |
| `make docker-down` | Остановить стек |
| `make clean` | Удалить бинарник и volumes |
| `make lint` | golangci-lint |

## Пример запроса

```bash
# регистрация
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret1"}' | jq -r .token)

# поиск
curl -s "http://localhost:8080/api/v1/foods/search?q=oat" \
  -H "Authorization: Bearer $TOKEN" | jq
```
