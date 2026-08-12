# AIX → WIN 兑换 & WIN 提现 — 前端接口文档

> 版本：v2 · 更新时间：2026-08-12  
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

将 AIX 代币按当前 WIN 价格折算为 WIN 代币，扣除手续费后实时扣减 AIX 余额并增加 WIN 余额。

#### 兑换计算公式

```
WIN 毛量 = AIX 数量 ÷ WIN 价格
手续费   = WIN 毛量 × 手续费率（默认 5%）
WIN 净量 = WIN 毛量 − 手续费（用户实际到账）
```

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
| to_amount | string | **扣除手续费后**实际到账的 WIN 数量 |
| exchange_price | string | 兑换时使用的 WIN 价格（USDT/枚） |
| exchange_fee_rate | number(float64) | 当前手续费率（如 `0.05` = 5%） |
| status | string | 固定 `"completed"` |
| aix_balance | string | 兑换后 AIX 剩余余额 |
| win_balance | string | 兑换后 WIN 最新余额 |
| created_at | number(int64) | Unix 时间戳（秒） |

> **⚠️ 类型提示：** `exchange_fee_rate` 为 `float64` 类型（如 `0.05`），前端展示时建议乘以 100 转为百分比字符串（如 `5%`）。其余金额字段均为 `string` 类型。

**响应示例：**
```json
{
  "record_id": 1001,
  "from_asset": "AIX",
  "from_amount": "100.0000000000000000",
  "to_asset": "WIN",
  "to_amount": "47.5000000000000000",
  "exchange_price": "2.0000000000000000",
  "exchange_fee_rate": 0.05,
  "status": "completed",
  "aix_balance": "900.0000000000000000",
  "win_balance": "47.5000000000000000",
  "created_at": 1755936000
}
```

> **计算过程示例：** 100 AIX ÷ 2.0 USDT/枚 = 50 WIN（毛量），扣除 5% 手续费 = 2.5 WIN，实际到账 47.5 WIN。

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 400 | `INVALID_AMOUNT` | 兑换金额 <= 0 |
| 400 | `INSUFFICIENT_AIX` | AIX 余额不足 |
| 400 | `WIN_PRICE_NOT_CONFIGURED` | WIN 价格未配置 |
| 400 | `WIN_NET_AMOUNT_TOO_SMALL` | 扣除手续费后 WIN 净量 <= 0（兑换金额太小） |
| 401 | `UNAUTHORIZED` | Token 无效或过期 |

---

### 2.2 查询本人兑换记录

查询当前用户的 AIX→WIN 兑换历史记录，包含手续费明细。

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
| to_amount | string | **扣除手续费后**实际到账的 WIN 数量 |
| fee_amount | string | 扣除的手续费（WIN 数量） |
| fee_rate | string | 兑换时的手续费率（decimal，如 `0.050000` 表示 5%） |
| exchange_price | string | 兑换时的 WIN 价格（USDT/枚） |
| status | string | 状态，固定 `"completed"` |
| remark | string | 备注信息（含手续费率说明） |
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
      "to_amount": "47.5000000000000000",
      "fee_amount": "2.5000000000000000",
      "fee_rate": "0.050000",
      "exchange_price": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000, fee 5.00%",
      "created_at": 1755936000
    },
    {
      "id": 1002,
      "from_asset": "AIX",
      "from_amount": "50.0000000000000000",
      "to_asset": "WIN",
      "to_amount": "23.7500000000000000",
      "fee_amount": "1.2500000000000000",
      "fee_rate": "0.050000",
      "exchange_price": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000, fee 5.00%",
      "created_at": 1756022400
    }
  ]
}
```

---

### 2.3 管理端：查询所有兑换记录

管理员查看全平台 AIX→WIN 兑换记录，支持按地址筛选。

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
| toAmount | string | 扣除手续费后到账的 WIN 数量 |
| feeAmount | string | 扣除的手续费（WIN 数量） |
| feeRate | string | 手续费率（decimal，如 `0.050000`） |
| exchangePrice | string | 兑换时 WIN 价格 |
| status | string | 状态 `"completed"` |
| remark | string | 备注（含手续费率说明） |
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
      "toAmount": "47.5000000000000000",
      "feeAmount": "2.5000000000000000",
      "feeRate": "0.050000",
      "exchangePrice": "2.0000000000000000",
      "status": "completed",
      "remark": "AIX→WIN exchange at price 2.0000000000000000, fee 5.00%",
      "createdAt": "2026-08-12 10:00:00"
    }
  ],
  "count": 150,
  "page": 1
}
```

