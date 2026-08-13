# WIN 充值 — 前端接口文档

> 版本：v3 · 更新时间：2026-08-13  
> 适用：用户端 H5 / DApp  
> 说明：WIN 在 EOEO 链上是 **原生 Gas 币**（类似 ETH），**不是** ERC-20。充值通过合约 `BuySomething.buy` 附带 `msg.value` 完成，**无需** `approve` / `transfer`。

---

## 1. 通用约定

| 项 | 说明 |
|----|------|
| 请求格式 | `application/json` |
| 金额字段 | 一律用 **字符串**（如 `"10"`），禁止 JS 浮点直接运算 |
| 用户 Token | `Authorization: Bearer <jwt>`（亦可 Body `token` / Header `Access-Token`） |
| 局域网 API | `http://192.168.3.6:8081`（经 Nginx）或直连 `http://192.168.3.6:9000` |
| Radmin VPN | `http://26.3.54.166:8081` |

同域部署时推荐相对路径，例如 `/v1/wallet/aix-profile`。

---

## 2. 链信息

| 项 | 值 |
|----|-----|
| 网络名 | EOEO |
| RPC | `https://rpc1.eoeo.info` |
| Chain ID | `86233268`（十进制） / `0x523E714`（十六进制） |
| 原生币 | WIN（18 decimals，Gas + 充值） |
| 充值合约 | `BuySomething` |
| 合约地址 | `0x6A82cFF59da0cC4E31C13E92C396Cbdcafcf3cA9` |

### 2.1 钱包加链 / 切链

```js
await window.ethereum.request({
  method: 'wallet_addEthereumChain',
  params: [{
    chainId: '0x523E714',
    chainName: 'EOEO',
    nativeCurrency: { name: 'WIN', symbol: 'WIN', decimals: 18 },
    rpcUrls: ['https://rpc1.eoeo.info'],
  }],
})

await window.ethereum.request({
  method: 'wallet_switchEthereumChain',
  params: [{ chainId: '0x523E714' }],
})
```

项目内亦可复用 `web/src/tools/contract.ts`（`VITE_CHAINID` 默认 `86233268`）。

---

## 3. 业务规则

### 3.1 怎么充

```text
buy(num)  +  transaction value = num × 10^18
```

| 规则 | 说明 |
|------|------|
| `num` | 充值份数，**整数** |
| 最小值 | 以接口 `min_win_recharge` 为准（默认 `"10"`，管理端可改，且 **≥ 10**） |
| 应付原生币 | `num` 个 WIN（`value = num × 1e18` wei） |
| 合约校验 | `msg.value` 必须 **精确等于** `num × 1e18`，否则回滚 `bad value` |

示例：

| num | 应付 WIN | value（wei） |
|-----|----------|--------------|
| 10 | 10 | `10000000000000000000` |
| 100 | 100 | `100000000000000000000` |

### 3.2 合约内自动分账（前端不要自己拆）

| 比例 | 说明 |
|------|------|
| 98% | 主收款地址 |
| 1.5% | 分账地址 A |
| 0.5% | 分账地址 B |

前端只调用一次 `buy(num)`。

### 3.3 与 ERC-20 方案的区别（勿混用）

| | 原生 WIN（当前） | ERC-20 WIN（旧，勿用） |
|--|------------------|------------------------|
| 调用 | `buy` + `msg.value` | `approve` + `transfer` |
| 代币合约字段 | 无 | `win_contract` |
| 授权 | 不需要 | 需要 |

---

## 4. 获取充值最小值 / 余额

充值页打开时先拉资产画像，用服务端下发的最小值做校验（与管理端配置实时同步）。

- **Method:** `GET`
- **Path:** `/v1/wallet/aix-profile`
- **Auth:** ✅

**与 WIN 充值相关字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `win_balance` | string | 平台内 WIN 余额（入账后刷新可见） |
| `win_price` | number | WIN 现价（USDT/枚），认购折算用 |
| `min_win_recharge` | string | **WIN 充值最小值**（默认 `"10"`，管理端可配，≥10） |
| `min_usdt_recharge` | string | USDT 充值最小值（同规则，充值 USDT 时用） |

**响应节选：**

```json
{
  "win_balance": "50.0000000000000000",
  "win_price": 2,
  "min_win_recharge": "10",
  "min_usdt_recharge": "10"
}
```

**前端校验示例：**

```ts
const profile = await getAixProfile()
const minWin = Number(profile.min_win_recharge || 10)
if (!Number.isInteger(num) || num < minWin) {
  throw new Error(`充值数量不能小于 ${minWin}`)
}
```

封装：`web/src/api/aix.ts` → `getAixProfile()`。

---

## 5. 合约 ABI（前端最小集）

