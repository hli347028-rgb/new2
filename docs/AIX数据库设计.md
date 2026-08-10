# AIX 数据库设计

> 版本：v1.0  
> 库名建议：`aix`  
> 引擎：InnoDB，字符集：`utf8mb4`  
> 金额字段统一使用 `DECIMAL(36,18)`，避免浮点误差。

## 1. 设计原则

1. **USDT 分账本**：`usdt_recharge`（可发直推）与 `usdt_reward`（不可发直推）物理分离。
2. **流水可审计**：充值、报单、奖励、转账均落表明细，余额变更可追溯。
3. **日结幂等**：`settlement_batches` 按结算日唯一，防止重复发放。
4. **配置热更**：业务参数存 `settings`（JSON），与基础设施配置分离。

## 2. 表清单

| 表名 | 用途 |
|---|---|
| `users` | 用户、邀请关系、余额、管理等级 |
| `orders` | 报单订单、出局进度 |
| `recharges` | 链上充值记录 |
| `transfers` | 上下级内部转账 |
| `withdrawals` | **仅 AIX** 提现申请（合约/打款待补） |
| `reward_logs` | 奖励 / 加速流水 |
| `aix_prices` | AIX 价格历史 |
| `settlement_batches` | 日结批次 |
| `settings` | 系统配置键值 |

## 3. ER 关系（逻辑）

```text
users 1───* users              (inviter_id → users.id)
users 1───* orders             (orders.user_id → users.id)
users 1───* recharges          (recharges.user_id → users.id)
users 1───* transfers          (from_user_id / to_user_id → users.id)
users 1───* withdrawals        (withdrawals.user_id → users.id，仅 AIX)
users 1───* reward_logs        (user_id / from_user_id → users.id)
orders 1───* reward_logs       (reward_logs.order_id → orders.id)
settlement_batches 1───* reward_logs  (reward_logs.batch_id → settlement_batches.id)
```

关联字段在下方各表「说明」列中标注（格式：`关联 表名.字段`）。应用层保证引用有效；DDL 示例未强制建物理外键，便于高并发写入与历史归档。

## 4. 枚举约定

### 4.1 管理等级 `mgmt_level`

| 值 | 含义 |
|---|---|
| 0 | W0 未达标 |
| 1–10 | W1–W10 |

### 4.2 订单状态 `orders.status`

| 值 | 含义 |
|---|---|
| `active` | 进行中，可拿静态 |
| `exited` | 已出局 |

### 4.3 资金来源 `orders.fund_source`（用户报单时选择，二选一）

| 值 | 含义 |
|---|---|
| `recharge` | 用户选择扣 `usdt_recharge`；`direct_base = principal` |
| `reward` | 用户选择扣 `usdt_reward`；`direct_base = 0` |

### 4.4 充值状态 `recharges.status`

| 值 | 含义 |
|---|---|
| `pending` | 待确认 |
| `confirmed` | 已入账 |
| `rejected` | 拒绝 / 无效 |

### 4.5 转账扣款账本 `transfers.pay_from`（USDT 时用户选择，二选一）

| 值 | 含义 |
|---|---|
| `recharge` | 付款方扣 `usdt_recharge` |
| `reward` | 付款方扣 `usdt_reward` |

### 4.6 转账币种 `transfers.asset`

| 值 | 含义 |
|---|---|
| `USDT` | USDT（收款方一律进 `usdt_reward`） |
| `AIX` | AIX |

### 4.7 奖励类型 `reward_logs.type`

| 值 | 含义 | 通常入账 |
|---|---|---|
| `static_aix` | 静态奖 | `aix_balance` |
| `dynamic_usdt` | 直推动态奖 | `usdt_reward` |
| `mgmt` | 管理奖 | `usdt_reward` |
| `exit_accel` | 出局加速记一笔（若与发放拆分） | 视实现 |
| `transfer_in` | 转入（可选记入流水） | `usdt_reward` / `aix_balance` |
| `transfer_out` | 转出 | 对应扣减 |

## 5. 表结构明细

