# AIX 用户端 H5

精简黑金风格用户端（Vue 3 + Vite + Vant），对接本仓库后端 `/v1`。

## 运行

```bash
# 先启动后端（HTTP 9000）
# 再启动本前端
cd web
npm install
npm run dev
```

浏览器打开：http://localhost:9200/

- 开发代理：`/v1` → `http://127.0.0.1:9000`
- 需安装 MetaMask；BSC 链上充值依赖 `.env.prod` 中的 `VITE_USDT`

## 页面

| 路径 | 说明 |
|------|------|
| `/` | 首页 / 价格 / 规则 |
| `/order` | 报单（充值/奖励账本） |
| `/mine` | 资产与奖励流水 |
| `/recharge` | USDT 充值 |
| `/transfer` | 充值→奖励 / 上下级转账 |
| `/withdraw` | 仅提现 AIX |
| `/team` | 团队与邀请链接 |