```json
[
  {
    "inputs": [{ "internalType": "uint256", "name": "num", "type": "uint256" }],
    "name": "buy",
    "outputs": [],
    "stateMutability": "payable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "getUserLength",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      { "internalType": "uint256", "name": "startIndex", "type": "uint256" },
      { "internalType": "uint256", "name": "endIndex", "type": "uint256" }
    ],
    "name": "getUsersByIndex",
    "outputs": [{ "internalType": "address[]", "name": "", "type": "address[]" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      { "internalType": "uint256", "name": "startIndex", "type": "uint256" },
      { "internalType": "uint256", "name": "endIndex", "type": "uint256" }
    ],
    "name": "getUsersAmountByIndex",
    "outputs": [{ "internalType": "uint256[]", "name": "", "type": "uint256[]" }],
    "stateMutability": "view",
    "type": "function"
  }
]
```

常量建议：

```ts
export const EOEO_CHAIN_ID = 86233268
export const EOEO_RPC = 'https://rpc1.eoeo.info'
export const BUY_SOMETHING = '0x6A82cFF59da0cC4E31C13E92C396Cbdcafcf3cA9'
// 或使用环境变量：import.meta.env.VITE_BUY
```

---

## 6. 前端充值流程（推荐）

```
1. 登录拿到 JWT
2. GET /v1/wallet/aix-profile → min_win_recharge、win_balance
3. 连接钱包，确认 chainId = 86233268
4. 用户输入整数 num（≥ min_win_recharge）
5. 检查原生 WIN 余额 ≥ num（并预留 Gas）
6. 调用 buy(num)，value = parseEther(String(num))
7. 等待交易成功，拿到 txHash
8. 刷新 GET /v1/wallet/aix-profile，展示最新 win_balance
```

> 入账说明：链上成功后，平台按 `BuySomething` 合约账本扫描入账到 `users.win_balance`（与 USDT 充值扫 `deposit_contract` 类似）。前端以 **刷新 profile** 为准，不要写死立刻到账；可短轮询 2～5 次。

### 6.1 参考现有 USDT 充值页

USDT 现网逻辑在 `web/src/views/subpage/components/rechargeDialog.vue`：先 `approve` 再 `BUY.send("buy", [count])`。  
**WIN 原生充值不要 approve**，只需：

```ts
await BUY.send("buy", [num], { value: parseEther(String(num)) })
// 具体封装以项目 Contract 工具是否支持 value 参数为准；
// 若不支持，请直接用 ethers / viem 发 payable 交易。
```

### 6.2 ethers 示例

```ts
import { BrowserProvider, Contract, parseEther } from 'ethers'

const BUY_ABI = ['function buy(uint256 num) payable']

async function rechargeWin(num: number, minWin: number) {
  if (!Number.isInteger(num) || num < minWin) {
    throw new Error(`num 必须是 ≥ ${minWin} 的整数`)
  }

  const provider = new BrowserProvider(window.ethereum)
  const network = await provider.getNetwork()
  if (Number(network.chainId) !== 86233268) {
    throw new Error('请切换到 EOEO 网络')
  }

  const signer = await provider.getSigner()
  const balance = await provider.getBalance(await signer.getAddress())
  const value = parseEther(String(num))
  if (balance < value) throw new Error('WIN 余额不足')

  const contract = new Contract(BUY_SOMETHING, BUY_ABI, signer)
  const tx = await contract.buy(num, { value })
  const receipt = await tx.wait()
  return { txHash: receipt.hash, num }
}
```

### 6.3 viem / wagmi 示例

```ts
import { parseEther } from 'viem'

await writeContract({
  address: '0x6A82cFF59da0cC4E31C13E92C396Cbdcafcf3cA9',
  abi: [{
    name: 'buy',
    type: 'function',
    stateMutability: 'payable',
    inputs: [{ name: 'num', type: 'uint256' }],
    outputs: [],
  }],
  functionName: 'buy',
  args: [BigInt(num)],
  value: parseEther(String(num)),
  chainId: 86233268,
})
```

---

## 7. 链上查询（调试 / 对账）

| 方法 | 说明 |
|------|------|
| `getUserLength()` | 历史充值条数 |
| `getUsersByIndex(start, end)` | 按区间取用户地址（含端点） |
| `getUsersAmountByIndex(start, end)` | 按区间取对应 `num` |

每次成功 `buy` 会追加一条记录（同一地址多次充值会有多条）。

---

## 8. 管理端配置

管理后台「AIX 配置项」：

| 名称 | 配置 id | 规则 |
|------|---------|------|
| WIN充值最小值 | `32` | 任意数值 **≥ 10**，热更新 |
| USDT充值最小值 | `31` | 同上 |

修改后用户端再次请求 `/v1/wallet/aix-profile` 即可读到新的 `min_win_recharge`。

---

## 9. 常见错误与前端提示

