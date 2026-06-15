---
name: murabaha-engine
description: Use this skill for ANY work touching the financing core — creating installment contracts, building payment schedules, registering payments, computing outstanding balance, handling late payments, or early settlement. It defines the exact murabaha (cost-plus) math, the riba-prohibition invariants, deterministic rounding, the late-fee-as-charity mechanism, money handling, and the required table-driven tests. This is the financial correctness of the whole product — do not improvise; implement exactly as specified and cover every rule with tests.
---

# Skill: движок мурабахи (рассрочка без рибы)

Мурабаха — продажа товара по схеме «себестоимость + фиксированная наценка» с оплатой
в рассрочку. Ключевое отличие от кредита: **наценка фиксируется суммой один раз**, и
сумма долга **не зависит от времени**. Любой механизм, при котором долг растёт из-за
просрочки, — это риба и запрещён.

## Входные данные договора

- `CostPrice` (закупочная цена) — `Money`, `> 0`.
- `Markup` (наценка) — `Money`, `>= 0`. Может задаваться суммой или как процент от
  `CostPrice`, но после расчёта это **фиксированная сумма**, не ставка.
- `DownPayment` (первый взнос) — `Money`, `0 <= DownPayment < SalePrice`.
- `Installments` (число платежей) — `int`, `>= 1`.
- `Cadence` (периодичность) — напр. месяц. Влияет только на даты, не на суммы.
- `StartDate` — дата первого планового платежа.

## Формулы

```
SalePrice      = CostPrice + Markup                  // фиксируется при создании
FinancedAmount = SalePrice − DownPayment
Base           = FinancedAmount / Installments       // деление с остатком (см. округление)
```

Итоговый график: `Installments` плановых платежей с шагом `Cadence`, начиная со `StartDate`.

## Округление (детерминированное, без потери копеек)

Работаем в **наименьших единицах (копейки, int64)**, не во float.

1. `total := FinancedAmount` в копейках.
2. `base := total / Installments` (целочисленно), `remainder := total % Installments`.
3. Каждый платёж = `base`. Затем **первые `remainder` платежей** получают +1 копейку.
4. Гарантия: сумма всех платежей == `total` ровно. Проверяй это ассертом в тесте.

> Остаток распределяем на **ранние** платежи (детерминированно). Не «размазываем
> случайно» и не «суём всё в последний» — фиксируем одно правило и тестируем его.
> Условие `FinancedAmount >= Installments` (инвариант 3) гарантирует `base >= 1` —
> нулевых долей не возникает.

## Инварианты (обязательны в конструкторе/методах агрегата)

1. `SalePrice == CostPrice + Markup`.
2. `0 <= DownPayment < SalePrice`.
3. `FinancedAmount >= Installments` — каждая доля получает **минимум 1 копейку** (нулевых
   долей нет). Иначе конструктор возвращает типизированную ошибку.
4. `Σ Installment == FinancedAmount` (после округления — ровно).
5. `DownPayment + Σ Installment == SalePrice`.
6. `Outstanding` инициализируется как `SalePrice − DownPayment`, не возрастает,
   не уходит в минус.
7. `SalePrice`, график и `Outstanding` **не меняются от времени/просрочки**.
8. `0 < Payment.Amount <= Outstanding` на момент регистрации.

Если любой инвариант нарушается — конструктор/метод возвращает типизированную ошибку,
объект не создаётся / состояние не меняется.

## Статусы доли — производны от `Outstanding`

`Outstanding` — единственный источник правды по прогрессу. Статус каждой `Installment` —
**производный** от накопленной фактической оплаты (`paid := FinancedAmount − Outstanding`),
а не хранится независимо. Для доли с накопительной верхней границей `cum`:

- `Paid` — `paid >= cum` (доля полностью покрыта).
- `PartiallyPaid` — граница предыдущей доли `< paid < cum` (ровно одна такая доля).
- `Pending` — `paid <= нижняя граница` и срок не наступил.
- `Overdue` — не покрыта и срок прошёл (выставляет `HasOverdueInstallment`).

