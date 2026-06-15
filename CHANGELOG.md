# CHANGELOG — синхронизация документации «Амана»

Свод правок по результатам ревизии документов. Цель — убрать рассинхрон между
бэкенд-доками, продуктовой документацией и картой сайта, зафиксировать единые решения и
закрыть «висящие» места, которые иначе всплыли бы в коде или на демо.

Формат: **проблема → принятое решение → где применено**. Решения сквозные — если
встретишь старую формулировку где-то ещё, приводи к этому списку.

---

## Зафиксированные решения (canon)

### 1. Имя проекта — «Амана» (Amana)
Было: продуктовые доки — «Амана» (`amana/`), бэкенд-доки — «Барака» (`baraka/`).
**Решение:** единое имя **«Амана»**, корневая папка `amana/`, в текстах «Amana backend».
Имя «Барака»/`baraka` удалено везде.
*Применено:* `CLAUDE.md`, `ARCHITECTURE.md`, `docs/DOMAIN.md`, `.claude/skills/new-module/SKILL.md`.

### 2. Привилегированная роль — «владелец / менеджер»
Было: «admin/manager» (CLAUDE §7), «менеджер/админ» (DOMAIN), «владелец/менеджер»
(DOCUMENTATION, SITE_STRUCTURE).
**Решение:** продуктовый язык **«владелец / менеджер»**, коды `RoleOwner` / `RoleManager`.
Термин «admin» убран.
*Применено:* `CLAUDE.md` §7, `docs/DOMAIN.md` (iam).

### 3. Две поверхности API разведены явно
Было: публичный API на `/api/v1/` + `X-API-Key`; внутренний (для SPA по JWT) нигде не задан,
в `ARCHITECTURE.md` стояла заглушка `/api/.../...`.
**Решение:**
- **Внутренний API приложения** — префикс **`/api/app`**, авторизация **JWT** (`Authorization: Bearer`).
- **Публичный API** — префикс **`/api/v1`**, авторизация **`X-API-Key`**.
- Swagger UI на `/swagger/` документирует **публичный** API.
*Применено:* `CLAUDE.md` §6, `ARCHITECTURE.md` (поток запроса), `DOCUMENTATION.md` §12,
`docs/DOMAIN.md`, фронтовые `web/CLAUDE.md`, `web/ARCHITECTURE.md`.

### 4. Эндпоинт предпросмотра графика
Было: мастер показывает график **до** создания, но в домене только `CreateContract` и
`GetSchedule` (для существующего договора). Риск — расчёт графика в JS (нарушение анти-риба
правила «деньги считает домен»).
**Решение:** use-case **`PreviewContract`** в `financing/app`, эндпоинт
**`POST /api/app/contracts/preview`** — возвращает график, цену продажи и данные сравнения
**без сохранения**. Публичный API в MVP превью не отдаёт (нужен только мастеру).
*Применено:* `docs/DOMAIN.md` (financing/app), `.claude/skills/murabaha-engine/SKILL.md`,
`DOCUMENTATION.md` §6/§12.

### 5. Частичный платёж — модель статусов доли
Было: модал «Принять платёж» берёт произвольную сумму (≤ остатка), но статус `Installment`
только `оплачен/предстоит/просрочен` — непонятно, как красить недокрытую долю.
**Решение:** **`Outstanding` — единственный источник правды.** Статусы долей —
**производные** от накопленной оплаты. В перечисление статусов доли добавлен
**`PartiallyPaid` (частично оплачен)**:
- доля `Paid`, если суммарная оплата ≥ её верхней границы в графике;
- ровно одна доля на «фронте оплаты» может быть `PartiallyPaid`;
- остальные — `Pending` / `Overdue`.
Платёж — любая сумма `0 < amount ≤ Outstanding`.
*Применено:* `docs/DOMAIN.md` (статусы Installment, поведение платежа),
`.claude/skills/murabaha-engine/SKILL.md` (RegisterPayment, тест-кейс), `docs/SITE_STRUCTURE.md` §4.13.