---

## 3. WIN 提现接口模块

### 3.1 WIN 代币提现

将 WIN 代币提现到指定链上地址。AIX 代币当前**禁止**直接提现，需先兑换为 WIN。

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
| to_address | string | 提现目标地址（校验并规范化后的地址） |
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
| 400 | `INVALID_ADDRESS` | 提现地址格式无效（非合法以太坊地址） |
| 401 | `UNAUTHORIZED` | Token 无效或过期 |

---

### 3.2 查询本人提现记录

查询当前用户的 WIN 提现历史记录。

- **Method:** `GET`
- **Path:** `/v1/wallet/withdraw-records`
- **Auth:** ✅ 需要用户 Token
- **Query 参数：** 无

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| records | array | 提现记录数组 |

**records 数组元素字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | number(int64) | 提现记录 ID |
| asset | string | 资产类型，固定 `"WIN"` |
| amount | string | 提现数量 |
| fee | string | 手续费（当前为 `0`，预留给链上Gas费） |
| net_amount | string | 实际到账数量 = amount - fee |
| to_address | string | 提现目标地址 |
| status | string | `"pending"` / `"completed"` / `"failed"` |
| tx_hash | string | 链上交易哈希（pending 时为空） |
| remark | string | 备注 |
| created_at | number(int64) | Unix 时间戳（秒） |
| updated_at | number(int64) | 更新时间戳（秒） |

**响应示例：**
```json
{
  "records": [
    {
      "id": 2001,
      "asset": "WIN",
      "amount": "50.0000000000000000",
      "fee": "0.0000000000000000",
      "net_amount": "50.0000000000000000",
      "to_address": "0x1234abcd...",
      "status": "pending",
      "tx_hash": "",
      "remark": "",
      "created_at": 1755936000,
      "updated_at": 1755936000
    },
    {
      "id": 2002,
      "asset": "WIN",
      "amount": "30.0000000000000000",
      "fee": "0.0000000000000000",
      "net_amount": "30.0000000000000000",
      "to_address": "0x5678efgh...",
      "status": "completed",
      "tx_hash": "0xabcdef1234567890...",
      "remark": "",
      "created_at": 1755849600,
      "updated_at": 1755849600
    }
  ]
}
```

---

### 3.3 管理端：查询所有提现记录

管理员查看全平台提现记录（含 WIN），支持按地址筛选。

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
| fee | string | 手续费（当前为 `0`） |
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

### 3.4 AIX 提现接口（已禁用）

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

## 4. 资产总览 & WIN 余额

### 4.1 资产总览接口

- **Method:** `GET`
- **Path:** `/v1/wallet/aix-profile`
- **Auth:** ✅ 需要用户 Token

**响应中与 WIN 相关的字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| win_balance | string | 当前 WIN 代币余额 |
| win_price | **number(float64)** | **⚠️ 注意：此字段为 float64 类型**。前端展示时请勿直接与 string 类型金额做算术运算 |
| exchange_fee_rate | **number(float64)** | 当前兑换手续费率（如 `0.05` = 5%） |
| aix_balance | string | 当前 AIX 代币余额 |

**响应示例（节选）：**
```json
{
  "address": "0xUserAddress",
  "aix_balance": "1000.0000000000000000",
  "win_balance": "47.5000000000000000",
  "win_price": 2.0,
  "exchange_fee_rate": 0.05,
  "pending_mgmt_reward": "0.0000000000000000",
  "usdt_recharge": "1000.0000000000000000",
  "usdt_reward": "200.0000000000000000",
  ...
}
```

