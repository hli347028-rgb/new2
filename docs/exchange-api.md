# AIX → WIN 兑换 & WIN 提现 — 前端接口文档

> 版本：v1 · 更新时间：2026-08-12  
> 基础路径：用户端 `/v1/wallet/*` 管理端 `/api/admin_dhb/*`  
> 数据精度：所有金额字段均为 **decimal(36,18)**，JSON 中以字符串形式返回，前端务必使用 `decimal.js` / `BigNumber` 等库处理，禁止使用浮点直接运算。

---

## 1. 认证机制

所有接口均需携带 JWT Token，支持以下 4 种传递方式（优先级从高到低）：

| 优先级 | 位置 | 格式 |
|--------|------|------|
| 1      | JSON Body | `{ "token": "<jwt>" }` |
| 2      | HTTP Header | `Authorization: Bearer <jwt>` |
| 3      | HTTP Header | `Access-Token: <jwt>` |
| 4      | URL Query | `?token=<jwt>` |

> 用户端和管理端使用不同的 Token（管理端通过 `/api/admin_dhb/login` 获取），管理端 Token 对应的用户必须具有 `admin` 角色。

---

## 2. 兑换接口模块

### 2.1 AIX → WIN 兑换

将 AIX 代币按当前 WIN 价格折算为 WIN 代币，实时扣减 AIX 余额并增加 WIN 余额。

- **Method:** `POST`
- **Path:** `/v1/wallet/exchange-aix-to-win`
- **Auth:** ✅ 需要用户 Token

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 否* | JWT Token（推荐放在 Header 中） |
| aix_amount | string | ✅ | 要兑换的 AIX 数量，必须大于 0 |

**请求示例：**
```json
{
  "token": "eyJhbGciOi...",
  "aix_amount": "100.0000000000000000"
}
```

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| record_id | number(int64) | 兑换记录 ID |
| from_asset | string | 固定 `"AIX"` |
| from_amount | string | 兑换的 AIX 数量 |
| to_asset | string | 固定 `"WIN"` |
| to_amount | string | 折算得到的 WIN 数量 = aix_amount / win_price |
| exchange_price | string | 兑换时使用的 WIN 价格（USDT/枚） |
| status | string | 固定 `"completed"` |
| aix_balance | string | 兑换后 AIX 剩余余额 |
| win_balance | string | 兑换后 WIN 最新余额 |
| created_at | number(int64) | Unix 时间戳（秒） |

**响应示例：**
```json
{
  "record_id": 1001,
  "from_asset": "AIX",
  "from_amount": "100.0000000000000000",
  "to_asset": "WIN",
  "to_amount": "50.0000000000000000",
  "exchange_price": "2.0000000000000000",
  "status": "completed",
  "aix_balance": "900.0000000000000000",
  "win_balance": "50.0000000000000000",
  "created_at": 1755936000
}
```

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 400 | `INVALID_AMOUNT` | 兑换金额 <= 0 |
| 400 | `INSUFFICIENT_AIX` | AIX 余额不足 |
| 400 | `WIN_PRICE_NOT_CONFIGURED` | WIN 价格未配置（后台需先配置） |
| 401 | `UNAUTHORIZED` | Token 无效或过期 |

---

### 2.2 WIN 代币提现

将 WIN 代币提现到指定链上地址，AIX 代币当前**禁止**直接提现。

- **Method:** `POST`
- **Path:** `/v1/wallet/withdraw-win`
- **Auth:** ✅ 需要用户 Token

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 否* | JWT Token（推荐放在 Header 中） |
| amount | string | ✅ | 提现的 WIN 数量，必须大于 0 |
| to_address | string | 否 | 提现目标地址（留空则使用用户绑定地址） |