### 6. Статусы `Draft` и `Cancelled` — границы UI
Было: домен описывает `Draft → Active`, но мастер создаёт сразу `Active`, действия «отменить»
в карточке нет — два статуса «висят» без точки входа.
**Решение:** state machine в домене сохраняется полностью. В **UI MVP**:
- мастер выполняет **create+activate атомарно**;
- `Draft` — внутреннее/для API состояние, **отдельным экраном не показывается**;
- **`Cancelled`** достижим простым действием **«Отменить договор» (только владелец)** на
  карточке договора.
Ничего не «висит».
*Применено:* `docs/DOMAIN.md` (жизненный цикл), `docs/SITE_STRUCTURE.md` §4.12/§4.13,
`DOCUMENTATION.md` §6.

### 7. Хранение и представление денег
Было: трижды «`NUMERIC(18,2)` **или** `BIGINT` в копейках, единообразно» — но выбор не сделан;
представление в API не задано.
**Решение (единообразно по проекту):**
- **БД:** `NUMERIC(18,2)` (выбрано; «или BIGINT копейки» для **хранения** убрано).
- **Домен:** value object `Money` на `shopspring/decimal` (точный).
- **Внутренние расчёты графика/округления:** `int64` в копейках (как в skill).
- **API (JSON):** деньги — **строка-decimal**, напр. `"120000.00"` (однозначно, без float).
*Применено:* `CLAUDE.md` §0.1/§4, `docs/DOMAIN.md` (Money), `.claude/skills/murabaha-engine/SKILL.md` (Деньги),
`DOCUMENTATION.md` §10/§11, фронтовые доки (трактовка денег как строки).

### 8. Тело публичного «создать договор»
Было: договору нужны `ClientID` и `ProductID`, но публичный API не описывает создание
клиента/товара — неясно, как создать договор «извне».
**Решение:** публичный `POST /api/v1/contracts` принимает **ссылки на существующие**
`client_id` и `product_id` (+ условия: `cost_price`, `markup`, `down_payment`,
`installments`, `cadence`, `start_date`). Клиент/товар создаются заранее (через приложение
или сидируются для демо).
*Применено:* `DOCUMENTATION.md` §12, `docs/DOMAIN.md` (iam/публичный API упоминание).

### 9. Кто начисляет садаку
Было: реестр `/charity` менеджеру — «просмотр», но кнопка «Начислить садаку» в карточке
договора доступна обеим ролям, и записи создаются именно оттуда — менеджер де-факто создаёт
записи.
**Решение:** **начисление садаки — только владелец.** Менеджер: реестр и баннер просрочки
видит, кнопка «Начислить садаку» для него **скрыта**. Реестр для менеджера — только чтение.
*Применено:* `docs/SITE_STRUCTURE.md` §4.13/§4.17/§7, `DOCUMENTATION.md` §7.

### 10. Дерево репозитория в `ARCHITECTURE.md`
Было: в дереве не было `web/` и `docs/SITE_STRUCTURE.md`; папка названа `baraka/`.
**Решение:** дерево дополнено `web/` и `docs/SITE_STRUCTURE.md`, корень — `amana/`.
Синхронизировано с `DOCUMENTATION.md` §15.
*Применено:* `ARCHITECTURE.md`.

### 11. Краевой случай первого взноса
Было: `0 <= DownPayment < SalePrice` допускает `FinancedAmount` меньше числа платежей →
часть долей становится нулевой (0 копеек).
**Решение:** добавлен инвариант **`FinancedAmount >= Installments`** (каждая доля ≥ 1 копейки)
+ тест-кейс на ошибку.
*Применено:* `docs/DOMAIN.md` (инварианты), `.claude/skills/murabaha-engine/SKILL.md` (инварианты, тесты).

### 12. Генерация типов API на фронте
Было: `DOCUMENTATION.md` §8 не упоминал, что фронт берёт типы из OpenAPI.
**Решение:** дописано — фронт генерит TS-типы из `swagger.json` (`openapi-typescript`),
аналог `sqlc` (типобезопасный контракт end-to-end).
*Применено:* `DOCUMENTATION.md` §8.

