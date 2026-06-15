---
name: new-module
description: Use this skill whenever you need to create a NEW bounded context / module in the Amana backend (e.g. financing, scheduling, catalog, crm, iam) or add a new feature that deserves its own module. It defines the exact folder layout, where interfaces vs implementations go, the dependency rules, how to wire the module into main.go, and which tests are required. Follow it every time so all modules look identical. Do NOT use for one-off helpers inside an existing module.
---

# Skill: создание нового модуля (bounded context)

Каждый модуль в `internal/modules/<name>/` строится по одной и той же структуре.
Никаких отклонений — единообразие важнее «как удобнее в этот раз».

## Шаг 1. Структура каталогов

```
internal/modules/<name>/
├── domain/
│   ├── <entity>.go        # сущности и корни агрегатов (конструкторы возвращают (T, error))
│   ├── value_objects.go   # value objects (Money-обёртки, статусы-перечисления)
│   ├── errors.go          # типизированные доменные ошибки (var Err... = errors.New(...))
│   ├── ports.go           # ИНТЕРФЕЙСЫ: репозитории и внешние зависимости этого домена
│   └── service.go         # доменные сервисы (логика, не принадлежащая одной сущности)
├── app/
│   └── <usecase>.go       # application-сервисы: по одному типу на сценарий
├── infra/
│   ├── queries/*.sql      # SQL для sqlc
│   ├── repository.go      # реализация порта репозитория поверх sqlc
│   └── <adapter>.go       # адаптеры внешних библиотек (если есть)
└── http/
    ├── dto.go             # структуры запроса/ответа + теги validator
    ├── handler.go         # хендлеры: парс → валидация → app → маппинг
    └── routes.go          # регистрация роутов модуля в chi.Router
```

## Шаг 2. Правила зависимостей (проверяй на каждом файле)

- `domain` импортирует **только** stdlib и `shopspring/decimal`. Ни `pgx`, ни `chi`,
  ни sqlc-кода, ни логгера.
- Интерфейсы (`ports.go`) объявляются **в domain/app** (на стороне потребителя),
  реализуются в `infra`. Это инверсия зависимостей — infra зависит от domain.
- `app` зависит от `domain` (сущности + порты). Не лезет в `net/http` и не знает про DTO.
- `http` зависит от `app`. Маппит DTO ↔ домен явно. Логики не содержит.

## Шаг 3. Доменный слой

1. Сущности и агрегаты — с приватными полями и конструктором `New<Entity>(...) (<Entity>, error)`,
   который проверяет инварианты. Невалидную сущность создать нельзя.
2. Изменения состояния агрегата — только через методы корня, не через сеттеры полей.
3. Деньги — через `Money` (см. CLAUDE.md правило №1). Никакого `float`.
4. Ошибки — типизированные, в `errors.go`.

## Шаг 4. Application слой

- Один тип на сценарий: `type CreateContract struct { repo domain.ContractRepository; ... }`
  с методом `Execute(ctx, input) (output, error)`.
- Транзакции — здесь, через `platform/database.WithinTx`.
- Первым аргументом всегда `context.Context`.
- Возвращает доменные ошибки как есть (маппинг на HTTP — забота слоя http).

## Шаг 5. Infra слой

- Репозиторий реализует интерфейс из `domain/ports.go`.
- SQL — в `queries/*.sql`, типобезопасные методы генерит sqlc (`make sqlc`).
- Репозиторий маппит строки БД ↔ доменные сущности (БД-модель ≠ доменная сущность).
- Новые таблицы — только goose-миграцией в `/migrations`.

## Шаг 6. HTTP слой

- DTO с тегами `validate:"..."`.
- Хендлер: декодирует body → `validator.Struct(dto)` → вызывает app-сервис →
  маппит результат в response-DTO **или** доменную ошибку в HTTP-статус
  (через общий маппер `platform/apperror`).
- `routes.go` экспортирует `func (m *Module) RegisterRoutes(r chi.Router)`.

## Шаг 7. Сборка модуля и проводка

- Модуль экспортирует `func New(deps Deps) *Module`, собирающую repo → app → handlers.
- Регистрация в `cmd/api/main.go`: создать модуль, вызвать `RegisterRoutes` (внутренние роуты
  монтируются под `/api/app`). Вся проводка зависимостей — только там.

## Шаг 8. Тесты (обязательно)

- Доменная логика и инварианты — table-driven unit-тесты в `domain`.
- Хотя бы один тест на каждый недопустимый кейс конструктора (ожидаем ошибку).
- Репозиторий — опционально интеграционный тест через `testcontainers-go`.

## Чек-лист готовности модуля

- [ ] Структура каталогов соответствует шаблону.
- [ ] domain не импортирует инфраструктуру.
- [ ] Порты объявлены в domain, реализованы в infra.
- [ ] Конструкторы валидируют инварианты, возвращают error.
- [ ] Деньги через Money, без float.
- [ ] SQL через sqlc, схема — миграцией.
- [ ] Хендлеры без логики, ошибки маппятся централизованно.
- [ ] Table-driven тесты на домен зелёные.
- [ ] Модуль проводится в main.go, роуты зарегистрированы.
