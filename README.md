# Амана (Amana) — CRM честной рассрочки без рибы

> CRM для бизнеса Чеченской Республики, который продаёт товары **в рассрочку по
> исламской модели мурабаха**: фиксированная цена без процента, **садака вместо
> штрафов** за просрочку и планирование задач **мимо времён намаза**. Плюс
> собственный **публичный API** для интеграций.

Бэкенд: **Go**, модульный монолит, чистая архитектура + тактический DDD,
PostgreSQL, типобезопасный SQL (`sqlc`), точная денежная математика (без `float`).

---

## Быстрый старт — одна команда

```bash
docker compose up --build
```

Поднимется PostgreSQL 16 и приложение; на старте **прогоняются миграции** и
**засеваются демо-данные** (грозненский колорит).

- **Веб-приложение** (лендинг + CRM): `http://localhost:8080/` — встроено в бинарь.
- Swagger UI (публичный API): `http://localhost:8080/swagger/`
- Проверка живости: `GET http://localhost:8080/health` → `{"status":"ok"}`

Фронтенд — самостоятельный SPA (React + htm, без сборки), встроенный в Go-бинарь
через `//go:embed` и подключённый к тому же API. Исходники — в `web/`.

### Локально (без Docker)

Нужен PostgreSQL на `localhost:5432` (`amana`/`amana`/`amana`).

```bash
cp .env.example .env
make run            # миграции на старте; для сидов: SEED_ON_BOOT=true make run
```

---

## Демо-доступы (засеваются автоматически)

| Роль | Email | Пароль |
|------|-------|--------|
| Владелец | `owner@amana.ru` | `owner12345` |
| Менеджер | `manager@amana.ru` | `manager12345` |

**Демо `X-API-Key` (публичный API):** `amana_demo_marketplace_key_2026`

Сид создаёт организацию «Грозный Мебель», 3 клиентов, 5 товаров (разные
халяль-статусы), 3 договора (частично оплаченный; просроченный с начисленной
садакой; свежий) и 3 напоминания (часть — со сдвигом мимо намаза/джума).

---

## Демо-сценарий (3 минуты)

1. **Вход** владельцем (`owner@amana.ru`).
2. **Договор мурабаха**: `POST /api/app/contracts/preview` показывает график и блок
   «без рибы vs обычный кредит» (цена фиксирована, 0% риба) — расчёт на бэкенде.
3. **Календарь намаза**: `POST /api/app/schedule/reminders` на доставку в пятницу 13:00 —
   система сдвигает её за джума-намаз (`was_shifted: true`, причина в ответе).
4. **Реестр садаки**: `GET /api/app/charity` — штраф за просрочку ушёл на
   благотворительность и **не изменил долг**.
5. **Расширяемость**: открываем Swagger `/swagger/`, с демо `X-API-Key` создаём договор
   «от имени внешнего маркетплейса» через `POST /api/v1/contracts`.

---

## Две поверхности API

| Поверхность | Префикс | Авторизация | Назначение |
|-------------|---------|-------------|------------|
| Внутренний API приложения | `/api/app` | JWT (`Authorization: Bearer`) | все экраны SPA |
| Публичный API (+3) | `/api/v1` | `X-API-Key` | интеграции (создать договор, статус платежей) |
| Swagger UI | `/swagger/` | — | документирует **публичный** API |

Ключевые эндпоинты `/api/app`: `auth/login`, `setup`, `users`, `api-keys`,
`catalog`, `clients`, `contracts` (+ `preview`, `payments`, `settle`, `cancel`,
`charity`), `charity`, `schedule/reminders`, `schedule/preview`.

Публичные `/api/v1`: `POST /contracts`, `GET /contracts/{id}/payments`.

---

## Ключевые доменные правила

- **Запрет рибы:** сумма обязательства фиксируется при создании (`себестоимость + наценка`)
  и **не зависит от времени**. Просрочка не меняет цену, остаток и график.
- **Садака вместо пени:** сбор за просрочку — фиксированный, в отдельный реестр
  благотворительности, **не** в выручку и **не** в долг. Начисляет только владелец.
- **Намаз:** задачи не попадают в окна 5 молитв и пятничного джума (Грозный,
  мазхаб шафиитский) — сдвигаются вперёд с объяснением причины.
- **Халяль:** договор нельзя оформить на товар со статусом «харам».
- **Деньги:** в домене — `shopspring/decimal`; во внутренних расчётах графика —
  `int64` копейки; в БД — `NUMERIC(18,2)`; на границе API — **строка-decimal**
  (`"120000.00"`). Никакого `float`.

Точные формулы, инварианты и тест-кейсы — в `.claude/skills/murabaha-engine/SKILL.md`
и `.claude/skills/namaz-scheduler/SKILL.md`.

---

## Архитектура

Модульный монолит: один бинарь, внутри изолированные bounded contexts
(`iam`, `catalog`, `crm`, `financing`, `scheduling`) с границами через интерфейсы.
Зависимости направлены внутрь: `http → app → domain ← infra`. Любой модуль
вынимается в микросервис без переписывания домена (меняется только реализация
портов). Деньги — общий kernel `internal/shared/money`. Полное описание —
`ARCHITECTURE.md`, доменная модель — `docs/DOMAIN.md`.

**Стек:** Go 1.23+, `go-chi/chi`, PostgreSQL 16, `jackc/pgx/v5` + `sqlc`,
`pressly/goose`, `shopspring/decimal`, `log/slog`, `golang-jwt/jwt/v5`, `bcrypt`,
`swaggo/swag`, `hablullah/go-prayer` (за портом), `stretchr/testify`.

---

## Разработка

```bash
make run        # запустить API локально
make test       # все тесты (table-driven доменная математика)
make build      # собрать бинарь в ./bin
make migrate    # применить миграции и выйти
make sqlc       # сгенерировать типобезопасный SQL (нужен sqlc)
make swag       # перегенерировать OpenAPI спецификацию (нужен swag)
make lint       # go vet
make docker-up  # docker compose up --build
make docker-down
```

Инструменты для кодогенерации (если нужны локально):

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Тесты

Доменная математика покрыта table-driven тестами:
- `financing/domain` — все 11 кейсов мурабахи (округление до копейки, инварианты,
  статусы долей, просрочка не меняет долг, садака отдельно, досрочка, preview==создание);
- `scheduling/domain` — все 7 кейсов планировщика (на фиктивном `PrayerTimesProvider`);
- `shared/money`, `iam`, `catalog`, `crm` — конструкторы и инварианты.

```bash
go test ./...
```

---

## Структура репозитория

```
amana/
├── cmd/api/                 # точка входа (вся DI-проводка)
├── internal/
│   ├── platform/            # config, database, httpserver, logger, apperror, authctx, web, pgconv
│   ├── shared/money/        # value object Money (shared kernel)
│   ├── modules/             # iam, catalog, crm, financing, scheduling (domain/app/infra/http)
│   ├── publicapi/v1/        # +3: публичный API поверх app-сервисов
│   └── seed/                # демо-данные
├── migrations/              # goose-миграции (эмбедятся в бинарь)
├── api/openapi/             # сгенерированный swagger.json/yaml
├── docs/                    # DOMAIN.md, SITE_STRUCTURE.md
├── .claude/skills/          # murabaha-engine, namaz-scheduler, new-module
├── docker-compose.yml · Dockerfile · Makefile · sqlc.yaml · .env.example
├── CLAUDE.md · ARCHITECTURE.md · DOCUMENTATION.md · CHANGELOG.md
```
