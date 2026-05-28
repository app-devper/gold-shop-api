# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This is one subproject inside the `/Users/admin/ProjectPos` workspace — the workspace-level [`CLAUDE.md`](../../CLAUDE.md) covers cross-service rules (auth hub `um-api`, JWT/Redis session contract, role conventions, **no-code-comments rule**). Read this file together with that one.

## Common commands

Run from this directory ([gold-shop-api/](.)):

```bash
go mod download
go run main.go                                 # binds :SERVER_PORT (8085 in .env, 8080 default)
go test ./...                                  # full suite
go test ./app/feature/sale                     # single package
go test -run TestCreateSale ./app/feature/sale # single test
gofmt -w <file>
```

OpenAPI spec lives at [docs/openapi.yaml](docs/openapi.yaml); preview with `npx @redocly/cli preview-docs docs/openapi.yaml`.

## Configuration

Loaded by [app/config/config.go](app/config/config.go) from `.env` via `godotenv`. Required by the running service: `SERVER_PORT`, `SERVER_ENV`, `MONGODB_URI`, `MONGODB_DATABASE` (the **DB prefix** — see Multi-tenant below), `REDIS_HOST` (Redis URL — `redis://...`), `SECRET_KEY` (HMAC for JWTs from `um-api`), `CLIENT_ID`, `SYSTEM` (defaults to `GOLD`), `GOLD_API_URL` (Thai gold price API).

`MONGODB_DATABASE` is the database name for the default tenant (`clientId="000"`); other tenants get a database named `<MONGODB_DATABASE>_<clientId>` on the same cluster. The env name is preserved for backward compatibility — internally it is stored as `Config.MongoDB.DBPrefix`.

`JWT_SECRET` is loaded into `Config.JWT.Secret` and is required when `SERVER_ENV=production` — but the actual JWT verification uses `cfg.Auth.SecretKey` (`SECRET_KEY`), which is what must match `um-api`. The `JWT_SECRET` check is currently dead config; don't rely on it for auth.

## Architecture

Clean Architecture, but the directory naming differs from sibling Go services in this workspace:

- `app/feature/<domain>/` — **not** `app/featues/` (the typo-preserving name used by `pos-api` / `snook-api` / `um-api`). Don't try to "fix" sibling services to match this one or vice versa.
- Each feature has a `service.go` and (often) `service_test.go`. There is no per-feature handler — HTTP handlers live in [app/infrastructure/http/handler/](app/infrastructure/http/handler/), routes in [app/infrastructure/http/router/router.go](app/infrastructure/http/router/router.go), middleware in [app/infrastructure/http/middleware/](app/infrastructure/http/middleware/).
- Domain entities are in [app/domain/entity/](app/domain/entity/); repository interfaces in [app/domain/repository/interfaces.go](app/domain/repository/interfaces.go); Mongo/Redis implementations in [app/infrastructure/mongo/](app/infrastructure/mongo/) and [app/infrastructure/redis/](app/infrastructure/redis/).
- DI wiring (and the only place repos/services/handlers are constructed) is [app/init.go](app/init.go) — `App.StartApp`.
- `main.go` is a 5-line entry point.

### Multi-tenant (DB per clientId)

One instance serves many shops. The Mongo `Client` ([app/infrastructure/mongo/client.go](app/infrastructure/mongo/client.go)) wraps a single `*mongo.Client` plus a per-tenant `*mongo.Database` cache keyed by `clientId`. DB naming: `clientId="000"` → `<dbPrefix>` (e.g. `gold_shop`); others → `<dbPrefix>_<clientId>` (e.g. `gold_shop_test1`). Existing single-tenant deployments on tenant `000` continue working without any data migration.

- Repositories are constructed once at startup but resolve their collection per-request via `Client.CollectionFromCtx(ctx, name)`, which reads the tenant from ctx and looks up the cached DB. **Repos must not be given a `*mongo.Collection` at construction time** — it would leak to whatever DB was active first.
- Tenant init runs lazily on first access (see [app/seed.go](app/seed.go)): seeds HQ branch (so the `RequireBranch` fallback works), seeds initial GoldPrice (each shop sets its own price — no shared price collection), and creates the unique `userId` + `branchId` indexes on `employees`. A failed seed is logged and **retried on the next request** for that tenant (per-tenant mutex + `atomic.Bool` flag), so a transient Mongo blip doesn't permanently lock a tenant out.
- Counters (`sale-*`, `pawn-*`) live in each tenant's database, so per-day sequences are isolated automatically — branch 01 of shop A and branch 01 of shop B both start at `S2605030001` on the same day. No global counter prefix is needed.
- Mongo transactions still work — `client.StartSession` is client-level and works across any DB in the same cluster as long as all repo calls inside `fn(sessCtx)` read tenant from the same ctx.
- **Don't write to Redis from this service** — `session:<id>` is shared globally with `um-api` and is keyed by session ID (already globally unique), so there's no tenant collision there.