**请求示例：**
```json
{
  "token": "eyJhbGciOi...",
  "amount": "50.0000000000000000",
  "to_address": "0x1234abcd..."
}
```

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| withdraw_id | number(int64) | 提现记录 ID |
| asset | string | 固定 `"WIN"` |
| amount | string | 提现的 WIN 数量 |
| to_address | string | 提现目标地址 |
| status | string | 初始状态 `"pending"`（链上打款完成后变更为 `"completed"`） |
| tx_hash | string | 交易哈希（链上打款前为空字符串） |
| win_balance | string | 提现后 WIN 剩余余额 |
| win_contract | string | WIN 代币合约地址（当前为空，待配置） |

**响应示例：**
```json
{
  "withdraw_id": 2001,
  "asset": "WIN",
  "amount": "50.0000000000000000",
  "to_address": "0x1234abcd...",
  "status": "pending",
  "tx_hash": "",
  "win_balance": "0.0000000000000000",
  "win_contract": ""
}
```

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 400 | `INVALID_AMOUNT` | 提现金额 <= 0 |
| 400 | `INSUFFICIENT_WIN` | WIN 余额不足 |
| 400 | `INVALID_ADDRESS` | 提现地址格式无效 |
| 401 | `UNAUTHORIZED` | Token 无效或过期 |

---

### 2.3 AIX 提现接口（已禁用）

AIX 代币当前**禁止**直接提现，需先兑换为 WIN 后再通过 `/v1/wallet/withdraw-win` 提现。

- **Method:** `POST`
- **Path:** `/v1/wallet/withdraw-aix`
- **状态：** 调用将直接返回错误

**错误响应：**
```json
{
  "code": 400,
  "reason": "AIX_WITHDRAW_FORBIDDEN",
  "message": "AIX 不可直接提现，请先兑换为 WIN"
}
```

---

## 3. WIN 余额字段

WIN 余额在以下 3 个接口中返回，字段名统一为 `win_balance`，类型为字符串（decimal 精度）。

### 3.1 资产总览接口

- **Method:** `GET`
- **Path:** `/v1/wallet/aix-profile`
- **Auth:** ✅ 需要用户 Token

**响应中与 WIN 相关的字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| win_balance | string | 当前 WIN 代币余额 |
| win_price | **number(float64)** | **⚠️ 注意：此字段为 float64 类型，其余金额字段均为 string(decimal)**。前端展示时请勿直接与 string 类型金额做算术运算，建议先将 float64 转为 decimal 后再计算 |
| aix_balance | string | 当前 AIX 代币余额（兑换前余额） |

**响应示例（节选）：**
```json
{
  "address": "0xUserAddress",
  "aix_balance": "1000.0000000000000000",
  "win_balance": "50.0000000000000000",
  "win_price": 2.0,
  "pending_mgmt_reward": "0.0000000000000000",
  "usdt_recharge": "1000.0000000000000000",
  "usdt_reward": "200.0000000000000000",
  ...
}
```

> **⚠️ 重要类型差异：** `win_price` 为 **`float64`** 类型（如 `2.0`），而 `win_balance`、`aix_balance` 等金额字段均为 **`string`** 类型（decimal 精度）。前端处理时必须：
> 1. 使用 `decimal.js` 的 `Decimal(win_price)` 将 float64 转为 decimal
> 2. 再与 string 类型的余额字段进行乘法运算（如 `Decimal(win_balance).mul(Decimal(win_price))` 估算 WIN 的 USDT 价值）
> 3. **禁止**直接使用 JavaScript 原生 `Number` 混合运算，否则会导致精度丢失

### 3.2 兑换接口返回

