# AIX 业务方案

> 版本：v1.0  
> 范围：仅后端业务设计，不含前端展示与本次代码实现。

## 1. 产品概述

AIX 是一套基于 USDT 充值报单、链下日结收益的投资/推荐业务系统。用户通过钱包地址识别身份，链上充值 USDT 后报单，获得：

- **静态奖**：按报单本金每日 0.5%，金本位发放 AIX
- **动态奖（直推）**：下级有效充值报单时，直推上级获得 50% USDT
- **管理奖**：按小区业绩达到 W1–W10 后，参与级差分配

出局规则为报单本金的 **4 倍**；静态与动态收益均计入出局进度，满额后订单出局，需复投才能继续产生收益。

## 2. 范围说明

| 项 | 说明 |
|---|---|
| 做 | 业务规则、账本模型、日结逻辑、接口清单、验收用例 |
| 不做（本阶段） | 代码开发、前端页面、真实建库部署 |
| 链上职责 | BSC USDT 充值核验；用户仅可提现 AIX（合约地址待补） |
| 收益结算 | 全部链下，日结 Job（建议 Asia/Shanghai 零点） |

## 3. 资产模型

每个用户维护三类余额 + 一项静态 USDT 累计（展示用）：

| 字段 | 含义 | 入账来源 | 可否触发直推 / 提现 |
|---|---|---|---|
| `usdt_recharge` | 充值钱包（USDT） | 仅链上充值确认 | 报单扣减部分 **可以** 触发直推；**不可提现** |
| `usdt_reward` | 奖励钱包（USDT） | 直推奖、管理奖、充值钱包转入、上下级划转入账 | **不可以** 触发直推；**不可提现** |
| `aix_balance` | AIX 代币数 | 静态释放按 `aix_price` 换算入账 | **仅此字段可提现**（合约信息待补） |
| `static_usdt_total` | 静态总收益（USDT 金本位累计） | 日结静态发放时累加金本位金额 | 仅展示 / 对账，不单独提现 |

另有内部定价：`aix_price`，表示 **1 枚 AIX 值多少 USDT**（由后台「配置项」手动调整，暂不自动涨价）：

| `aix_price` | USDT : AIX | 含义 |
|---|---|---|
| `1` | **1 : 1** | 1 USDT 金本位静态 → 入账 1 AIX |
| `2` | **2 : 1** | 2 USDT 金本位静态 → 入账 1 AIX（即发 AIX = 金本位 ÷ 价格） |

公式：`aix_amount = usdt_gold / aix_price`。修改配置后同步写入当日 `aix_prices`，日结按该价换算。

### 3.1 报单扣款（用户选择账本）

报单 / 复投时，用户必须指定从哪个账本扣款，**二选一，不自动拆分**：

| 用户选择 `pay_from` | 扣减账本 | `direct_base` | 是否产生直推 |
|---|---|---|---|
| `recharge` | `usdt_recharge` | = 报单本金 | **是**（×50%） |
| `reward` | `usdt_reward` | = 0 | **否** |

所选账本余额不足则报单失败。上下级转账扣 USDT 时同样由付款方指定扣 `usdt_recharge` 或 `usdt_reward`；收款方一律只进 `usdt_reward`。

## 4. 静态奖

| 项 | 规则 |
|---|---|
| 利率 | 固定每日 **0.5%**（相对报单本金，金本位） |
| 金本位 | `usdt_gold = principal × 0.5%`（裁剪后计入出局 `earned_total` / `static_usdt_total`） |
| 发放资产 | **AIX**：`aix_amount = usdt_gold / aix_price`，入账 `aix_balance` |
| 价格来源 | 现阶段由后台配置项 **AIX价格(USDT/枚)** 调整；日结读当日价 |
| 发放对象 | 状态为 `active` 的报单订单 |
| 出局 | 按 **USDT 金本位** 累计，与代币价格无关 |

示例：本金 1000 → 日静态金本位 5 USDT；价格=1 得 5 AIX；价格=2 得 2.5 AIX。

## 5. 动态奖（直推）

| 项 | 规则 |
|---|---|
| 比例 | **50%** |
| 计算基数 | 下级报单时的 `direct_base`（仅充值账本扣减额） |
| 发放资产 | USDT，计入上级 **`usdt_reward`** |
| 额度裁剪 | 不得超过上级当前活跃订单剩余出局额度；超出部分不发或按规则截断 |

### 5.1 直推触发规则（关键）

**只有本账户链上充值确认的钱，在报单时才产生直推奖。**