> **⚠️ 重要类型差异：** `win_price` 和 `exchange_fee_rate` 为 **`float64`** 类型（如 `2.0`、`0.05`），而 `win_balance`、`aix_balance` 等金额字段均为 **`string`** 类型（decimal 精度）。前端处理时必须：
> 1. 使用 `decimal.js` 的 `Decimal(win_price)` 将 float64 转为 decimal
> 2. 再与 string 类型的余额字段进行乘法运算
> 3. **禁止**直接使用 JavaScript 原生 `Number` 混合运算

### 4.2 WIN 余额变动时机

| 场景 | 接口 | win_balance 变化 |
|------|------|------------------|
| AIX→WIN 兑换 | `POST /v1/wallet/exchange-aix-to-win` | +WIN 净量（已扣手续费） |
| WIN 提现 | `POST /v1/wallet/withdraw-win` | −提现数量 |
| 查看资产 | `GET /v1/wallet/aix-profile` | 返回当前值 |

---

## 5. 管理后台配置项

### 5.1 配置项列表

管理员通过 Legacy 配置接口管理系统参数，包括 **WIN 价格** 和 **兑换手续费率**。

- **Method:** `GET`
- **Path:** `/api/admin_dhb/config`
- **Auth:** ✅ 需要管理员 Token

**响应示例：**
```json
{
  "config": [
    { "id": 1, "name": "静态利率", "value": "0.5" },
    { "id": 2, "name": "出局倍数", "value": "4" },
    { "id": 3, "name": "直推比例", "value": "0.5" },
    { "id": 7, "name": "AIX价格(USDT/枚)", "value": "1.0" },
    { "id": 8, "name": "WIN价格(USDT/枚)", "value": "2.0" },
    { "id": 9, "name": "兑换手续费率(%)", "value": "5" },
    ...
  ]
}
```

### 5.2 修改配置项

- **Method:** `POST`
- **Path:** `/api/admin_dhb/config_update`
- **Content-Type:** `application/x-www-form-urlencoded`
- **Auth:** ✅ 需要管理员 Token

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | ✅ | 配置项 ID |
| value | string | ✅ | 配置项新值 |

**重要配置项说明：**

| ID | 名称 | 格式 | 示例 | 说明 |
|----|------|------|------|------|
| 7 | AIX价格(USDT/枚) | float | `1.0` | AIX 代币单价 |
| 8 | WIN价格(USDT/枚) | float | `2.0` | WIN 代币单价（兑换比率） |
| 9 | 兑换手续费率(%) | float | `5` | 百分比输入，5 = 5%，存储为 0.05 |

**请求示例（修改 WIN 价格为 2.5 USDT/枚）：**
```
POST /api/admin_dhb/config_update
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer <admin_token>

id=8&value=2.5
```

**请求示例（修改手续费率为 3%）：**
```
POST /api/admin_dhb/config_update
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer <admin_token>

id=9&value=3
```

---

## 6. 数据字典

### ExchangeRecord（兑换记录）

| 字段 | 类型 | 数据库列 | 说明 |
|------|------|----------|------|
| ID | int64 | id | 主键自增 |
| UserID | int64 | user_id | 用户 ID（索引） |
| FromAsset | string(16) | from_asset | 源资产，固定 `"AIX"` |
| FromAmount | decimal(36,18) | from_amount | 兑换的 AIX 数量 |
| ToAsset | string(16) | to_asset | 目标资产，固定 `"WIN"` |
| ToAmount | decimal(36,18) | to_amount | 扣除手续费后实际到账的 WIN 数量 |
| FeeAmount | decimal(36,18) | fee_amount | 扣除的手续费（WIN 数量），默认 0 |
| ExchangePrice | decimal(36,18) | exchange_price | 兑换时的 WIN 价格（USDT/枚） |
| FeeRate | decimal(12,6) | fee_rate | 兑换时的手续费率（如 0.050000），默认 0 |
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

### User Wallet（用户钱包字段）

| 字段 | 类型 | 数据库列 | 说明 |
|------|------|----------|------|
| WinBalance | decimal(36,18) | win_balance | WIN 代币余额，默认 0 |
| AixBalance | decimal(36,18) | aix_balance | AIX 代币余额，默认 0 |
| PendingMgmtReward | decimal(36,18) | pending_mgmt_reward | 待释放管理奖，默认 0 |