### 13. Цифра «171 млрд ₽ инвестиций» — проверена
Проверено по открытым источникам: цифра подтверждается (данные правительства ЧР, 2025;
рост ≈5% г/г; независимо — >170 млрд ₽ за 9 мес. 2025 против 153,71 млрд годом ранее).
**Решение:** оставлена, добавлена лёгкая атрибуция «(по данным правительства ЧР, 2025)»,
чтобы утверждение было защитимо на демо.
*Применено:* `DOCUMENTATION.md` §2.

---

## Затронутые файлы

| Файл | Пункты |
|------|--------|
| `CLAUDE.md` (бэкенд) | 1, 2, 3, 7 |
| `ARCHITECTURE.md` | 1, 3, 10 |
| `docs/DOMAIN.md` | 1, 2, 3, 4, 5, 6, 7, 8, 11 |
| `docs/SITE_STRUCTURE.md` | 5, 6, 9 |
| `DOCUMENTATION.md` | 2(подтв.), 3, 4, 6, 7(9), 8, 10, 11, 12, 13 |
| `.claude/skills/murabaha-engine/SKILL.md` | 4, 5, 7, 11 |
| `.claude/skills/new-module/SKILL.md` | 1 |
| `web/CLAUDE.md`, `web/ARCHITECTURE.md` | 3, 7 (выравнивание под `/api/app`, деньги-строка) |
| `.claude/skills/namaz-scheduler/SKILL.md` | без изменений |

Все правки сохраняют исходный стиль и структуру документов — менялись только затронутые
места, остальное не трогалось.

---

## Реализация бэкенда

Журнал фактической реализации Go-бэкенда по фазам. Фаза закрыта только при зелёных
`go build ./...` и `go test ./...`.

### Фаза 1 — Скелет и platform ✅
- **Структура** строго по `ARCHITECTURE.md`: `cmd/api`, `internal/platform/{config,logger,database,httpserver,apperror}`, заготовки `internal/modules` и `internal/publicapi`, `migrations`, `api/openapi`. Go-модуль `github.com/astralis-s/hakaton-ansar`, Go 1.24 (≥1.23).
- **config** — `caarlos0/env/v11` + `joho/godotenv`; группы HTTP / DB / Auth(JWT) / Logger / Prayer (дефолты — Грозный, шафиитский мазхаб).
- **logger** — `slog` (JSON/text, уровень из env).
- **database** — пул `pgxpool`; прогон goose-миграций на старте (через `database/sql` + `pgx/v5/stdlib`, миграции эмбедятся в бинарь); helper `WithinTx(ctx, pool, fn)`.
- **httpserver** — chi со стеком middleware (RequestID, RealIP, recover→JSON, структурный лог запроса, timeout, **CORS** для будущего SPA), graceful shutdown; Swagger UI на `/swagger/` документирует **публичный** API.
- **apperror** — классификация ошибок (`Kind` → HTTP-статус) и JSON-ответ в одном месте; детали 5xx не утекают клиенту.
- **Две поверхности HTTP**: `/api/app` (заглушка JWT-mw) и `/api/v1` (заглушка X-API-Key-mw), плюс `GET /health`. Заглушки auth помечены `TODO(phase2)`.
- **Инфра**: multi-stage `Dockerfile` → distroless; `docker-compose.yml` (app + postgres:16, healthcheck + `depends_on`); `Makefile` (run/build/test/migrate/sqlc/swag/lint/docker-*); `sqlc.yaml` (схема = миграции, генерация по пакету на модуль); `.env.example`; baseline-миграция `00001_init` (расширение `pgcrypto`).
- **Тесты**: table-driven на маршруты (`/health`, обе поверхности, `/swagger/doc.json`) и на маппинг ошибок `apperror` — зелёные.
- **Проверка рантайма**: против реального PostgreSQL 16 — миграции применились на старте (`goose … successfully migrated database to version: 1`), `/health = 200`, обе поверхности и Swagger UI отвечают. (`docker compose up` в песочнице без запущенного docker-демона не прогонялся; `Dockerfile`/`compose` написаны и валидны.)