| 现象 | 前端提示建议 |
|------|----------------|
| `err num` | 充值数量不能低于最小值 |
| `bad value` | 支付金额必须等于 num 个 WIN（检查 `value = parseEther(num)`） |
| `send failed` | 链上转账失败，稍后重试 |
| 用户拒签 | 已取消 |
| 余额 / Gas 不足 | WIN 不够支付本金或 Gas |
| 错链 | 请切换到 EOEO（86233268） |
| profile 余额未变 | 等待扫链入账后重试刷新 |

---

## 10. UI 建议

- 输入：整数 `num`，`min = min_win_recharge`，步进 1  
- 展示：`应付 = num WIN` + 预估 Gas  
- 按钮：`充值 WIN`  
- 成功：展示 `txHash`，然后刷新 `win_balance`  
- **不要** 展示 ERC-20 合约地址，**不要** 引导 `approve`

---

## 11. WIN 替代充值钱包 USDT 认购

平台内 `win_balance` 可替代 **充值钱包 USDT** 认购。订单仍按 **USDT 面额** 记账；扣款按 **实时 `win_price`** 折成 WIN。

### 11.1 规则

| 项 | 说明 |
|----|------|
| 下单金额 `amount` | **USDT 面额**（例：认购 100U → `"100"`） |
| 扣 WIN 数量 | `WIN = amount ÷ win_price`（保留 8 位小数） |
| `win_price` | USDT/枚，由链上 Pair 预言机约每分钟更新（亦可管理端覆盖） |
| 订单本金 / 出局 | 仍按 USDT：本金=`amount`，出局上限=`amount × 倍数`（默认 4×） |
| 直推 | **与充值钱包认购相同**：`direct_base = amount`（USDT），产生直推奖 |
| 管理奖 | 与 USDT 认购相同，按本金级差计算 |

**示例：** 认购 `100` USDT，`win_price = 2` → 扣 `50` WIN；订单本金记 `100` USDT，出局 `400` USDT。

### 11.2 接口

- **Method:** `POST`
- **Path:** `/v1/wallet/subscribe-aix`
- **Auth:** ✅

**请求：**

```json
{
  "amount": "100",
  "pay_from": "win"
}
```

| 字段 | 说明 |
|------|------|
| `amount` | USDT 面额本金，≥ `min_subscribe`（默认 100） |
| `pay_from` | `recharge` \| `reward` \| **`win`** |

**`pay_from` 对照：**

| 值 | 扣款 | 直推 |
|----|------|------|
| `recharge` | `usdt_recharge` | ✅ |
| `reward` | `usdt_reward` | ❌ |
| `win` | `win_balance`（按价折算） | ✅（等同充值钱包） |

**成功响应关键字段：**

| 字段 | 说明 |
|------|------|
| `principal` | USDT 本金 |
| `exit_cap` | 出局上限 |
| `fund_source` | `"win"` |
| `from_win` | 实际扣减的 WIN |
| `win_price` | 下单时价格快照 |
| `win_balance` / `balance` | 扣款后 WIN 余额 |
| `direct_base` | 直推基数（= principal） |

**错误码：**

| Code | 说明 |
|------|------|
| `WIN_PRICE_NOT_CONFIGURED` | 价格未配置或为 0 |
| `INSUFFICIENT_WIN` | WIN 余额不足 |
| `MIN_SUBSCRIBE_LIMIT` | 低于最低认购额 |
| `INVALID_PAY_FROM` | pay_from 非法 |

### 11.3 前端封装

```ts
import { getAixProfile, subscribeAix } from '@/api/aix'

const profile = await getAixProfile()
const amountUsdt = 100
const needWin = amountUsdt / Number(profile.win_price) // 展示用
await subscribeAix(String(amountUsdt), 'win')
```

用户端节点页已支持三档支付：`报单` / `复投` / **`WIN 认购`**（`web/src/views/node.vue`）。

### 11.4 价格来源

`GET /v1/wallet/aix-profile` → `win_price`  
由 `WinPriceOracleJob` 从 Pair `0x15ad085fc866370b59936575565434b14d22281d` 轮询写入（约 1 分钟），认购时用内存实时价折算。

---

## 12. 变更摘要

| 项 | 说明 |
|----|------|
| 充值方式 | 原生币 `buy(num)` + `value`，非 ERC-20 |
| 最小值 | `GET /v1/wallet/aix-profile` → `min_win_recharge`（默认 10，管理端可配 ≥10） |
| 入账确认 | 链上成功后刷新 profile；由后端扫合约账本写入 `win_balance` |
| WIN 认购 | `pay_from=win`，`WIN = USDT ÷ win_price`，等同充值钱包直推 |
| 禁止 | `approve` / `transfer` / 依赖 `win_contract` 的旧 ERC-20 充值流程 |