见 [2.1 AIX → WIN 兑换](#21-aix--win-兑换)，响应中 `win_balance` 字段为兑换后的最新余额。

### 3.3 WIN 提现接口返回

见 [2.2 WIN 代币提现](#22-win-代币提现)，响应中 `win_balance` 字段为提现后的最新余额。

---

## 4. 兑换历史记录

### 4.1 用户端：查询本人兑换记录

- **Method:** `GET`
- **Path:** `/v1/wallet/exchange-records`
- **Auth:** ✅ 需要用户 Token
- **Query 参数：** 无（自动按当前登录用户筛选）

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| records | array | 兑换记录数组 |

**records 数组元素字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | number(int64) | 记录 ID |
| from_asset | string | 源资产，固定 `"AIX"` |
| from_amount | string | 兑换的 AIX 数量 |
| to_asset | string | 目标资产，固定 `"WIN"` |
| to_amount | string | 折算的 WIN 数量 |
| exchange_price | string | 兑换时的 WIN 价格（USDT/枚） |
| status | string | 状态，固定 `"completed"` |
| remark | string | 备注信息 |
| created_at | number(int64) | Unix 时间戳（秒） |

**响应示例：**
```json
{
  "records": [
    {
      "id": 1001,
      "from_asset": "AIX",
      "from_amount": "100.0000000000000000",
      "to_asset": "WIN",
      "to_amount": "50.0000000000000000",
      "exchange_price": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000",
      "created_at": 1755936000
    },
    {
      "id": 1002,
      "from_asset": "AIX",
      "from_amount": "50.0000000000000000",
      "to_asset": "WIN",
      "to_amount": "25.0000000000000000",
      "exchange_price": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000",
      "created_at": 1756022400
    }
  ]
}
```

---

### 4.2 管理端：查询所有兑换记录

- **Method:** `GET`
- **Path:** `/api/admin_dhb/exchange_list`
- **Auth:** ✅ 需要管理员 Token

**Query 参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页条数，默认 20，最大 1000 |
| address | string | 否 | 按用户地址模糊筛选 |

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 兑换记录数组（分页后） |
| count | number | 总记录数（筛选后） |
| page | number | 当前页码 |

**list 数组元素字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | number(int64) | 记录 ID |
| address | string | 用户钱包地址 |
| fromAsset | string | 源资产 `"AIX"` |
| fromAmount | string | 兑换的 AIX 数量 |
| toAsset | string | 目标资产 `"WIN"` |
| toAmount | string | 折算的 WIN 数量 |
| exchangePrice | string | 兑换时 WIN 价格 |
| status | string | 状态 `"completed"` |
| remark | string | 备注 |
| createdAt | string | 格式化时间 `YYYY-MM-DD HH:mm:ss` |

**请求示例：**
```
GET /api/admin_dhb/exchange_list?page=1&pageSize=20&address=0x1234
Authorization: Bearer <admin_token>
```

**响应示例：**
```json
{
  "list": [
    {
      "id": 1001,
      "address": "0x1234abcd...",
      "fromAsset": "AIX",
      "fromAmount": "100.0000000000000000",
      "toAsset": "WIN",
      "toAmount": "50.0000000000000000",
      "exchangePrice": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000",
      "createdAt": "2026-08-12 10:00:00"
    }
  ],
  "count": 150,
  "page": 1
}
```

---

### 4.3 管理端：查询所有提现记录（含 WIN）

- **Method:** `GET`
- **Path:** `/api/admin_dhb/withdraw_list`
- **Auth:** ✅ 需要管理员 Token

**Query 参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页条数，默认 20，最大 1000 |
| address | string | 否 | 按用户地址模糊筛选 |

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| list / withdraw | array | 提现记录数组（两个字段内容相同，兼容旧前端） |
| count | number | 总记录数 |
| page | number | 当前页码 |

**list 数组元素字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | number(int64) | 提现记录 ID |
| address | string | 用户钱包地址 |
| toAddress | string | 提现目标地址 |
| amount | string | 提现数量 |
| fee | string | 手续费（当前为 0） |
| netAmount | string | 实际到账数量 = amount - fee |
| asset | string | 资产类型 `"WIN"`（AIX 提现已禁用） |
| status | string | `"pending"` / `"completed"` / `"failed"` |
| txHash | string | 链上交易哈希（pending 时为空） |
| createdAt | string | 格式化时间 `YYYY-MM-DD HH:mm:ss` |

**响应示例：**
```json
{
  "withdraw": [
    {
      "id": 2001,
      "address": "0x1234abcd...",
      "toAddress": "0x5678efgh...",
      "amount": "50.0000000000000000",
      "fee": "0.0000000000000000",
      "netAmount": "50.0000000000000000",
      "asset": "WIN",
      "status": "pending",
      "txHash": "",
      "createdAt": "2026-08-12 10:30:00"
    }
  ],
  "list": [ /* 同上，兼容字段 */ ],
  "count": 25,
  "page": 1
}
```

---

## 5. 数据字典

### ExchangeRecord（兑换记录）

| 字段 | 类型 | 数据库列 | 说明 |
|------|------|----------|------|
| ID | int64 | id | 主键自增 |
| UserID | int64 | user_id | 用户 ID（索引） |
| FromAsset | string(16) | from_asset | 源资产，固定 `"AIX"` |
| FromAmount | decimal(36,18) | from_amount | 兑换的 AIX 数量 |
| ToAsset | string(16) | to_asset | 目标资产，固定 `"WIN"` |
| ToAmount | decimal(36,18) | to_amount | 折算的 WIN 数量 = FromAmount / ExchangePrice |
| ExchangePrice | decimal(36,18) | exchange_price | 兑换时的 WIN 价格（USDT/枚） |
| Status | string(16) | status | 状态，默认 `"completed"` |
| Remark | string(255) | remark | 备注信息 |
| CreatedTime | time | created_time | 创建时间（自动生成） |

### Withdrawal（提现记录）

| 字段 | 类型 | 数据库列 | 说明 |
|------|------|----------|------|
| ID | int64 | id | 主键自增 |
| UserID | int64 | user_id | 用户 ID（索引） |
| Asset | string(16) | asset | 资产类型 `"WIN"`（AIX 已禁用） |
| Amount | decimal(36,18) | amount | 提现数量 |
| Fee | decimal(36,18) | fee | 手续费，默认 0 |
| PayAmount | decimal(36,18) | pay_amount | 实际到账 = Amount - Fee |
| ToAddress | string(42) | to_address | 提现目标地址 |
| TxHash | string(66) | tx_hash | 链上交易哈希（pending 时为空） |
| Status | string(16) | status | `"pending"` / `"completed"` / `"failed"` |
| Remark | string(255) | remark | 备注 |
| CreatedTime | time | created_time | 创建时间 |
| UpdatedTime | time | updated_time | 更新时间 |

### User WIN 余额

| 字段 | 类型 | 数据库列 | 说明 |
|------|------|----------|------|
| WinBalance | decimal(36,18) | win_balance | WIN 代币余额，默认 0 |
| AixBalance | decimal(36,18) | aix_balance | AIX 代币余额，默认 0 |

---

## 6. 兑换 & 提现流程图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        AIX → WIN 兑换流程                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  用户前端                      后端 API                    数据库   │
│    │                             │                          │       │
│    │  POST /v1/wallet/           │                          │       │
│    │  exchange-aix-to-win        │                          │       │
│    │  { aix_amount: "100" }      │                          │       │
│    │ ──────────────────────────> │                          │       │
│    │                             │  1. 校验 Token            │       │
│    │                             │  2. 校验金额 > 0         │       │
│    │                             │  3. 读取 WIN 价格        │       │
│    │                             │  4. 计算 WIN 数量        │       │
│    │                             │     win = aix / price    │       │
│    │                             │                          │       │
│    │                             │  5. 开启事务             │       │
│    │                             │     ┌────────────────┐   │       │
│    │                             │     │ UPDATE users   │   │       │
│    │                             │     │ SET aix_bal-100 │   │       │
│    │                             │     │ SET win_bal+50  │   │       │
│    │                             │     │ INSERT         │   │       │
│    │                             │     │ exchange_records│   │       │
│    │                             │     └────────────────┘   │       │
│    │                             │  6. 提交事务             │       │
│    │                             │                          │       │
│    │  {                         │                          │       │
│    │    win_balance: "50",      │                          │       │
│    │    aix_balance: "900",     │                          │       │
│    │    to_amount: "50"         │                          │       │
│    │  }                         │                          │       │
│    │ <────────────────────────── │                          │       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                        WIN 提现流程                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  用户前端                      后端 API                    数据库   │
│    │                             │                          │       │
│    │  POST /v1/wallet/           │                          │       │
│    │  withdraw-win               │                          │       │
│    │  { amount: "50" }           │                          │       │
│    │ ──────────────────────────> │                          │       │
│    │                             │  1. 校验 Token            │       │
│    │                             │  2. 校验金额 > 0         │       │
│    │                             │  3. 校验地址（可选）      │       │
│    │                             │                          │       │
│    │                             │  4. 开启事务             │       │
│    │                             │     ┌────────────────┐   │       │
│    │                             │     │ UPDATE users   │   │       │
│    │                             │     │ SET win_bal-50  │   │       │
│    │                             │     │ INSERT         │   │       │
│    │                             │     │ withdrawals    │   │       │
│    │                             │     │ (pending)      │   │       │
│    │                             │     └────────────────┘   │       │
│    │                             │  5. 提交事务             │       │
│    │                             │                          │       │
│    │  {                         │                          │       │
│    │    status: "pending",      │                          │       │
│    │    win_balance: "0",       │                          │       │
│    │    tx_hash: ""             │                          │       │
│    │  }                         │                          │       │
│    │ <────────────────────────── │                          │       │
│    │                             │                          │       │
│    │     ┌─────────────────────────────────────────┐              │
│    │     │  后台链上服务（TODO: 待实现）              │              │
│    │     │  监听 pending 提现 → 合约转账 → 更新 tx_hash │              │
│    │     └─────────────────────────────────────────┘              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 7. 前端对接注意事项

1. **精度处理：** 所有金额字段为字符串类型，前端必须使用 `decimal.js` 或 `BigNumber` 进行运算，禁止使用 JavaScript 原生 `Number` 类型。

2. **WIN 价格展示：** 调用 `/v1/wallet/aix-profile` 获取 `win_price`，用于兑换页面展示实时比率：`1 AIX = (1 / win_price) WIN`。

3. **兑换校验：** 兑换前建议先查询 `aix_balance`，避免提交后因余额不足报错。

4. **AIX 提现已禁用：** 前端应隐藏 AIX 提现入口，引导用户先兑换为 WIN 再提现。

5. **提现状态轮询：** 提现提交后状态为 `pending`，需前端定期轮询或等待 WebSocket 推送，状态变更为 `completed` 时表示链上打款成功。

6. **管理后台筛选：** 管理端列表接口支持 `address` 模糊搜索，前端应将搜索词原样传递。

7. **分页参数差异：** 
   - 用户端：`page` + `page_size`
   - 管理端：`page` + `pageSize`

8. **WIN 价格配置：** WIN 价格由管理员通过系统配置接口设置，不属于 Legacy 配置项。需通过 Proto 管理端 API `/admin.v1.AdminService/UpdateConfig` 传入 `win_price` 字段（float64），或在数据库 `settings` 表中直接修改 `SystemConfigSnapshot` 的 `WinPrice` 字段。

---

## 附录：WIN 价格配置

WIN 价格是 AIX → WIN 兑换比率的核心参数，必须由管理员预先配置。

### 配置方式

**Proto 管理端 API：**
- **Method:** `POST`
- **Path:** `/admin.v1.AdminService/UpdateConfig`
- **Body:** 包含 `win_price` 字段（float64，USDT/枚）

**示例：**
```json
{
  "token": "<admin_token>",
  "win_price": 2.0
}
```

**Legacy 配置接口（暂未暴露 WIN 价格项）：**

Legacy 管理后台配置接口 `/api/admin_dhb/config` 的配置项列表中**暂无** WIN 价格项。如需通过 Legacy 后台配置，需在配置项列表中新增一项（建议 ID 31，名称 "WIN价格(USDT/枚)"），并在 `applyLegacyConfigUpdate` 中补充对应处理逻辑。

### 价格生效范围

- WIN 价格存储于 `SystemConfigSnapshot.WinPrice` 字段
- 兑换时调用 `biz.GetWinPrice()` 获取当前生效价格
- 价格修改后实时生效，无需重启服务