| 资金动作 | 是否产生直推奖 |
|---|---|
| 链上充值 → 本账户报单 | **是**（`direct_base × 50%` 给直推上级） |
| 动态奖金 → 本账户复投 / 报单 | **否** |
| 动态奖金 → 划转到其他账户 → 对方再报单 | **否** |
| 管理奖等非充值入账 → 报单 | **否** |
| 内部划转入账 → 报单 | **否** |

### 5.2 直推伪代码

```text
onSubscribe(user, amount, pay_from):  // pay_from = recharge | reward，用户选择
  if pay_from == recharge:
    require user.usdt_recharge >= amount
    debit usdt_recharge -= amount
    direct_base = amount
    fund_source = recharge
  else:
    require user.usdt_reward >= amount
    debit usdt_reward -= amount
    direct_base = 0
    fund_source = reward

  create order(principal=amount, exit_cap=amount*4, direct_base, fund_source, ...)

  if direct_base > 0 and user.inviter exists:
    pay = min(direct_base * 0.5, inviter.remaining_exit_capacity)
    credit inviter.usdt_reward += pay
    accelerate inviter active order earned_total by pay
```

## 6. 出局与复投

| 项 | 规则 |
|---|---|
| 出局倍数 | **本金 × 4** |
| 计入收益 | 静态奖 + 动态奖（含因直推/管理等加速计入订单的部分，以产品最终定义为准；至少静态与直推动态计入） |
| 出局后果 | 订单 `status = exited`，不再产生静态奖 |
| 复投 | 用户再次报单；资金若来自 `usdt_reward`，不产生新直推 |

示例：

- A 报单 100 → 出局额度 400；累计收益达 400（相对本金回收叙事可为本金+收益，实现以 `earned_total >= exit_cap` 为准）后出局，需复投
- B 报单 1000 → 出局额度 4000

### 6.1 出局伪代码

```text
onReward(user, order, amount, kind):
  remain = order.exit_cap - order.earned_total
  pay = min(amount, remain)
  credit(kind)  // static → aix_balance; dynamic/mgmt → usdt_reward
  order.earned_total += pay
  if order.earned_total >= order.exit_cap:
    order.status = exited
```

## 7. 管理奖（W1–W10）

依据用户 **小区业绩**（USDT）评定等级，日结时按级差思路发放管理奖（具体级差底数建议取当日团队相关静态/业绩池，开发阶段在配置中固化）。

| 级别 | 小区业绩（USDT） | 比例 |
|---|---|---|
| W1 | 5,000 | 20% |
| W2 | 20,000 | 30% |
| W3 | 50,000 | 40% |
| W4 | 200,000 | 50% |
| W5 | 500,000 | 60% |
| W6 | 1,500,000 | 70% |
| W7 | 4,000,000 | 80% |
| W8 | 8,000,000 | 90% |
| W9 | 15,000,000 | 100% |
| W10 | 30,000,000 | 110% |

- 未达 W1：等级 W0，不拿管理奖  
- 发放进入 `usdt_reward`  
- 管理奖是否计入出局进度：建议 **计入**（与「动态静态都加速出局」一致），最终以配置开关控制  

## 8. 上下级转账

| 项 | 规则 |
|---|---|
| 关系约束 | 仅允许邀请树上的上下级互转 |
| 币种 | USDT（及可选 AIX） |
| USDT 扣款 | 付款方指定扣 `usdt_recharge` 或 `usdt_reward`（二选一） |
| USDT 入账 | 收款方 **只增加 `usdt_reward`** |
| 直推影响 | 划转 **永不** 创造新的可发直推额度（即使付款方扣的是 `usdt_recharge`） |

目的：防止通过划转「洗」出直推奖。

### 8.1 同账户：充值钱包 → 奖励钱包

| 项 | 规则 |
|---|---|
| 动作 | 用户（或管理端代操作）将本账户 `usdt_recharge` 划入 `usdt_reward` |
| 直推 | **不产生** 直推；划出的充值余额不再具备「可发直推」属性 |
| 用途 | 便于用充值资金走奖励账本复投 / 其它 USDT 用途 |

## 9. 充值、报单与提现

| 动作 | 规则 |
|---|---|
| 充值 | 用户向平台地址转 USDT → 提交 tx 核验 → 增加 `usdt_recharge` |
| 报单 / 复投 | 用户选择扣 `usdt_recharge` 或 `usdt_reward` → 创建订单；仅选充值账本时按 `direct_base` 发直推 |
| 出局目标 | 仍按 **USDT** 金本位累计（`earned_total` / `exit_cap`），与 AIX 代币提现无关 |
| 提现 | **仅允许提现 AIX**（`aix_balance`）；禁止提现 USDT。AIX 合约地址、链上打款参数待配置，未就绪前申请可落库为 `pending`，`tx_hash` 留空 |

