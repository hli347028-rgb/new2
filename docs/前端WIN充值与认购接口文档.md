# WIN 充值 & WIN 认购 — 前端接口文档

> 版本：v1 · 更新时间：2026-08-13  
> 基础路径：用户端 `/v1/wallet/*` · 管理端 `/api/admin_dhb/*`  
> 数据精度：金额字段均为 **decimal(36,18)** 字符串，前端请使用 `decimal.js` / `BigNumber`，禁止浮点直接运算。

---

## 0. 业务说明

| 能力 | 说明 |
|------|------|
| WIN 充值 | 用户链上向平台收款地址转入 WIN，确认后入账 `users.win_balance` |
| WIN 认购 | `pay_from=win`，认购金额仍为 **USDT 面额**；按当前 `win_price` 折算扣减 WIN：`WIN = USDT ÷ win_price` |
| 直推 | WIN 认购 **等同充值钱包报单**，`direct_base = USDT 本金`，产生直推奖 |
| 出局/管理奖 | 订单本金、出局倍数仍按 USDT 记账，与 USDT 认购一致 |

**折算示例：** 认购 `100` USDT，`win_price = 2` → 扣 `50` WIN，订单本金记 `100` USDT。

---

## 1. 认证机制

与现有钱包接口一致，JWT 优先级：

| 优先级 | 位置 | 格式 |
|--------|------|------|
| 1 | JSON Body | `{ "token": "<jwt>" }` |
| 2 | Header | `Authorization: Bearer <jwt>` |
| 3 | Header | `Access-Token: <jwt>` |
| 4 | Query | `?token=<jwt>` |

---

## 2. 资产总览（含 WIN 价格 / 合约）

### 2.1 获取 AIX/WIN 资产画像

- **Method:** `GET`
- **Path:** `/v1/wallet/aix-profile`
- **Auth:** ✅

**成功响应关键字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| usdt_recharge | string | 充值钱包 USDT |
| usdt_reward | string | 奖励钱包 USDT |
| win_balance | string | WIN 余额 |
| win_price | number | WIN 现价（USDT/枚） |
| win_contract | string | WIN ERC20 合约（未配置为空） |
| exchange_fee_rate | number | AIX→WIN 兑换手续费率 |

前端认购页用 `win_balance` + `win_price` 估算：`need_win = amount_usdt / win_price`。

---

## 3. WIN 充值模块

### 3.1 创建 WIN 充值单

- **Method:** `POST`
- **Path:** `/v1/wallet/recharge-win`
- **Auth:** ✅

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | string | ✅ | WIN 数量，≥ 1 |
| token | string | 否* | JWT（推荐 Header） |

**请求示例：**

```json
{
  "amount": "100"
}
```

**成功响应 (200)：**

| 字段 | 类型 | 说明 |
|------|------|------|
| recharge_id | number | 充值单 ID |
| asset | string | 固定 `"WIN"` |
| amount | string | 充值数量 |
| deposit_address | string | 主收款地址 |
| deposit_addresses | string[] | 全部收款地址 |
| win_contract | string | WIN 合约地址 |
| win_decimals | number | WIN 精度（默认 18） |
| token_symbol | string | `"WIN"` |
| message | string | 待签名原文（确认时用） |
| expire_at | number | 过期 Unix 秒 |
| dev_mode | boolean | 开发模式（无 RPC 时为 true） |
| win_price | number | 当前 WIN 价格（展示用） |

**响应示例：**

```json
{
  "recharge_id": 2001,
  "asset": "WIN",
  "amount": "100",
  "deposit_address": "0x...",
  "deposit_addresses": ["0x..."],
  "win_contract": "0x...",
  "win_decimals": 18,
  "token_symbol": "WIN",
  "message": "Recharge WIN to AIX account\n...",
  "expire_at": 1755057600,
  "dev_mode": false,
  "win_price": 2
}
```

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 400 | `INVALID_AMOUNT` | 数量 < 1 |
| 400 | `DEPOSIT_NOT_CONFIGURED` | 收款地址未配置 |
| 400 | `WIN_NOT_CONFIGURED` | `win_contract` 未配置（非开发模式） |

**前端流程：**

1. 调用本接口拿到 `deposit_address` / `win_contract` / `message`
2. 用户钱包：`WIN.approve`（如需）→ `WIN.transfer(deposit_address, amount)`
3. 对 `message` 做 personal_sign
4. 调用确认接口（见 3.2）

---

### 3.2 确认 WIN 充值

- **Method:** `POST`
- **Path:** `/v1/wallet/recharge-win/confirm`
- **Auth:** ✅

校验链上 ERC20 `Transfer` 日志（from=用户，to=平台收款地址，合约=`win_contract`，数量≥下单量）后，**直接入账** `win_balance`。

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| recharge_id | number | ✅ | 创建接口返回的 ID |
| tx_hash | string | ✅ | 链上交易哈希 |
| signature | string | ✅ | 对创建接口 `message` 的个人签名 |
| token | string | 否* | JWT |

**请求示例：**

```json
{
  "recharge_id": 2001,
  "tx_hash": "0xabc...",
  "signature": "0xdef..."
}
```

**成功响应 (200)：**

```json
{
  "asset": "WIN",
  "amount": "100",
  "win_balance": "150.0000000000000000"
}
```

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 404 | `RECHARGE_NOT_FOUND` | 充值单不存在 |
| 403 | `RECHARGE_FORBIDDEN` | 非本人订单 |
| 400 | `INVALID_ASSET` | 非 WIN 充值单 |
| 400 | `RECHARGE_CONFIRMED` | 已确认 |
| 400 | `RECHARGE_EXPIRED` | 已过期 |
| 400 | `INVALID_TX_HASH` | 哈希为空 |
| 400 | `TX_HASH_USED` | 哈希已被占用 |
| 401 | `INVALID_SIGNATURE` | 签名失败 |
| 400 | `TX_VERIFY_FAILED` | 链上校验失败 |

