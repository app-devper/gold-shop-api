# Gold Shop API

A robust backend API for a Gold Shop Management System, built with Go using Clean Architecture principles.

## 🚀 Technologies

- **Language:** Go 1.21+
- **Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** MongoDB
- **Session Store:** Redis
- **Authentication:** JWT issued by external `um-api`, validated via Redis session
- **Middleware:** CORS, Logger, Recovery, Auth, Session, Role-based Access Control

## 📁 Project Structure

```
.
├── main.go                         # Entry point
├── app/
│   ├── init.go                     # App bootstrap (config, DI, router)
│   ├── config/                     # Configuration loading
│   ├── domain/
│   │   ├── entity/                 # Core domain entities & errors
│   │   └── repository/             # Repository interfaces
│   ├── feature/                    # Feature modules (handler, service, per domain)
│   │   ├── branch/
│   │   ├── employee/
│   │   ├── sale/
│   │   ├── pawn/
│   │   ├── gold_saving/
│   │   ├── gold_price/
│   │   ├── product/
│   │   ├── customer/
│   │   ├── inventory/
│   │   ├── reward/
│   │   ├── expense/
│   │   └── report/
│   └── infrastructure/
│       ├── http/
│       │   ├── middleware/         # Auth, session, branch, CORS, logger
│       │   └── router/             # Route definitions
│       ├── mongo/                  # MongoDB repository implementations
│       └── redis/                  # Redis session repository
├── pkg/                            # Shared utilities (jwt, utils)
└── docs/
    └── openapi.yaml                # OpenAPI 3.0 specification
```

## ⚙️ Configuration

Create a `.env` file in the root directory:

```env
SERVER_PORT=8080
SERVER_ENV=development

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=gold_shop

REDIS_HOST=localhost:6379

SECRET_KEY=your-jwt-secret-key
CLIENT_ID=your-client-id
SYSTEM=gold-shop

GOLD_API_URL=https://api.chnwt.dev/thai-gold-api/latest
```

| Variable | Description |
|---|---|
| `SERVER_PORT` | HTTP listen port (default `8080`) |
| `MONGODB_URI` | MongoDB connection string |
| `MONGODB_DATABASE` | MongoDB database name |
| `REDIS_HOST` | Redis address (`host:port`) |
| `SECRET_KEY` | HMAC secret used to verify JWTs issued by `um-api` |
| `CLIENT_ID` | Client identifier embedded in JWT claims |
| `SYSTEM` | System identifier embedded in JWT claims |
| `GOLD_API_URL` | External gold price API URL |

## 🔐 Authentication

All `/api/v1` endpoints require a valid JWT Bearer token.

**Flow:**
1. Client authenticates against the external **`um-api`** service and receives a JWT.
2. Client sends `Authorization: Bearer <token>` on every request.
3. `RequireAuthenticated` middleware validates the JWT signature (HMAC-SHA256) and checks `clientId` / `system` claims.
4. `RequireSession` middleware looks up the session ID in **Redis** to confirm the session is still active.
5. `RequireBranch` middleware resolves the employee's branch and role from MongoDB. Employees not registered default to the `HQ` branch with `STAFF` role.

**Roles:**

| Role | Permissions |
|---|---|
| `ADMIN` | Full access to all endpoints |
| `MANAGER` | Gold prices, products, product categories, reports, expenses |
| `STAFF` | Sales, pawns, gold savings, customers, inventory (default) |

## 📖 API Spec

The full OpenAPI 3.0.3 specification is located at [`docs/openapi.yaml`](docs/openapi.yaml).

To browse interactively with Swagger UI:

```bash
npx @redocly/cli preview-docs docs/openapi.yaml
```

