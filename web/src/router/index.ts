import { createRouter, createWebHashHistory } from "vue-router";
import Index from "@/views/index.vue";
import Count from "@/views/share/count.vue";
import Withdrawal from "@/views/withdrawal/index.vue";
import Node from "@/views/node.vue";
import Community from "@/views/community.vue";
import Wallet from "@/views/subpage/wallet.vue";
import Recharge from "@/views/recharge.vue";
import Mine from "@/views/mine.vue";

/**
 * 金牛主路径路由（与后端 app_server 已对接能力对齐）。
 * 未接老页统一重定向到首页，避免深链进 404/坏接口。
 */
const mainRoutes = [
  { path: "/", component: Index },
  { path: "/recharge", component: Recharge },
  { path: "/transfer", component: () => import("@/views/transfer.vue") },
  { path: "/node", component: Node },
  { path: "/community", component: Community },
  { path: "/wallet", component: Wallet },
  { path: "/withdrawal", component: Withdrawal },
  { path: "/count", component: Count },
  { path: "/profile", component: () => import("@/views/profile/index.vue") },
  { path: "/mine", component: Mine },
];

/** 旧路径 → 主路径（质押入口并到认购） */
const redirects: Array<{ path: string; redirect: string }> = [
  { path: "/pledge", redirect: "/node" },
  { path: "/home", redirect: "/" },
  { path: "/index", redirect: "/" },
  { path: "/idoDetails", redirect: "/" },
  { path: "/payment", redirect: "/" },
  { path: "/trade", redirect: "/" },
  { path: "/contact", redirect: "/" },
  { path: "/address", redirect: "/" },
  { path: "/shop", redirect: "/" },
  { path: "/Web3Shop", redirect: "/" },
  { path: "/order", redirect: "/" },
  { path: "/order/:id", redirect: "/" },
  { path: "/powerShop", redirect: "/" },
  { path: "/stat", redirect: "/" },
  { path: "/level", redirect: "/" },
  { path: "/powerOrder", redirect: "/" },
  { path: "/withdraw/:type", redirect: "/withdrawal" },
];

const routes = [
  ...mainRoutes,
  ...redirects,
  { path: "/:pathMatch(.*)*", redirect: "/" },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
