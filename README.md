# AIX Backend (Kratos)

Go Kratos backend for the AIX business (USDT recharge / subscribe / static AIX settlement / mgmt rewards). Module name: `backend`.

## Prerequisites

- Go 1.24+
- MySQL 8+

## Database

```bash
mysql -uroot -proot < scripts/aix.sql
```

This creates database `aix`, tables from `docs/AIX数据库设计.md`, seeds `system_config`, and today's `aix_prices` row (`price=1`).

Config: `configs/config.yaml` → `data.database.dbname: aix`.

## Run backend

```bash
cd c:\Users\86186\Desktop\new2
go build -o bin/server.exe ./cmd/server
.\bin\server.exe -conf ./configs
```

- HTTP: `0.0.0.0:9000`
- gRPC: `0.0.0.0:9100`
- Empty `wallet.rpc_url` = **dev mode** (recharge confirm credits without chain verify)

## Run admin (Vue)

```bash
cd c:\Users\86186\Desktop\new2\admin
npm install
# Windows PowerShell (Node 17+ OpenSSL):
$env:NODE_OPTIONS="--openssl-legacy-provider"
npm run serve
```

- Dev server: `http://localhost:9300/admin/` (proxies `/api` → `http://127.0.0.1:9000`)
- Login credentials are configured locally in `configs/config.yaml` under `auth.admin_account` / `auth.admin_password`.
- Prefer Node 16/18 LTS if `npm install` fails on Node 22 (`deasync` native build)

## Key APIs

| Method | Path | Notes |
|--------|------|--------|
| GET | `/v1/auth/challenge` | Wallet challenge |
| POST | `/v1/auth/login` | Signature login + invite code |
| GET | `/v1/wallet/balance` | `balance`=usdt_recharge, `released_balance`=usdt_reward, `claimed_amount`/`claimable_amount`=aix_balance |
| POST | `/v1/wallet/recharge` + confirm | Credits **usdt_recharge** only |
| POST | `/v1/wallet/subscribe-aix` | Body: `{amount, pay_from: recharge\|reward}` |
| POST | `/v1/wallet/transfer` | Upline/downline; USDT receiver → usdt_reward |
| GET | `/v1/wallet/orders` | principal / exit_cap / earned_total |
| GET | `/v1/wallet/rewards` | reward_logs |
| GET | `/v1/wallet/aix-price` | Current AIX price |

Withdraw / claim / products / eco APIs return `not supported in AIX`.

## Settlement

China midnight job (`internal/job`) runs static 0.5%/day AIX + mgmt W1–W10 differential rewards. Admin can trigger via legacy `/api/admin_dhb/settlement_trigger`.

## Docs

- `docs/AIX业务方案.md`
- `docs/AIX数据库设计.md`