### 5.1 `users`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键；被 `orders.user_id`、`recharges.user_id`、`transfers.from_user_id` / `to_user_id`、`reward_logs.user_id` / `from_user_id`、`users.inviter_id` 引用 |
| address | VARCHAR(42) | N | | 钱包地址，唯一；逻辑关联充值地址字段 |
| inviter_id | BIGINT UNSIGNED | Y | NULL | 直推上级；关联 `users.id`（自关联，可空表示无上级） |
| invite_code | VARCHAR(64) | N | | 邀请码（可用地址或短码），唯一 |
| usdt_recharge | DECIMAL(36,18) | N | 0 | 充值钱包；不可提现 |
| usdt_reward | DECIMAL(36,18) | N | 0 | 奖励钱包（直推/管理奖）；不可提现 |
| aix_balance | DECIMAL(36,18) | N | 0 | AIX 代币数（静态换算入账）；**唯一可提现** |
| static_usdt_total | DECIMAL(36,18) | N | 0 | 静态总收益（USDT 金本位累计，展示用） |
| mgmt_level | TINYINT UNSIGNED | N | 0 | W0–W10 |
| large_area_perf | DECIMAL(36,18) | N | 0 | 大区业绩缓存（最大直推分支的有效订单本金） |
| small_area_perf | DECIMAL(36,18) | N | 0 | 小区业绩缓存（其余分支的有效订单本金） |
| team_perf | DECIMAL(36,18) | N | 0 | 总业绩缓存（全部下级有效订单本金；订单出局后扣除） |
| status | TINYINT | N | 1 | 1正常 0禁用 |
| created_time | DATETIME(3) | N | CURRENT | |
| updated_time | DATETIME(3) | N | CURRENT | |

索引：

- `uk_users_address` UNIQUE(`address`)
- `uk_users_invite_code` UNIQUE(`invite_code`)
- `idx_users_inviter_id` (`inviter_id`)
- `idx_users_mgmt_level` (`mgmt_level`)

### 5.2 `orders`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键；被 `reward_logs.order_id` 引用 |
| user_id | BIGINT UNSIGNED | N | | 报单用户；关联 `users.id` |
| principal | DECIMAL(36,18) | N | | 报单本金 |
| exit_cap | DECIMAL(36,18) | N | | 出局上限 = principal × 倍数 |
| earned_total | DECIMAL(36,18) | N | 0 | 已计入出局的累计收益 |
| direct_base | DECIMAL(36,18) | N | 0 | 可计直推金额；`fund_source=recharge` 时等于 `principal`，否则为 0 |
| from_recharge | DECIMAL(36,18) | N | 0 | 实际扣 `usdt_recharge`；用户选 recharge 时 = principal，否则 0 |
| from_reward | DECIMAL(36,18) | N | 0 | 实际扣 `usdt_reward`；用户选 reward 时 = principal，否则 0 |
| fund_source | VARCHAR(16) | N | | 用户选择的扣款账本：`recharge` / `reward`（二选一，不自动拆分） |
| status | VARCHAR(16) | N | active | active/exited |
| exited_time | DATETIME(3) | Y | NULL | 出局时间 |
| created_time | DATETIME(3) | N | CURRENT | |
| updated_time | DATETIME(3) | N | CURRENT | |

索引：

- `idx_orders_user_status` (`user_id`, `status`)
- `idx_orders_status_created` (`status`, `created_time`)

### 5.3 `recharges`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| user_id | BIGINT UNSIGNED | N | | 充值用户；关联 `users.id` |
| amount | DECIMAL(36,18) | N | | 到账金额 |
| tx_hash | VARCHAR(66) | N | | 交易哈希 |
| from_address | VARCHAR(42) | Y | NULL | 付款地址；逻辑关联 `users.address`（链上付款方，可不等于本用户） |
| to_address | VARCHAR(42) | Y | NULL | 收款地址；平台充值地址（配置项，非用户表外键） |
| status | VARCHAR(16) | N | pending | |
| confirmed_time | DATETIME(3) | Y | NULL | |
| created_time | DATETIME(3) | N | CURRENT | |
| updated_time | DATETIME(3) | N | CURRENT | |

索引：

- `uk_recharges_tx_hash` UNIQUE(`tx_hash`)
- `idx_recharges_user_status` (`user_id`, `status`)

入账规则：`status` 变为 `confirmed` 时，`users.usdt_recharge += amount`。

### 5.4 `transfers`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| from_user_id | BIGINT UNSIGNED | N | | 付款方；关联 `users.id` |
| to_user_id | BIGINT UNSIGNED | N | | 收款方；关联 `users.id`（须与付款方为邀请树上下级） |
| asset | VARCHAR(16) | N | | USDT/AIX |
| amount | DECIMAL(36,18) | N | | |
| pay_from | VARCHAR(16) | Y | NULL | USDT 时必填：用户选择 `recharge` / `reward`；AIX 转账可空 |
| from_recharge_debit | DECIMAL(36,18) | N | 0 | 付款方扣充值账本；`pay_from=recharge` 时 = amount，否则 0 |
| from_reward_debit | DECIMAL(36,18) | N | 0 | 付款方扣奖励账本；`pay_from=reward` 时 = amount，否则 0 |
| to_credit_reward | DECIMAL(36,18) | N | 0 | 收款方进奖励账本（USDT） |
| to_credit_aix | DECIMAL(36,18) | N | 0 | 收款方进 AIX |
| remark | VARCHAR(255) | Y | NULL | |
| created_time | DATETIME(3) | N | CURRENT | |

