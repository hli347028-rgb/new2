# AIX 链上任务改用 crontab

## 原则

1. **先关进程内轮询，再装 crontab**，禁止双开（重复扫链 / 重复读价）。
2. 充值写入按 `tx_hash` 幂等：已存在记录不插库、不加余额；可安全每分钟触发。

## 配置（`/opt/aix/configs/config.yaml`）

```yaml
wallet:
  recharge_monitor_enabled: false   # 关闭进程内 USDT/WIN 充值 Ticker
  win_price_oracle_enabled: false   # 关闭进程内 WIN 价格循环
```

HTTP 触发（本机，立即返回，后台跑一轮）：

- `GET /api/admin_dhb/deposit_only` — USDT
- `GET /api/admin_dhb/deposit_only_win` — 原生 WIN
- `GET /api/admin_dhb/win_price_oracle` — 拉 Pair 写价

每轮内部逻辑（配置可调，默认）：

- 共查询 **10** 次
- 相邻间隔 **5** 秒
- 约 45 秒内完成；若上一轮未结束，再次触发返回 `accepted:false`

对应配置：

- 充值：`recharge_scan_queries_per_cycle` / `recharge_scan_query_interval_seconds`
- 价格：`win_price_queries_per_cycle` / `win_price_query_interval_seconds`

## crontab

```cron
* * * * * /opt/aix/scripts/aix-chain-jobs.sh >> /opt/aix/logs/cron-chain.log 2>&1
```

脚本：`scripts/cron/aix-chain-jobs.sh`（默认 `AIX_HTTP=http://127.0.0.1:9000`）。  
crontab **每分钟只调一次**；10×5s 轮询在 `aix-server` 进程内完成。

确保 `/opt/aix/logs` 属主为运行 crontab 的用户（如 `ubuntu:ubuntu`），否则无法写 `cron-chain.log`。

确认旧的 `aix-deposit-only.timer` 为 disabled。

## 验收

1. 重启后日志含 `recharge monitor disabled` 与 `win price oracle disabled`。
2. 静置无「进程内 Ticker」自动启动；仅 cron/HTTP 触发后出现 `cycle started` / `#n/10`。
3. curl 三个接口返回 `accepted:true, queries:10, interval_seconds:5`。
4. 触发后约 1 分钟内日志应有约 10 次 sync/price；重复触发不重复入账（`tx_hash` 幂等）。
