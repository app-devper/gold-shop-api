# Gold Shop API

A robust backend API for a Gold Shop Management System, built with Go using Clean Architecture principles.

## 🚀 Technologies

- **Language:** Go 1.21+
- **Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** MongoDB
- **Authentication:** JWT (JSON Web Tokens)
- **Middleware:** CORS, Logger, Recovery, Auth, Role-based Access Control

## 📁 Project Structure

The project follows Clean Architecture:
- `cmd/server`: Application entry point.
- `config`: Configuration management.
- `internal/application`: Application services and business logic.
- `internal/domain`: Core entities and repository interfaces.
- `internal/infrastructure`: External implementations (HTTP handlers, MongoDB repositories, external clients).
- `pkg`: Shared utilities and helper packages.

## ⚙️ Configuration

Create a `.env` file in the root directory with the following variables:

```env
SERVER_PORT=8080
SERVER_ENV=development
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=gold_shop
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION_HOURS=24
GOLD_API_URL=https://api.chnwt.dev/thai-gold-api/latest
```

## 🛣️ API Endpoints

All API endpoints are prefixed with `/api/v1`.

### 🔐 Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout (protected)
- `POST /auth/refresh` - Refresh JWT token (protected)
- `GET /auth/me` - Get current user profile (protected)
- `PUT /auth/password` - Change password (protected)

### 🏢 Branches
- `GET /branches` - List all branches
- `GET /branches/:id` - Get branch details
- `POST /branches` - Create branch (Admin)
- `PUT /branches/:id` - Update branch (Admin)
- `DELETE /branches/:id` - Delete branch (Admin)

### 💰 Gold Prices
- `GET /gold-prices/current` - Get current gold prices
- `GET /gold-prices/history` - Get gold price history
- `POST /gold-prices` - Set manual gold price (Admin/Manager)
- `POST /gold-prices/sync` - Sync gold price from external API (Admin/Manager)

### 🛒 Sales & POS
- `GET /sales` - List sales
- `GET /sales/unpaid` - Get unpaid sales
- `GET /sales/:id` - Get sale details
- `POST /sales` - Record new sale
- `POST /sales/:id/cancel` - Cancel sale
- `GET /sales/:id/receipt` - Generate receipt

### 💍 Pawns (Gold Pawning)
- `GET /pawns` - List pawn tickets
- `GET /pawns/due-soon` - Get pawns due for interest payment/redemption
- `POST /pawns` - Create new pawn ticket
- `POST /pawns/:id/pay-interest` - Record interest payment
- `POST /pawns/:id/redeem` - Redeem pawned item
- `POST /pawns/:id/extend` - Extend pawn term
- `POST /pawns/:id/forfeit` - Forfeit pawned item

### 🏦 Gold Savings
- `GET /gold-savings` - List saving accounts
- `POST /gold-savings` - Open new saving account
- `POST /gold-savings/:id/deposit` - Deposit into account
- `POST /gold-savings/:id/withdraw` - Withdraw from account
- `POST /gold-savings/:id/close` - Close account
- `GET /gold-savings/:id/statement` - Get account statement

### 📦 Inventory
- `GET /inventory/transfers` - List inventory transfers
- `POST /inventory/transfers` - Create transfer request
- `POST /inventory/transfers/:id/approve` - Approve transfer
- `POST /inventory/transfers/:id/receive` - Record receipt of items
- `POST /inventory/transfers/:id/cancel` - Cancel transfer

### 🎁 Rewards
- `GET /rewards` - List available rewards
- `POST /rewards` - Create new reward (Admin/Manager)
- `POST /rewards/redeem` - Redeem reward using points

### 📉 Reports
- `GET /reports/dashboard` - Dashboard summary data
- `GET /reports/profit-loss` - Profit and Loss report
- `GET /reports/top-products` - Top selling products
- `GET /reports/employee-performance` - Employee sales performance
- `GET /reports/trends` - Sales trends over time

## 🏃 Getting Started

1.  **Clone the repository:**
    ```bash
    git clone git@github.com:app-devper/gold-shop-api.git
    cd gold-shop-api
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Setup environment variables:**
    Copy `.env.example` to `.env` (if provided) or create one manually.

4.  **Run the application:**
    ```bash
    go run main.go
    ```
    The server will start on default port `8080`.

## 📜 License

This project is licensed under the MIT License.