索引：

- `idx_transfers_from_created` (`from_user_id`, `created_time`)
- `idx_transfers_to_created` (`to_user_id`, `created_time`)

业务约束（应用层）：

- `from_user_id` 与 `to_user_id` 必须为上下级
- USDT：`pay_from` 二选一，不自动拆分；收款方 `to_credit_reward = amount`，**禁止**增加收款方 `usdt_recharge`

### 5.5 `withdrawals`（仅 AIX）

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| user_id | BIGINT UNSIGNED | N | | 提现用户；关联 `users.id` |
| asset | VARCHAR(16) | N | AIX | 固定 `AIX`；禁止 USDT |
| amount | DECIMAL(36,18) | N | | 申请数量 |
| fee | DECIMAL(36,18) | N | 0 | 手续费（待定） |
| pay_amount | DECIMAL(36,18) | N | | 实付/实发数量 |
| to_address | VARCHAR(42) | N | | 收款地址 |
| tx_hash | VARCHAR(66) | Y | NULL | 链上哈希；合约未就绪前可空 |
| status | VARCHAR(16) | N | pending | pending/paid/rejected 等 |
| remark | VARCHAR(255) | Y | NULL | |
| created_time | DATETIME(3) | N | CURRENT | |
| updated_time | DATETIME(3) | N | CURRENT | |

索引：`idx_withdrawals_user_created` (`user_id`, `created_time`)

说明：AIX 代币合约地址、打款通道待配置，字段/配置占位即可。

### 5.6 `reward_logs`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| user_id | BIGINT UNSIGNED | N | | 受益人；关联 `users.id` |
| from_user_id | BIGINT UNSIGNED | Y | NULL | 来源用户（直推时为下级）；关联 `users.id`，可空 |
| order_id | BIGINT UNSIGNED | Y | NULL | 关联订单；关联 `orders.id`，可空 |
| batch_id | BIGINT UNSIGNED | Y | NULL | 日结批次；关联 `settlement_batches.id`，可空 |
| type | VARCHAR(32) | N | | 见枚举 |
| asset | VARCHAR(16) | N | | USDT/AIX |
| amount | DECIMAL(36,18) | N | | 实发金额 |
| base_amount | DECIMAL(36,18) | Y | NULL | 计算基数 |
| rate | DECIMAL(36,18) | Y | NULL | 使用比例 |
| exit_applied | DECIMAL(36,18) | N | 0 | 计入出局的金额 |
| meta | JSON | Y | NULL | 扩展（等级、结算日等） |
| settlement_date | DATE | Y | NULL | 归属结算日 |
| created_time | DATETIME(3) | N | CURRENT | |

索引：

- `idx_reward_user_type_created` (`user_id`, `type`, `created_time`)
- `idx_reward_batch` (`batch_id`)
- `idx_reward_order` (`order_id`)
- `idx_reward_settlement_date` (`settlement_date`)
- 可选幂等：`uk_reward_static_daily` UNIQUE(`order_id`, `type`, `settlement_date`)（仅对静态奖启用时注意 NULL 行为）

### 5.7 `aix_prices`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| price | DECIMAL(36,18) | N | | USDT/AIX，建议 ≥ 前值；日结时写入 `settlement_batches.aix_price` |
| effective_date | DATE | N | | 生效日，唯一；与 `settlement_batches.settlement_date` 按日对应 |
| remark | VARCHAR(255) | Y | NULL | |
| created_time | DATETIME(3) | N | CURRENT | |

索引：

- `uk_aix_prices_effective_date` UNIQUE(`effective_date`)

### 5.8 `settlement_batches`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键；被 `reward_logs.batch_id` 引用 |
| settlement_date | DATE | N | | 结算日，唯一；逻辑对应 `reward_logs.settlement_date` |
| aix_price | DECIMAL(36,18) | N | | 当日使用价格；可与 `aix_prices.price`（同生效日）一致 |
| status | VARCHAR(16) | N | running | running/success/failed |
| static_count | INT UNSIGNED | N | 0 | 静态发放笔数 |
| static_amount | DECIMAL(36,18) | N | 0 | 静态发放合计（AIX 或金本位，实现时固定一种并在 remark 说明） |
| mgmt_count | INT UNSIGNED | N | 0 | |
| mgmt_amount | DECIMAL(36,18) | N | 0 | |
| started_time | DATETIME(3) | Y | NULL | |
| finished_time | DATETIME(3) | Y | NULL | |
| error_msg | VARCHAR(512) | Y | NULL | |
| created_time | DATETIME(3) | N | CURRENT | |