---

### 3.3 WIN 充值记录

- **Method:** `GET`
- **Path:** `/v1/wallet/recharges-win`
- **Auth:** ✅

**成功响应：**

```json
{
  "recharges": [
    {
      "id": 2001,
      "asset": "WIN",
      "amount": "100",
      "tx_hash": "0xabc...",
      "status": "confirmed",
      "created_at": 1755054000,
      "confirmed_at": 1755054100
    }
  ]
}
```

`status`：`pending` | `confirmed` | `rejected`

---

## 4. WIN 替代 USDT 认购

### 4.1 认购 / 复投 / WIN 支付

- **Method:** `POST`
- **Path:** `/v1/wallet/subscribe-aix`
- **Auth:** ✅

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | string | ✅ | **USDT 面额**本金，≥ `min_subscribe`（默认 100） |
| pay_from | string | ✅ | `recharge` \| `reward` \| **`win`** |
| token | string | 否* | JWT |

**`pay_from` 对照：**

| 值 | 扣款账户 | 是否直推 | 说明 |
|----|----------|----------|------|
| `recharge` | `usdt_recharge` | ✅ | 充值钱包 USDT |
| `reward` | `usdt_reward` | ❌ | 奖励钱包复投 |
| `win` | `win_balance` | ✅ | 按 `win_price` 折算扣 WIN |

**请求示例（WIN 认购）：**

```json
{
  "amount": "100",
  "pay_from": "win"
}
```

**成功响应 (200) — pay_from=win：**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | number | 订单 ID |
| principal | string | USDT 本金 |
| exit_cap | string | 出局上限（默认 4×） |
| fund_source | string | `"win"` |
| direct_base | string | 直推基数（= principal） |
| status | string | `"active"` |
| balance | string | 扣款后 WIN 余额 |
| total_amount | string | 同 principal |
| from_win | string | 实际扣减的 WIN 数量 |
| win_price | string | 下单时 WIN 价格快照 |
| win_balance | string | 同 balance |

**响应示例：**

```json
{
  "order_id": 501,
  "principal": "100",
  "exit_cap": "400",
  "fund_source": "win",
  "direct_base": "100",
  "status": "active",
  "balance": "50.00000000",
  "total_amount": "100",
  "from_win": "50.00000000",
  "win_price": "2",
  "win_balance": "50.00000000"
}
```

**错误码：**

| HTTP | Code | 说明 |
|------|------|------|
| 400 | `INVALID_PAY_FROM` | 非 recharge/reward/win |
| 400 | `INVALID_AMOUNT` | 金额 ≤ 0 |
| 400 | `MIN_SUBSCRIBE_LIMIT` | 低于最低认购额 |
| 400 | `WIN_PRICE_NOT_CONFIGURED` | WIN 价格未配置 |
| 400 | `INSUFFICIENT_WIN` | WIN 余额不足 |
| 400 | `INSUFFICIENT_BALANCE` | USDT 钱包余额不足 |

### 4.2 订单列表（资金来源展示）

- **Method:** `GET`
- **Path:** `/v1/wallet/orders`

订单 `product_name` / `fund_source` 可能为：`recharge` | `reward` | `win`。  
前端文案建议：`recharge`→充值，`reward`→奖励，`win`→WIN。

---

## 5. 管理端：人工充值 WIN

- **Method:** `POST`
- **Path:** `/api/admin_dhb/admin_recharge_win`
- **Content-Type:** `application/x-www-form-urlencoded`
- **Auth:** ✅ 管理端 Token

**表单字段：**

| 字段 | 说明 |
|------|------|
| address | 用户钱包地址 |
| amount | WIN 数量 |

**成功响应：**

```json
{
  "status": "ok",
  "asset": "WIN",
  "win_balance": "100",
  "amount": "100",
  "message": "WIN 充值成功"
}
```

> USDT 人工充值仍用原接口 `POST /api/admin_dhb/admin_recharge`。

---

## 6. 配置项（运维）

`configs/*.yaml` → `wallet:`：

```yaml
win_contract: "0x..."   # WIN ERC20，用户链上充值必填
win_decimals: 18
deposit_address: "0x..." # 与 USDT 共用平台收款地址
# win_price 由预言机 / 管理后台维护，认购时读取内存价
```

---

## 7. 前端封装示例（TypeScript）

项目内已提供：`web/src/api/aix.ts`

```ts
import {
  getAixProfile,
  createWinRecharge,
  confirmWinRecharge,
  listWinRecharges,
  subscribeAix,
} from '@/api/aix'

// 画像
const profile = await getAixProfile()

// WIN 充值
const order = await createWinRecharge('100')
// ... 链上 transfer + personal_sign(order.message) ...
await confirmWinRecharge(order.recharge_id, txHash, signature)

// WIN 认购（amount 为 USDT 面额）
await subscribeAix('100', 'win')

// 记录
await listWinRecharges()
```

**Axios 直调示例：**

```ts
await axios.post('/v1/wallet/subscribe-aix', {
  amount: '100',
  pay_from: 'win',
}, {
  headers: { Authorization: `Bearer ${token}` },
})
```

---

## 8. 与旧接口关系

| 旧路径 / 行为 | 新能力 |
|---------------|--------|
| `POST /v1/wallet/recharge` | 仍为 **USDT** 充值 |
| `POST /v1/wallet/subscribe-aix` `pay_from=recharge\|reward` | 新增 `win` |
| 适配器 `app_server/buy*` | 支持 `pay_from=win` |
| 适配器新增 | `deposit_win` / `deposit_win_confirm` / `deposit_win_list` |