### SystemConfigSnapshot（系统配置）

| 字段 | 类型 | JSON | 说明 |
|------|------|------|------|
| WinPrice | float64 | `win_price` | WIN 代币价格（USDT/枚），默认 1.0 |
| ExchangeFeeRate | float64 | `exchange_fee_rate` | 兑换手续费率，默认 0.05 (5%) |

---

## 7. 兑换 & 提现流程图

### 7.1 AIX → WIN 兑换流程（含手续费）

```
┌──────────────────────────────────────────────────────────────────────┐
│                     AIX → WIN 兑换流程（含手续费）                      │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  用户前端                        后端 API                      数据库  │
│    │                               │                            │     │
│    │  POST /v1/wallet/             │                            │     │
│    │  exchange-aix-to-win          │                            │     │
│    │  { aix_amount: "100" }        │                            │     │
│    │ ────────────────────────────> │                            │     │
│    │                               │  1. 校验 Token              │     │
│    │                               │  2. 校验金额 > 0           │     │
│    │                               │  3. 读取 WIN 价格 (2.0)     │     │
│    │                               │  4. 读取手续费率 (5%)       │     │
│    │                               │                            │     │
│    │                               │  5. 计算:                   │     │
│    │                               │     winGross = 100 / 2.0   │     │
│    │                               │              = 50 WIN     │     │
│    │                               │     fee      = 50 × 0.05   │     │
│    │                               │              = 2.5 WIN    │     │
│    │                               │     winNet   = 50 − 2.5    │     │
│    │                               │              = 47.5 WIN   │     │
│    │                               │                            │     │
│    │                               │  6. 开启事务               │     │
│    │                               │     ┌──────────────────┐   │     │
│    │                               │     │ UPDATE users     │   │     │
│    │                               │     │ SET aix_bal-100  │   │     │
│    │                               │     │ SET win_bal+47.5 │   │     │
│    │                               │     │ INSERT           │   │     │
│    │                               │     │ exchange_records │   │     │
│    │                               │     │ (fee=2.5,rate=5%)│   │     │
│    │                               │     └──────────────────┘   │     │
│    │                               │  7. 提交事务               │     │
│    │                               │                            │     │
│    │  {                           │                            │     │
│    │    to_amount: "47.5",        │                            │     │
│    │    exchange_fee_rate: 0.05,  │                            │     │
│    │    win_balance: "47.5"       │                            │     │
│    │  }                           │                            │     │
│    │ <─────────────────────────── │                            │     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 7.2 WIN 提现流程

```
┌──────────────────────────────────────────────────────────────────────┐
│                          WIN 提现流程                                  │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  用户前端                        后端 API                      数据库  │
│    │                               │                            │     │
│    │  POST /v1/wallet/             │                            │     │
│    │  withdraw-win                 │                            │     │
│    │  { amount: "50" }             │                            │     │
│    │ ────────────────────────────> │                            │     │
│    │                               │  1. 校验 Token              │     │
│    │                               │  2. 校验金额 > 0           │     │
│    │                               │  3. 校验/规范化目标地址     │     │
│    │                               │                            │     │
│    │                               │  4. 开启事务               │     │
│    │                               │     ┌──────────────────┐   │     │
│    │                               │     │ UPDATE users     │   │     │
│    │                               │     │ SET win_bal-50   │   │     │
│    │                               │     │ INSERT           │   │     │
│    │                               │     │ withdrawals      │   │     │
│    │                               │     │ (pending)        │   │     │
│    │                               │     └──────────────────┘   │     │
│    │                               │  5. 提交事务               │     │
│    │                               │                            │     │
│    │  {                           │                            │     │
│    │    status: "pending",        │                            │     │
│    │    win_balance: "0",         │                            │     │
│    │    tx_hash: ""               │                            │     │
│    │  }                           │                            │     │
│    │ <─────────────────────────── │                            │     │
│    │                               │                            │     │
│    │     ┌─────────────────────────────────────────┐                │
│    │     │  后台链上服务（TODO: 待实现）              │                │
│    │     │  监听 pending 提现 → 合约转账 → 更新 tx_hash │                │
│    │     └─────────────────────────────────────────┘                │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 8. 前端对接注意事项