## Предпросмотр графика (без сохранения)

`PreviewContract(input) -> {Schedule, SalePrice, FinancedAmount, ComparisonData}` —
**чистый расчёт** по тем же формулам и округлению, **без создания агрегата и без записи в
БД**. Используется мастером договора (`POST /api/app/contracts/preview`). Клиент график не
считает — он его только отображает.

## Регистрация платежа

```
RegisterPayment(amount):
  require status == Active
  require 0 < amount <= Outstanding
  Outstanding -= amount
  // статусы долей пересчитываются из paid = FinancedAmount − Outstanding:
  // полностью покрытые → Paid, одна «на фронте оплаты» → PartiallyPaid, прочие → Pending/Overdue
  if Outstanding == 0: status = Completed
```

Сумма платежа — **любая** в пределах `0 < amount <= Outstanding`; кратность доле не требуется.

## Просрочка (анти-риба механика)

- При наступлении срока без оплаты `Installment` помечается `Overdue`.
- **Сумма долга не меняется.** `Outstanding` и `SalePrice` остаются прежними.
- Если в настройках бизнеса включён сбор за просрочку — он **фиксированный**
  (`LatePenaltyCharity`, напр. 500 ₽), создаётся как отдельная запись с флагом
  «благотворительность» (действие владельца). Он:
  - не зависит от длительности просрочки (никакого «штраф за каждый день»);
  - не прибавляется к `Outstanding` и не входит в график;
  - учитывается в отдельном реестре садаки, **не** в выручке продавца.
- Запрещено: процент на просрочку, капитализация, пеня, растущая со временем.

## Досрочное погашение

```
SettleEarly():
  require status == Active
  принять оплату Outstanding целиком
  Outstanding = 0
  status = Completed
  // штрафа за досрочное погашение НЕТ
```

## Деньги

- Тип `Money` (см. CLAUDE.md). Внутренние расчёты — копейки `int64`. **Хранение** —
  `NUMERIC(18,2)`. **На границе API** — строка-decimal (`"120000.00"`). Никаких операций
  `+ - * /` над деньгами вне методов `Money`. Ни одного `float`.

## Обязательные table-driven тесты

Покрыть минимум:

1. **Ровное деление:** Cost=100000, Markup=20000, Down=30000, N=6 →
   Sale=120000, Financed=90000, каждый платёж=15000, Σ=90000, Σ+Down=120000.
2. **Деление с остатком:** Financed=100000.00, N=3 → платежи 33333.34 / 33333.33 /
   33333.33, Σ=100000.00 (первый платёж получил лишнюю копейку).
3. **Нулевой первый взнос:** Down=0 → Financed=Sale, график на всю сумму.
4. **Наценка из процента:** Markup = 10% от Cost — проверить, что зафиксировалась
   суммой и дальше не пересчитывается.
5. **Полное погашение платежами:** последовательные RegisterPayment → Outstanding=0,
   status=Completed.
6. **Частичный платёж:** платёж, не кратный доле → `Outstanding` уменьшился ровно на сумму,
   полностью покрытые доли = `Paid`, ровно одна доля = `PartiallyPaid`, остальные `Pending`.
7. **Просрочка не меняет долг:** пометить Installment Overdue → Outstanding и Sale
   не изменились.
8. **Сбор за просрочку — садака:** начислен фиксированный сбор → он в реестре садаки,
   не в Outstanding, не в выручке.
9. **Досрочное погашение:** SettleEarly при остатке → Outstanding=0, без штрафа.
10. **Предпросмотр == создание:** `PreviewContract(input).Schedule` совпадает с графиком
    реально созданного `Contract(input)` (тот же расчёт, без записи).
11. **Недопустимые входы (ожидаем ошибку):** Down >= Sale; N < 1; Cost <= 0;
    Markup < 0; **FinancedAmount < Installments** (нулевая доля); Payment > Outstanding;
    Payment <= 0.

Каждый тест-кейс проверяет точные суммы и инвариант «копейки сходятся».
