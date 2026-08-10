import { createRouter, createWebHashHistory } from 'vue-router'
import Home from '@/views/home.vue'
import Order from '@/views/order.vue'
import Mine from '@/views/mine.vue'
import Team from '@/views/team.vue'
import Recharge from '@/views/recharge.vue'
import Transfer from '@/views/transfer.vue'
import Withdraw from '@/views/withdraw.vue'

const routes = [
  { path: '/', name: 'home', component: Home },
  { path: '/order', name: 'order', component: Order },
  { path: '/mine', name: 'mine', component: Mine },
  { path: '/team', name: 'team', component: Team },
  { path: '/recharge', name: 'recharge', component: Recharge },
  { path: '/transfer', name: 'transfer', component: Transfer },
  { path: '/withdraw', name: 'withdraw', component: Withdraw },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