索引：

- `uk_settlement_date` UNIQUE(`settlement_date`)

### 5.9 `settings`

| 字段 | 类型 | 空 | 默认 | 说明 |
|---|---|---|---|---|
| id | BIGINT UNSIGNED | N | AI | 主键 |
| `key` | VARCHAR(64) | N | | 如 `system_config` |
| value | JSON | N | | 配置快照 |
| updated_time | DATETIME(3) | N | CURRENT | |
| created_time | DATETIME(3) | N | CURRENT | |

索引：

- `uk_settings_key` UNIQUE(`key`)

#### `system_config` JSON 建议结构

```json
{
  "static_rate": 0.5,
  "exit_multiplier": 4,
  "direct_rate": 0.5,
  "mgmt_thresholds": [5000, 20000, 50000, 200000, 500000, 1500000, 4000000, 8000000, 15000000, 30000000],
  "mgmt_rates": [0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1],
  "aix_price_initial": 1,
  "mgmt_counts_toward_exit": true,
  "min_subscribe": "100"
}
```

## 6. 关键资金流 ↔ 表

| 业务动作 | 写表 | 余额变化 |
|---|---|---|
| 充值确认 | `recharges` | `usdt_recharge ↑` |
| 报单 | `orders` + 可选流水 | 按用户选择扣 `usdt_recharge` 或 `usdt_reward`；仅 `fund_source=recharge` 时上级 `usdt_reward ↑` + `reward_logs(dynamic_usdt)` |
| 静态日结 | `reward_logs` + `settlement_batches` + `orders` | `aix_balance ↑`，`static_usdt_total ↑`，`earned_total ↑`（出局仍按 USDT），可能 `exited` |
| 管理奖 | `reward_logs` + `users` | `usdt_reward ↑`，等级/业绩字段更新 |
| 充值→奖励（同账户） | 余额字段 | `usdt_recharge ↓`，`usdt_reward ↑`；不产生直推 |
| 转账 | `transfers` | 仅邀请链祖先/后代之间可互转；付款方 `usdt_reward ↓`，收款方 `usdt_reward ↑`；不同分支不可互转 |
| AIX 提现 | `withdrawals` | `aix_balance ↓`；`asset=AIX`；合约/tx 未就绪时可 `pending` 且 `tx_hash` 空 |

**禁止** USDT 提现申请；`withdrawals.asset` 固定为 `AIX`。

## 7. 直推与账本约束（库表层落地）

1. 任何使 `usdt_recharge` 增加的路径，**仅允许** `recharges` 确认入账。
2. `transfers`、奖励发放 **禁止** 增加收款方 / 受益人的 `usdt_recharge`。
3. 报单扣款由用户选择：`fund_source` 为 `recharge` 或 `reward` 之一；`from_recharge + from_reward = principal`，且二者仅一方等于 `principal`。
4. `orders.direct_base` 必须等于 `from_recharge`（选 reward 时二者均为 0）。
5. 直推奖只根据 `direct_base` 计算，结果只进上级 `usdt_reward`。

## 8. 附录：MySQL DDL 示例