### Фаза 2 — IAM ✅
Модуль `internal/modules/iam` строго по skill `new-module` (domain → app → infra → http).
- **domain** (стерилен, только stdlib): `Organization`, `User` (email через `net/mail`), `ApiKey` (+ `HashAPIKey` на sha256), VO `Role` (`RoleOwner`/`RoleManager`, без «admin»); конструкторы валидируют инварианты; порты `OrganizationRepository`/`UserRepository`/`ApiKeyRepository`, `PasswordHasher`, `TokenService`, `TxManager`, тип `Principal`.
- **infra**: sqlc-репозитории (генерация `make sqlc`), `BcryptHasher` (bcrypt), `JWTService` (golang-jwt/v5, HS256), маппинг `pgtype`↔домен, перевод unique-violation → `ErrEmailTaken`.
- **app**: `SetupOrganization` (создание орг+владельца атомарно в транзакции, повторно — `ErrAlreadyInitialized`), `Login` (без раскрытия существования email), `CreateUser`/`ListUsers`/`GetUser`, `CreateApiKey` (секрет один раз)/`ListApiKeys`/`RevokeApiKey`, `AuthenticateApiKey`.
- **http**: реальные middleware **JWT** (`/api/app`) и **X-API-Key** (`/api/v1`), `RequireOwner` (owner-only), DTO+validator, централизованный маппинг ошибок в HTTP. Заглушки auth из Фазы 1 удалены.
- **platform**: рефактор `database` на транзакции-через-контекст + `TxManager` + аксессор `Querier` (совместим с sqlc `DBTX`); новые `platform/authctx` (principal в контексте, без связности с модулем) и `platform/web` (decode+validate→apperror, JSON). Установлены CLI `sqlc`/`swag`.
- **Миграция** `00002_iam` (organizations, users, api_keys; FK, unique email/hash, индексы).
- **Тесты**: table-driven доменные (Role, конструкторы User/Org/ApiKey + невалидные кейсы, детерминизм `HashAPIKey`) и `Login` на фейках (успех / неизвестный email / неверный пароль) — зелёные.
- **Проверка рантайма** (реальный PG 16, сквозной флоу): setup→201, повтор→409, login→JWT, `/auth/me` 200/401, владелец создаёт менеджера→201, дубль email→409, выпуск API-ключа (секрет один раз), `/api/v1/ping` по `X-API-Key` 200, неверный/без ключа→401, **менеджер на owner-only→403**, неверный пароль→401.

### Фаза 3 — Catalog + CRM ✅
Два модуля по skill `new-module`, плюс shared-kernel `Money`.
- **shared kernel `internal/shared/money`**: value object `Money` на `shopspring/decimal` (чистый — только stdlib + decimal), нормализация до 2 знаков, `Cents() int64` для будущей графиковой математики, сериализация строкой; методы `Add/Sub/Cmp/Equals/IsPositive/...` с проверкой валюты. Лежит в shared kernel, т.к. его используют и catalog, и financing — чтобы не связывать модули (`catalog → financing`). Table-driven тесты (parse, Cents, round-trip, валютный mismatch, округление).
- **catalog**: `Product` (+ VO `HalalStatus` halal/haram/doubtful — обязателен), инвариант `CostPrice > 0`, `IsHaram()/CanBeFinanced()` (харам в каталоге допустим, к финансированию — нет); CRUD (`Create/Get/List/Update`), всё scoped по организации; деньги ↔ `numeric(18,2)`.
- **crm**: `Client` (ФИО обязательно, телефон/документ опциональны), CRUD scoped по организации.
- **platform**: новый `platform/pgconv` (UUID/Timestamp/Numeric ↔ домен, детект unique-violation) — переиспользуют все репозитории.
- **http**: оба модуля под `/api/app` (JWT, обе роли); деньги на границе — строка-decimal; централизованный маппинг ошибок.
- **Миграции** `00003_catalog` (products: `cost_price numeric(18,2) CHECK >0`, `halal_status CHECK IN`, FK, индекс), `00004_crm` (clients: FK, индекс).
- **Тесты**: доменные table-driven (HalalStatus, конструкторы Product/Client + невалидные кейсы, `Update` сохраняет identity) — зелёные.
- **Проверка рантайма** (реальный PG 16): CRUD товаров и клиентов; round-trip денег `85000.00 → 90000.00` через `numeric`; `can_be_financed` true/haram→false; изоляция по организации (несуществующий → 404); невалидные входы (`cost_price=abc/0`, `halal_status=mushbooh`, пустое ФИО) → 400; без токена → 401.

