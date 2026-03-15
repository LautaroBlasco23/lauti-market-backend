# Order Module Implementation Plan

## Context

The marketplace backend needs an Order module so users can purchase products from stores. Orders are scoped to a single store, track status through a shipping lifecycle, and manage product stock automatically. This is the first module requiring JWT-protected endpoints.

## Requirements Summary

- **Actor**: Only users (accountType='user') place orders
- **Scope**: One order = one store's products
- **Statuses**: pending → confirmed → shipped → delivered → cancelled
- **Stock**: Decrement on creation, restore on cancellation
- **Auth**: JWT middleware on all order endpoints
- **Role-based transitions**: Store confirms/ships/delivers; user cancels (pending only)
- **Endpoints**: Create, get by ID, list by user, list by store, update status

---

## Step 1: Add Order Errors

**File**: `internal/api/domain/errors.go` (modify)

Add: `ErrOrderNotFound`, `ErrInsufficientStock`, `ErrEmptyOrderItems`, `ErrInvalidQuantity`, `ErrInvalidOrderStatus`, `ErrForbiddenTransition`, `ErrUnauthorized`, `ErrForbidden`, `ErrItemsFromMultipleStores`

## Step 2: Create Auth Middleware

**New file**: `internal/api/infrastructure/auth_middleware.go`

- Extracts `Authorization: Bearer <token>` header
- Validates JWT using existing `auth/infrastructure/utils.JWTGenerator.Validate()`
- Injects `*Claims` into request context
- Helper `GetClaims(ctx)` to retrieve claims
- Only applied to order routes (not global)

## Step 3: Domain Layer

**New files** in `internal/order/domain/`:

| File | Contents |
|------|----------|
| `order_status.go` | `OrderStatus` type with constants (pending, confirmed, shipped, delivered, cancelled) and `IsValid()` |
| `order_entity.go` | `Order` entity (id, userID, storeID, status, items, totalPrice, timestamps). Constructor sets status=pending. `TransitionStatus(newStatus, accountType, accountID)` enforces the state machine |
| `order_item.go` | `OrderItem` value object (id, orderID, productID, quantity, unitPrice). Validated constructor |
| `order_repository.go` | `Repository` interface: `Save`, `FindByID`, `FindByUserID`, `FindByStoreID`, `UpdateStatus` |

**State machine rules** (in `TransitionStatus`):
- pending → confirmed: store only (must match storeID)
- confirmed → shipped: store only
- shipped → delivered: store only
- pending → cancelled: user only (must match userID)
- All others → `ErrForbiddenTransition`

## Step 4: Application Layer

**New file**: `internal/order/application/order_service.go`

Dependencies: `order.Repository`, `product.Repository`, `IDGenerator`

**CreateOrder(ctx, input)**:
1. Validate items non-empty
2. For each item: fetch product, verify `product.StoreID() == input.StoreID`, verify sufficient stock
3. Compute totalPrice from `product.Price() * quantity` (snapshot)
4. Build Order + OrderItems, call `repo.Save()` (transactional, decrements stock atomically)

**UpdateStatus(ctx, input)**:
1. Fetch order
2. Call `order.TransitionStatus()` (domain enforces rules)
3. If cancelling: restore stock via `productRepo` for each item
4. Call `repo.UpdateStatus()`

**GetByID, GetByUserID, GetByStoreID**: Delegation with pagination defaults (limit=10, offset=0)

## Step 5: Infrastructure Layer

**New files** in `internal/order/infrastructure/`:

| File | Contents |
|------|----------|
| `repository/order_postgres.go` | PostgreSQL implementation. `Save` uses `db.BeginTx()` to insert order + items + `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1` atomically. Rollback with `ErrInsufficientStock` if any update affects 0 rows |
| `dto/order_dto.go` | `CreateOrderRequest` (store_id, items[]), `UpdateOrderStatusRequest` (status), `OrderResponse`, `OrderItemResponse` |
| `controller/order_controller.go` | HTTP handlers. `Create` verifies `claims.AccountType == "user"`. `UpdateStatus` passes claims to service. Error mapping: 404/400/403/401/500 |
| `routes/order_routes.go` | Routes wrapped with auth middleware |
| `wiring.go` | `Wire(mux, db, idGen, productRepo, authMw)` → constructs repo → service → controller → routes |

**Routes**:
```
POST   /orders                    → Create (user only)
GET    /orders/{id}               → GetByID
GET    /users/{user_id}/orders    → GetByUserID
GET    /stores/{store_id}/orders  → GetByStoreID
PUT    /orders/{id}/status        → UpdateStatus
```

## Step 6: Database Schema

**File**: `migrations/init.sql` (modify)

```sql
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','confirmed','shipped','delivered','cancelled')),
    total_price DECIMAL(12,2) NOT NULL CHECK (total_price > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price > 0)
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_store_id ON orders(store_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
```

## Step 7: Main Wiring

**File**: `cmd/api/main.go` (modify)

1. Capture product module return value: `productModule := productinfra.Wire(...)`
2. Create JWT generator + auth middleware
3. Wire order module: `orderinfra.Wire(mux, db, uuidGen, productModule.Repository, authMw)`

---

## Key Design Decisions

- **Stock decrement in order repo transaction**: The order repo directly runs `UPDATE products` SQL inside its save transaction for atomicity. This avoids needing a distributed transaction or saga.
- **Price snapshot**: `unit_price` on order items captures the price at order time, decoupled from future price changes.
- **Cancellation stock restore**: Done via product repo (not raw SQL) since cancellation is less performance-critical and benefits from domain validation.
- **Auth middleware scoped to order routes only**: Other modules remain open, consistent with current codebase.

## Files Summary

**New (11 files)**:
- `internal/api/infrastructure/auth_middleware.go`
- `internal/order/domain/order_status.go`
- `internal/order/domain/order_entity.go`
- `internal/order/domain/order_item.go`
- `internal/order/domain/order_repository.go`
- `internal/order/application/order_service.go`
- `internal/order/infrastructure/repository/order_postgres.go`
- `internal/order/infrastructure/dto/order_dto.go`
- `internal/order/infrastructure/controller/order_controller.go`
- `internal/order/infrastructure/routes/order_routes.go`
- `internal/order/infrastructure/wiring.go`

**Modified (3 files)**:
- `internal/api/domain/errors.go`
- `migrations/init.sql`
- `cmd/api/main.go`

## Verification

1. `make db-up` → verify new tables created
2. `make dev` → verify server starts without errors
3. Register a user and a store via auth endpoints, create products
4. Test with curl/Postman:
   - Create order (with valid JWT user token) → verify 201, stock decremented
   - Get order by ID → verify items and prices
   - List orders by user/store → verify pagination
   - Update status (confirm with store token) → verify transition
   - Cancel order (with user token) → verify stock restored
   - Test forbidden transitions → verify 403
   - Test without token → verify 401