```sql
CREATE DATABASE IF NOT EXISTS aix DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE aix;

CREATE TABLE users (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  address           VARCHAR(42)     NOT NULL,
  inviter_id        BIGINT UNSIGNED NULL,
  invite_code       VARCHAR(64)     NOT NULL,
  usdt_recharge     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  usdt_reward       DECIMAL(36,18)  NOT NULL DEFAULT 0,
  aix_balance       DECIMAL(36,18)  NOT NULL DEFAULT 0,
  static_usdt_total DECIMAL(36,18)  NOT NULL DEFAULT 0,
  mgmt_level        TINYINT UNSIGNED NOT NULL DEFAULT 0,
	large_area_perf  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  small_area_perf   DECIMAL(36,18)  NOT NULL DEFAULT 0,
  team_perf         DECIMAL(36,18)  NOT NULL DEFAULT 0,
  status            TINYINT         NOT NULL DEFAULT 1,
  created_time      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_address (address),
  UNIQUE KEY uk_users_invite_code (invite_code),
  KEY idx_users_inviter_id (inviter_id),
  KEY idx_users_mgmt_level (mgmt_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orders (
  id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id        BIGINT UNSIGNED NOT NULL,
  principal      DECIMAL(36,18)  NOT NULL,
  exit_cap       DECIMAL(36,18)  NOT NULL,
  earned_total   DECIMAL(36,18)  NOT NULL DEFAULT 0,
  direct_base    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_recharge  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_reward    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  fund_source    VARCHAR(16)     NOT NULL,
  status         VARCHAR(16)     NOT NULL DEFAULT 'active',
  exited_time    DATETIME(3)     NULL,
  created_time   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_orders_user_status (user_id, status),
  KEY idx_orders_status_created (status, created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE recharges (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id          BIGINT UNSIGNED NOT NULL,
  amount           DECIMAL(36,18)  NOT NULL,
  tx_hash          VARCHAR(66)     NOT NULL,
  from_address     VARCHAR(42)     NULL,
  to_address       VARCHAR(42)     NULL,
  status           VARCHAR(16)     NOT NULL DEFAULT 'pending',
  confirmed_time   DATETIME(3)     NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_recharges_tx_hash (tx_hash),
  KEY idx_recharges_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transfers (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  from_user_id         BIGINT UNSIGNED NOT NULL,
  to_user_id           BIGINT UNSIGNED NOT NULL,
  asset                VARCHAR(16)     NOT NULL,
  amount               DECIMAL(36,18)  NOT NULL,
  pay_from             VARCHAR(16)     NULL,
  from_recharge_debit  DECIMAL(36,18)  NOT NULL DEFAULT 0,
  from_reward_debit    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  to_credit_reward     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  to_credit_aix        DECIMAL(36,18)  NOT NULL DEFAULT 0,
  remark               VARCHAR(255)    NULL,
  created_time         DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_transfers_from_created (from_user_id, created_time),
  KEY idx_transfers_to_created (to_user_id, created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE withdrawals (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id       BIGINT UNSIGNED NOT NULL,
  asset         VARCHAR(16)     NOT NULL DEFAULT 'AIX',
  amount        DECIMAL(36,18)  NOT NULL,
  fee           DECIMAL(36,18)  NOT NULL DEFAULT 0,
  pay_amount    DECIMAL(36,18)  NOT NULL,
  to_address    VARCHAR(42)     NOT NULL,
  tx_hash       VARCHAR(66)     NULL,
  status        VARCHAR(16)     NOT NULL DEFAULT 'pending',
  remark        VARCHAR(255)    NULL,
  created_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_withdrawals_user_created (user_id, created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE reward_logs (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id          BIGINT UNSIGNED NOT NULL,
  from_user_id     BIGINT UNSIGNED NULL,
  order_id         BIGINT UNSIGNED NULL,
  batch_id         BIGINT UNSIGNED NULL,
  type             VARCHAR(32)     NOT NULL,
  asset            VARCHAR(16)     NOT NULL,
  amount           DECIMAL(36,18)  NOT NULL,
  base_amount      DECIMAL(36,18)  NULL,
  rate             DECIMAL(36,18)  NULL,
  exit_applied     DECIMAL(36,18)  NOT NULL DEFAULT 0,
  meta             JSON            NULL,
  settlement_date  DATE            NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_reward_user_type_created (user_id, type, created_time),
  KEY idx_reward_batch (batch_id),
  KEY idx_reward_order (order_id),
  KEY idx_reward_settlement_date (settlement_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE aix_prices (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  price           DECIMAL(36,18)  NOT NULL,
  effective_date  DATE            NOT NULL,
  remark          VARCHAR(255)    NULL,
  created_time    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_aix_prices_effective_date (effective_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE settlement_batches (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  settlement_date  DATE            NOT NULL,
  aix_price        DECIMAL(36,18)  NOT NULL,
  status           VARCHAR(16)     NOT NULL DEFAULT 'running',
  static_count     INT UNSIGNED    NOT NULL DEFAULT 0,
  static_amount    DECIMAL(36,18)  NOT NULL DEFAULT 0,
  mgmt_count       INT UNSIGNED    NOT NULL DEFAULT 0,
  mgmt_amount      DECIMAL(36,18)  NOT NULL DEFAULT 0,
  started_time     DATETIME(3)     NULL,
  finished_time    DATETIME(3)     NULL,
  error_msg        VARCHAR(512)    NULL,
  created_time     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_settlement_date (settlement_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE settings (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `key`         VARCHAR(64)     NOT NULL,
  value         JSON            NOT NULL,
  created_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_time  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_settings_key (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 9. 文档关系

- 业务含义与验收标准见 [`AIX业务方案.md`](./AIX业务方案.md)
- 本文仅定义数据落库方式；应用层须强制执行「用户选择 `fund_source=recharge` 报单才可发直推」，且仅允许 `withdrawals.asset=AIX`