### Auth & middleware chain

All `/api/v1` routes go through (in [router.go:46-51](app/infrastructure/http/router/router.go#L46-L51)):

1. `RequireAuthenticated(secretKey)` — validates HMAC-SHA256 JWT from `um-api`, sets `SessionId`, `Role` (UM role from JWT), `System`, `ClientId`.
2. `RequireTenant()` — validates `ClientId` against `mongo.ValidateClientID` (rejects empty / unsafe DB-name characters with 401) and copies it from gin ctx into the request's `context.Context` via `mongo.WithClientID`. Must run **before** any middleware/handler that touches Mongo.
3. `RequireSession(sessionRepo)` — looks up `SessionId` in Redis under key `session:<id>`, parses JSON for `userId`, sets `UserId`. See [redis/session_repo.go](app/infrastructure/redis/session_repo.go).
4. `RequireBranch(employeeRepo, branchRepo)` — resolves the employee's branch and local role into context. **Falls back to `HQ` branch with `STAFF` role only when `entity.ErrNotFound` is returned**; any other error is a 500. Don't widen this fallback — it would mask data-layer failures.

Per-route role gates layer on top:
- `RoleMiddleware(EmployeeRoleAdmin, ...)` — checks the **employee role** from `RequireBranch` (`ADMIN` / `MANAGER` / `STAFF`).
- `RequireRole(UMRoleSuper, UMRoleAdmin)` — checks the **UM role** from the JWT (`SUPER` / `ADMIN` / `USER`).

Used differently on purpose: employee CRUD uses `RequireRole` (UM-level), most domain mutations use `RoleMiddleware` (employee-level). See [app/domain/entity/employee.go](app/domain/entity/employee.go) for the constants.

### Bootstrap & seed data

Seeding is **per-tenant and lazy** — no longer runs at startup. The first request for any `clientId` calls the seeder from [app/seed.go](app/seed.go) via `mongo.Client.ForClient`:
- Creates `HQ` branch (`สำนักงานใหญ่`) if missing — **the `RequireBranch` fallback depends on this branch existing**.
- Creates an initial `GoldPrice` if none exists, fetched from `GOLD_API_URL`, falling back to hardcoded values (`Bar 42350/42450`, `Ornament 41850/42950`) if the API call fails.
- Creates the `userId`-unique + `branchId` indexes on `employees`.

Seed failures are logged as warnings and the seeder is **re-attempted on the next request** for that tenant (a per-tenant mutex serializes concurrent first requests; an `atomic.Bool` shortcuts subsequent successful calls without contention). Each individual seed step (`seedHQBranch`, `seedInitialGoldPrice`) checks `count > 0` first, so retries after a partial success are idempotent.

### Transactions

`repository.TransactionManager` ([interfaces.go:11-15](app/domain/repository/interfaces.go#L11-L15)) abstracts atomic ops. `mongo.MongoTransactionManager` uses `session.WithTransaction`; `testutils.MockTransactionManager` just invokes `fn(ctx)` directly so unit tests don't need a Mongo replica set.

The **sale create** flow ([feature/sale/service.go](app/feature/sale/service.go)) is the canonical pattern and worth understanding before editing other mutating services:

- **Phase 1 (no DB writes)**: validate every item, look up products/items, check branch ownership and status, compute totals into a `[]resolvedItem`. Validate customer points budget *before* the transaction.
- **Phase 2 (inside `WithTransaction`)**: generate the sale number (so it rolls back on failure), persist sale, then for each resolved item either flip a piece-based `ProductItem` to `Sold` or atomically `DeductWeight`; write a `StockLog` entry; update customer points/spending.
- **Cancel** is the inverse: piece items go back to `Available`, weight items get `AddWeight`, stock logs record `cancel`, customer points/spend are reversed (capped at 0). See [service.go:399-473](app/feature/sale/service.go#L399-L473).

Inventory transfers and reward redemption follow the same Phase 1 / Phase 2 pattern via the same `TransactionManager`.

### Sale-number / pawn-number / account-number generation

Atomic via the per-tenant `counters` collection ([mongo/counter_repo.go](app/infrastructure/mongo/counter_repo.go)) — `FindOneAndUpdate` with `$inc` and `upsert`. Generation runs *inside* the transaction so a failed sale doesn't burn a number gap (in practice gaps still happen on Mongo retries, but they're rare). Because `counters` lives in each tenant's database, sequences cannot collide across shops.

## Domain rules worth knowing before editing

- **Money & weight rounding** — always round through [pkg/utils/money.go](pkg/utils/money.go): `RoundBaht` (2dp, satang) for currency, `RoundGram` (6dp) for gold weight. `RoundGram`'s 6-decimal precision exists specifically to kill floating-point tail noise from `amount/goldPrice` divisions while preserving sub-milligram increments. Don't introduce ad-hoc rounding.
- **Gold-type price math** — see [entity/product.go](app/domain/entity/product.go). `BahtPerGramOrnament = 15.16` (96.5% gold), `BahtPerGramBar = 15.244` (99.99% gold). `Product.IsBarGold()` matches `"99.99%"` or `"99.99"` (both forms exist in the wild). Per-gram sell price is `goldPrice.GoldXxxSell / BahtPerGramXxx` ([sale/service.go:102-107](app/feature/sale/service.go#L102-L107)). Manual `Price` on a sale item overrides the dynamic calc.
- **Stock types** — `Product.StockType` is either `piece` or `weight`; empty string defaults to `weight` for legacy rows (`DefaultStockType()`). Piece-based products require a `ProductItemID` per sale line and decrement by flipping that item's status; weight-based use atomic `DeductWeight`/`AddWeight` on the product row.
- **Pawn interest** ([entity/pawn.go](app/domain/entity/pawn.go)) — accrues in **30-day blocks**, not calendar months. Partial 30-day periods accrue zero. `monthsBetween(start, end) = (end.Sub(start).Hours()/24) / 30`. `CalculateTotalInterestDue` measures from the *last paid period's `PeriodTo`*, not from `StartDate`, when payments exist.
- **Gold savings** ([entity/gold_saving.go](app/domain/entity/gold_saving.go)) — `SavingType` is `money` (deposit cash, convert to gold weight at current price) or `weight` (deposit gold directly). Withdraw can be `asCash` (deduct gold weight equivalent + decrement cash balance) or physical gold. `CashBalance` is floored at 0 on withdraw.
- **Customer points** — sales accrue 1 point per 100 baht of `NetTotal`; cancellation restores `PointsUsed` and removes `PointsEarned` (capped at 0 if the customer already spent them).
- **Sale number format** (SRS 6.2) — `<prefix>{YYMMDD}{XXXX}` where prefix encodes `SaleType`: `S` = sell, `B` = buy_old, `TR` = exchange. The 4-digit sequence is per-day, per-prefix, per-branch, generated atomically inside the sale transaction via the `counters` collection (`saleNumberPrefix()` in [mongo/sale_repo.go](app/infrastructure/mongo/sale_repo.go)). Never reuse or assume continuity across days/prefixes — gaps can occur on Mongo retries.
- **Trade-in audit fields** — `OldGoldItem.Condition` (`good`/`fair`/`damaged`, SRS 3.7) and `Sale.OldItemDestination` (`melt`/`resell`/`scrap`) are recorded for audit/inventory routing only. Neither participates in price math; the buyback amount is `weight × price_per_unit × (1 − deduction_percent/100)`. Set `OldItemDestination` for `buy_old` and `exchange`; leave empty for `sell`.
- **Product category** ([entity/product.go](app/domain/entity/product.go)) — `Product.Category` is an ornament-only enum (`necklace`/`bracelet`/`ring`/`bangle`/`earring`/`pendant`/`amulet`, SRS 3.2). The service rejects categories on `kind=bar`. Validate via `entity.IsValidProductCategory`. The legacy `category_id` reference to a master `ProductCategory` collection still exists alongside this enum but the enum is the SRS-aligned source of truth for ornament classification.

## Cross-service awareness

- **Don't change `SECRET_KEY` semantics or session-key format** without also updating `um-api` — JWT signing and the `session:<id>` Redis key shape are a contract owned by `um-api`.
- This service shares Redis with all other workspace services. The only key it reads is `session:<id>`. Don't write to Redis from this service.
- Mongo cluster is shared across tenants of this service; if another service shares the same `MONGODB_URI`, make sure its DB names don't collide with the `<dbPrefix>_<clientId>` pattern used here (default prefix `gold_shop`).
- API base path is `/api/v1` (note: gold uses `/api/v1`, not `/api/gold/v1` like the sibling services).