直推奖、管理奖一律以 **USDT** 进入 `usdt_reward`；静态释放一律换算为 **AIX** 进入 `aix_balance`。

## 10. 日结流程（建议）

时区：Asia/Shanghai，每日零点（可支持管理端手动触发）。

1. 抬升或固化当日 `aix_price`（若启用单向涨价）
2. 遍历 `active` 订单，按本金 × 0.5% 发放静态 AIX，累加出局
3. 刷新全网/相关用户小区业绩与 W 等级
4. 按管理奖规则发放（进 `usdt_reward`，并处理出局加速）
5. 写入 `settlement_batches` 与明细流水，保证同日幂等

## 11. 建议 API 清单（后续开发对照）

### 11.1 用户侧 `/v1`

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/v1/auth/challenge` | 钱包登录挑战码 |
| POST | `/v1/auth/login` | 签名登录，可选邀请人 |
| GET | `/v1/wallet/profile` | 余额、等级、小区业绩 |
| POST | `/v1/wallet/recharge/confirm` | 充值 tx 确认 |
| POST | `/v1/wallet/subscribe` | 报单 / 复投（必传 `pay_from=recharge\|reward`） |
| POST | `/v1/wallet/transfer` | 上下级转账（USDT 必传扣款账本） |
| POST | `/v1/wallet/recharge-to-reward` | 充值钱包 → 奖励钱包（同账户） |
| POST | `/v1/wallet/withdraw-aix` | 提现 AIX（合约信息待补） |
| GET | `/v1/wallet/orders` | 订单与出局进度 |
| GET | `/v1/wallet/rewards` | 奖励流水 |
| GET | `/v1/wallet/aix-price` | 当前 AIX 价格（`aix_contract` 待补） |
| GET | `/v1/wallet/team` | 直推 / 团队概览 |

### 11.2 管理侧

| 能力 | 说明 |
|---|---|
| 登录 | 账号密码 |
| 系统配置 | 静态利率、出局倍数、直推比例、W 档位、涨价参数 |
| 用户 / 订单 / 流水查询 | 只读运维 |
| 结算触发 | 手动跑日结 |

## 12. 关键配置项（`system_config`）

| 键 | 建议默认值 |
|---|---|
| `static_rate` | 0.5（%） |
| `exit_multiplier` | 4 |
| `direct_rate` | 0.5（50%） |
| `mgmt_thresholds` | [5000, 20000, 50000, 200000, 500000, 1500000, 4000000, 8000000, 15000000, 30000000] |
| `mgmt_rates` | [0.20, 0.30, 0.40, 0.50, 0.60, 0.70, 0.80, 0.90, 1.00, 1.10] |
| `aix_price_initial` / 配置项「AIX价格」 | 1（表示 1 USDT/枚；改成 2 即为 2:1） |
| `mgmt_counts_toward_exit` | true |

## 13. 验收用例

1. **充值账本报单有直推**：用户 B 链上充值 1000，选择 `pay_from=recharge` 报单，直推上级 A 的 `usdt_reward` 增加 500（若出局额度足够）。
2. **奖励账本报单无直推**：A 选择 `pay_from=reward` 用直推所得 500 复投，A 的上级不因该 500 获得直推。
3. **划转后再报单无直推**：A 将余额划给下级 C（收款进 `usdt_reward`），C 无论选哪个账本，只要不是本人链上充值进 `usdt_recharge` 再选 `recharge` 报单，则不产生直推；C 用划转入账的 `usdt_reward` 报单无直推。
4. **静态日结**：活跃订单本金 1000，日结增加金本位 5 USDT 等值的 AIX，并推进 `earned_total`。
5. **4 倍出局**：订单 `earned_total` 达到 `principal × 4` 后变为 `exited`，次日不再发静态。
6. **管理奖升级**：小区业绩跨过档位后 `mgmt_level` 更新，日结产生管理奖流水。
7. **转账约束**：同一邀请链的祖先/后代之间可从奖励钱包互转；不同分支、同级或无关系用户转账失败；成功转账收款方只增 `usdt_reward`。

## 14. 文档关系

- 业务规则以本文为准
- 表结构、字段、DDL 见同目录 [`AIX数据库设计.md`](./AIX数据库设计.md)