Or use the [Swagger Editor](https://editor.swagger.io/) — paste the contents of `docs/openapi.yaml`.

## 🛣️ API Endpoints

All endpoints are prefixed with `/api/v1` and require `Authorization: Bearer <token>`.

### 🏢 Branches
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/branches` | Any | List all branches |
| `GET` | `/branches/:id` | Any | Get branch by ID |
| `POST` | `/branches` | ADMIN | Create branch |
| `PUT` | `/branches/:id` | ADMIN | Update branch |
| `DELETE` | `/branches/:id` | ADMIN | Delete branch |

### 👤 Employees
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/employees` | Any | List all employees |
| `GET` | `/employees/branch/:branchId` | Any | List employees by branch |
| `GET` | `/employees/:id` | Any | Get employee by ID |
| `POST` | `/employees` | ADMIN | Create employee |
| `PUT` | `/employees/:id` | ADMIN | Update employee |
| `DELETE` | `/employees/:id` | ADMIN | Delete employee |

### 💰 Gold Prices
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/gold-prices/current` | Any | Get current prices |
| `GET` | `/gold-prices/history` | Any | Get price history |
| `POST` | `/gold-prices` | ADMIN, MANAGER | Set price manually |
| `POST` | `/gold-prices/sync` | ADMIN, MANAGER | Sync from external API |

### 🛒 Sales
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/sales` | Any | List sales |
| `GET` | `/sales/unpaid` | Any | Get unpaid sales |
| `GET` | `/sales/:id` | Any | Get sale by ID |
| `POST` | `/sales` | Any | Create sale |
| `POST` | `/sales/:id/cancel` | Any | Cancel sale |
| `GET` | `/sales/:id/receipt` | Any | Get receipt |

> **Sale number format** (SRS 6.2): `<prefix>{YYMMDD}{XXXX}` where the prefix encodes `sale_type` — `S` (sell), `B` (buy_old), `TR` (exchange). The 4-digit sequence is per-day, per-prefix, per-branch. For `buy_old` and `exchange`, also send `old_item_destination` (`melt` / `resell` / `scrap`) and per-item `condition` (`good` / `fair` / `damaged`) on each `OldGoldItem`. See OpenAPI for the full schema.

### 💍 Pawns
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/pawns` | Any | List pawn tickets |
| `GET` | `/pawns/due-soon` | Any | Get pawns due soon |
| `GET` | `/pawns/:id` | Any | Get pawn by ID |
| `POST` | `/pawns` | Any | Create pawn ticket |
| `POST` | `/pawns/:id/pay-interest` | Any | Record interest payment |
| `POST` | `/pawns/:id/redeem` | Any | Redeem pawned item |
| `POST` | `/pawns/:id/extend` | Any | Extend pawn term |
| `POST` | `/pawns/:id/forfeit` | Any | Forfeit pawned item |

### 🏦 Gold Savings
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/gold-savings` | Any | List accounts |
| `GET` | `/gold-savings/:id` | Any | Get account by ID |
| `POST` | `/gold-savings` | Any | Open new account |
| `POST` | `/gold-savings/:id/deposit` | Any | Deposit |
| `POST` | `/gold-savings/:id/withdraw` | Any | Withdraw |
| `POST` | `/gold-savings/:id/close` | Any | Close account |
| `GET` | `/gold-savings/:id/statement` | Any | Get statement |

### 📦 Products & Categories
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/product-categories` | Any | List categories |
| `POST` | `/product-categories` | ADMIN, MANAGER | Create category |
| `GET` | `/products` | Any | List products |
| `GET` | `/products/:id` | Any | Get product by ID |
| `POST` | `/products` | ADMIN, MANAGER | Create product |
| `PUT` | `/products/:id` | ADMIN, MANAGER | Update product |
| `DELETE` | `/products/:id` | ADMIN, MANAGER | Delete product |

> **Ornament categories** (SRS 3.2): when creating an ornament product, send the optional `category` enum — one of `necklace`, `bracelet`, `ring`, `bangle`, `earring`, `pendant`, `amulet`. Bar products do not take a category. The `note` field is an internal staff note, separate from the customer-facing `description`.

### 🧑‍🤝‍🧑 Customers
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/customers` | Any | List customers |
| `GET` | `/customers/rfid/:rfid` | Any | Get by RFID card |
| `GET` | `/customers/:id` | Any | Get by ID |
| `POST` | `/customers` | Any | Create customer |
| `PUT` | `/customers/:id` | Any | Update customer |
| `DELETE` | `/customers/:id` | ADMIN, MANAGER | Delete customer |

### � Inventory
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/inventory/transfers` | Any | List transfers |
| `GET` | `/inventory/transfers/:id` | Any | Get transfer by ID |
| `POST` | `/inventory/transfers` | Any | Create transfer request |
| `POST` | `/inventory/transfers/:id/approve` | Any | Approve transfer |
| `POST` | `/inventory/transfers/:id/receive` | Any | Mark as received |
| `POST` | `/inventory/transfers/:id/cancel` | Any | Cancel transfer |

### 🎁 Rewards
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/rewards` | Any | List rewards |
| `POST` | `/rewards` | ADMIN, MANAGER | Create reward |
| `POST` | `/rewards/redeem` | Any | Redeem with points |

### 💸 Expenses
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/expenses` | Any | List expenses |
| `POST` | `/expenses` | Any | Create expense |
| `GET` | `/expenses/categories` | Any | List categories |
| `POST` | `/expenses/categories` | ADMIN | Create category |

### � Reports
| Method | Path | Role | Description |
|---|---|---|---|
| `GET` | `/reports/dashboard` | Any | Dashboard summary |
| `GET` | `/reports/profit-loss` | Any | Profit & loss |
| `GET` | `/reports/multi-branch` | ADMIN | Multi-branch report |
| `GET` | `/reports/top-products` | Any | Top selling products |
| `GET` | `/reports/employee-performance` | Any | Employee performance |
| `GET` | `/reports/trends` | Any | Sales trends |

## 🏃 Getting Started

1. **Clone the repository:**
   ```bash
   git clone git@github.com:app-devper/gold-shop-api.git
   cd gold-shop-api
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Setup environment variables:**
   ```bash
   cp .env.example .env
   # edit .env with your values
   ```

4. **Run the application:**
   ```bash
   go run main.go
   ```
   The server will start on port `8080` (or `SERVER_PORT`).

## 📜 License

This project is licensed under the MIT License.
