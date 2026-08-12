import { BasicLayout } from '@/layouts'
export const asyncRouterMap = [
    {
        path: '/',
        name: 'index',
        component: BasicLayout,
        meta: { title: '首页' },
        redirect: '/home',
        children: [
            {
                path: '/home',
                name: 'home',
                component: () => import('@/views/home/index'),
                meta: { title: '数据统计', keepAlive: true, icon: "pie-chart", permission: ['dashboard'] },
            },
            {
                path: '/member',
                name: 'member',
                component: () => import('@/views/orders/member'),
                meta: { title: '用户数据', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/recharge',
                name: 'recharge',
                component: () => import('@/views/orders/recharge'),
                meta: { title: '充值列表', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/subscription',
                name: 'subscription',
                component: () => import('@/views/orders/subscription'),
                meta: { title: '订单列表', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/ordersList',
                name: 'ordersList',
                component: () => import('@/views/orders/ordersList'),
                meta: { title: '订单奖励', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/config',
                name: 'config',
                component: () => import('@/views/orders/config'),
                meta: { title: '配置项', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/exchangeList',
                name: 'exchangeList',
                component: () => import('@/views/orders/exchangeList'),
                meta: { title: '兑换记录', keepAlive: true, icon: "swap", permission: ['dashboard'] },
            },
            {
                path: '/settlement',
                name: 'settlement',
                component: () => import('@/views/orders/settlement'),
                meta: { title: '每日结算', keepAlive: true, icon: "setting", permission: ['dashboard'] },
            },
            {
                path: '/lookChildren',
                name: 'lookChildren',
                component: () => import('@/views/member/lookChildren'),
                hidden: true,
                meta: { title: '查看下级', keepAlive: true, permission: ['dashboard'] },
            },
        ]
    },
    {
        path: '*', redirect: '/404', hidden: true
    }
]