### Фаза 4 — Financing (★ ядро, мурабаха) ✅
Модуль `internal/modules/financing` строго по skill `murabaha-engine`. Это финансовая корректность всего продукта.
- **domain** (стерилен): агрегат `Contract` (приватные поля, мутации только через методы), VO `Markup` (наценка суммой **или** из % — фиксируется суммой), `Cadence`, `ContractStatus` (state machine Draft→Active→Completed/Cancelled), `InstallmentStatus`; `Installment`/`Payment`; `CharityEntry` (садака — отдельный агрегат). `Money` берётся из shared kernel.
- **Округление** детерминированное в копейках (`int64`): `base = total/N`, остаток — на **ранние** доли; `Σ == FinancedAmount` ровно. Инвариант `FinancedAmount >= Installments`.
- **Инварианты 1–8** в конструкторе/методах; невалидный договор не создаётся.
- **Статусы долей производны** от `paid = FinancedAmount − Outstanding` (не хранятся): `Paid`/`PartiallyPaid` (ровно одна на «фронте»)/`Pending`/`Overdue`.
- **Анти-риба**: просрочка не меняет `SalePrice`/`Outstanding`/график; садака — **фиксированная**, в отдельный реестр, **не** в долг и **не** в выручку; начисляет **только владелец**. Досрочное погашение без штрафа.
- **`Preview`** — чистый расчёт (график/цена/сравнение «без рибы vs кредит») без записи; общий код с `NewContract` → **preview == создание**.
- **app**: `PreviewContract`, `CreateContract` (create+activate атомарно, блок Haram, проверка client/product через кросс-модульные порты `ProductReader`/`ClientReader`), `RegisterPayment`, `SettleEarly`, `CancelContract`, `AccrueLateCharity`, `Get/ListContracts`, `ListCharity` (транзакции через `TxManager`).
- **infra**: sqlc-репозитории (contract-агрегат: contracts+installments+payments в одной tx; charity), деньги ↔ `numeric(18,2)`, даты ↔ `date`; адаптеры `ProductReader`/`ClientReader` поверх catalog/crm (шов для выноса в микросервис).
- **http** (`/api/app`): `POST /contracts/preview`, CRUD договоров, `…/payments`, `…/settle`; **owner-only** `…/cancel` и `…/charity` (через `iam.OwnerMiddleware`); реестр `GET /charity` (обе роли). Деньги — строка-decimal.
- **Миграция** `00005_financing` (contracts, installments, payments, charity_entries; FK на clients/products/users, CHECK статусов, индексы).
- **Тесты**: **все 11 обязательных table-driven кейсов** skill (ровное деление; деление с остатком 33333.34/.33/.33; нулевой взнос; наценка из %; полное погашение; частичный платёж и статусы; просрочка не меняет долг; садака отдельно от долга; досрочное; preview==создание; недопустимые входы вкл. `FinancedAmount < Installments`, `Payment>Outstanding`, `Payment<=0`) — зелёные, с проверкой «копейки сходятся».
- **Проверка рантайма** (реальный PG 16, сквозной флоу): preview (sale 120000.00 / financed 90000.00 / 6×15000.00 / overpayment 12600.00) == create; харам→409; частичный платёж → outstanding 70000.00, статусы `paid, partially_paid, pending×4`, прогресс 22.22%; переплата→400; **садака (owner) не изменила долг** (100000.00→100000.00), реестр total 500.00; досрочное→completed/0.00; отмена→cancelled; менеджер на charity/cancel→403.