### 8.1 精度处理

所有金额字段为字符串类型，前端必须使用 `decimal.js` 或 `BigNumber` 进行运算，禁止使用 JavaScript 原生 `Number` 类型。

### 8.2 兑换页面对接要点

1. **获取实时费率：** 调用 `GET /v1/wallet/aix-profile` 获取：
   - `win_price`（float64）：当前 WIN 价格
   - `exchange_fee_rate`（float64）：当前手续费率
   - `aix_balance`（string）：AIX 可用余额
   
2. **前端预估展示：** 兑换输入框旁实时展示预估结果：
   ```
   输入 100 AIX → 预估获得 47.5 WIN（扣 2.5 WIN 手续费）
   ```
   ```javascript
   const winGross = Decimal(aixAmount).div(Decimal(winPrice))
   const fee = winGross.mul(Decimal(exchangeFeeRate))
   const winNet = winGross.sub(fee)
   ```

3. **兑换校验：** 兑换前先查询 `aix_balance`，避免提交后因余额不足报错。

4. **提交兑换：** `POST /v1/wallet/exchange-aix-to-win`，成功响应中 `to_amount` 为实际到账数量。

### 8.3 提现页面对接要点

1. **前置条件：** 用户需先将 AIX 兑换为 WIN，才能提现。
2. **地址校验：** 前端提交前可先用 ethers.js 校验地址格式（`ethers.isAddress()`）。
3. **状态轮询：** 提现提交后状态为 `pending`，需前端定期轮询或等待 WebSocket 推送，状态变更为 `completed` 时表示链上打款成功。
4. **历史展示：** 调用 `GET /v1/wallet/withdraw-records` 展示提现历史。

### 8.4 管理后台对接要点

1. **配置项管理：** 管理后台配置页面 `GET /api/admin_dhb/config` 动态渲染所有配置项，包括：
   - ID=7: AIX价格(USDT/枚)
   - ID=8: WIN价格(USDT/枚)
   - ID=9: 兑换手续费率(%) — 输入百分数（如 `5` 表示 5%），后端自动存储为 `0.05`
   
2. **兑换记录：** `GET /api/admin_dhb/exchange_list` 展示所有兑换记录，含手续费明细。

3. **提现记录：** `GET /api/admin_dhb/withdraw_list` 展示全平台提现记录。

4. **用户管理：** 用户列表新增字段：
   - `win_balance`: WIN 代币数
   - `pending_mgmt_reward`: 待释放管理奖

### 8.5 类型差异汇总

| 字段 | 类型 | 说明 |
|------|------|------|
| `win_price` | float64 | WIN 价格（前端 `Number` 类型） |
| `exchange_fee_rate` | float64 | 手续费率（前端 `Number` 类型，如 `0.05`） |
| `win_balance` | string(decimal) | WIN 余额（前端 `String`，必须转 decimal 运算） |
| `aix_balance` | string(decimal) | AIX 余额 |
| `to_amount` | string(decimal) | 兑换到账 WIN 数量（已扣手续费） |
| `fee_amount` | string(decimal) | 扣除手续费数量 |

### 8.6 分页参数差异

| 端 | 页码参数 | 每页条数参数 |
|----|----------|-------------|
| 用户端 | `page` | `page_size` |
| 管理端 | `page` | `pageSize` |

---

## 附录：配置接口汇总

| 配置项 ID | 名称 | 存储格式 | 输入格式 | 说明 |
|-----------|------|----------|----------|------|
| 7 | AIX价格(USDT/枚) | float64 | float | 默认 1.0 |
| 8 | WIN价格(USDT/枚) | float64 | float | 默认 1.0，兑换比率 |
| 9 | 兑换手续费率(%) | float64 | float 百分数 | 默认 5（即 5%），存储为 0.05 |

所有配置项修改后**实时生效**，无需重启服务。