### Фаза 5 — Scheduling (namaz-aware планировщик) ✅
Модуль `internal/modules/scheduling` строго по skill `namaz-scheduler`.
- **domain** (стерилен): порт `PrayerTimesProvider` (домен про библиотеку не знает), `Location`, `PrayerTimes`, `Policy` (буферы, окно джума), `PrayerWindow`, `ReminderType` (call/delivery/payment_followup), `Reminder` (+ `WasShifted`/`Reason`), `ScheduledTime`; **доменный сервис `Scheduler`** — построение заблокированных окон (5 молитв `[p−before, p+after]`, дефолт after=20мин; пятничный джума 12:30–14:00) и поиск **ближайшего свободного слота вперёд** (детерминированный, завершается; корректная обработка длительности доставки и каскадных сдвигов; точечный слот звонка vs интервал).
- **infra**: адаптер `PrayerProvider` поверх `hablullah/go-prayer` (мазхаб **шафиитский**, метод MWL, координаты Грозного из конфига; годовой график кэшируется) — за портом; sqlc-репозиторий напоминаний.
- **app**: `ScheduleReminder` (создаёт со сдвигом), `PreviewSlot` (предложить слот без записи — для календаря), `ListReminders`.
- **http** (`/api/app`): `POST /schedule/reminders`, `POST /schedule/preview`, `GET /schedule/reminders` (обе роли). Конфиг намаза вынесен в env (`PRAYER_*`).
- **Миграция** `00006_scheduling` (reminders; FK на clients/contracts, CHECK типа, индекс).
- **Тесты**: **все 7 обязательных кейсов** skill с фиктивным `PrayerTimesProvider` (вне окон; внутри окна молитвы; доставка с длительностью задевает окно; каскадный сдвиг; пятница-джума; не-пятница; граничное касание) — зелёные.
- **Проверка рантайма** (реальный PG 16 + **настоящий** go-prayer): пятница 13:00→**14:00** (джума); понедельник 10:00→без сдвига; понедельник 19:45 → реальный Магриб Грозного ≈19:39 → **19:59** (Магриб+20мин); создание доставки со сдвигом сохранено; неверный тип→400; список работает.

### Фаза 6 — Публичный API (+3 за расширяемость) ✅
Слой `internal/publicapi/v1` — **тонкий** транспорт поверх тех же app-сервисов financing (логика не дублируется), авторизация по `X-API-Key`.
- **Эндпоинты**: `POST /api/v1/contracts` (создать договор по ссылкам на существующие `client_id`/`product_id` + условия `cost_price`, `markup` (суммой), `down_payment`, `installments`, `cadence`, `start_date`; деньги — строка-decimal); `GET /api/v1/contracts/{id}/payments` (статус, остаток, прогресс, график долей со статусами, платежи).
- **Переиспользование**: `financing.Module` отдаёт use-cases `CreateContract`/`GetContract`, публичный слой их вызывает (не свои). Middleware `X-API-Key` — `iam.APIKeyMiddleware` (org берётся из ключа через `authctx`).
- **OpenAPI/Swagger**: swag-аннотации на хендлерах + общий info в `cmd/api/main.go`; `make swag` генерит `api/openapi/swagger.json|yaml` (только json+yaml, без `docs.go`); spec эмбедится и **Swagger UI на `/swagger/` документирует публичный API** (схемы с примерами, `securityDefinitions: ApiKeyAuth`).
- **Проверка рантайма** (реальный PG 16): создание договора «от имени внешнего маркетплейса» по `X-API-Key` → 201 (active, sale 120000.00, остаток 90000.00, 6 долей); без ключа→401; харам-товар→409; `GET …/payments`→график 6 долей (pending); неверный ключ→401; `/swagger/doc.json` содержит реальный `/contracts`, `/swagger/` отвечает 200.
