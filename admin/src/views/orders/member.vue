<template>
    <PageView>
        <a-card title="用户数据">
            <a-row :gutter="10" class="inputGroup">
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-input v-model="searchData.address" placeholder="账户地址" @keyup.enter="getListTwo" />
                </a-col>
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-button-group>
                        <a-button type="primary" :loading="loading" @click="getListTwo">确定筛选</a-button>
                    </a-button-group>
                </a-col>
            </a-row>
            <a-table :loading="loading" :columns="columns" :dataSource="data" :pagination="{ total, pageSize, current }"
                @change="changePagination" bordered :scroll="{ x: true }">
            </a-table>
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import Gai from '../../api/Gai'
import listMixin from '../mixin/listMixin'

export default {
    name: 'member',
    mixins: [listMixin],
    data() {
        return {
            columns: [
                {
                    title: '创建时间',
                    dataIndex: 'createdAt',
                },
                {
                    title: '地址',
                    dataIndex: 'address',
                },
                {
                    title: '充值钱包',
                    dataIndex: 'usdt_recharge',
                },
                {
                    title: '奖励钱包',
                    dataIndex: 'usdt_reward',
                },
                {
                    title: 'AIX代币数',
                    dataIndex: 'aix_balance',
                },
                {
                    title: 'WIN代币数',
                    dataIndex: 'win_balance',
                    customRender: (v) => v || '0'
                },
                {
                    title: '待释放管理奖',
                    dataIndex: 'pending_mgmt_reward',
                    customRender: (v) => v || '0'
                },
                {
                    title: '静态总收益',
                    dataIndex: 'static_usdt_total',
                },
                {
                    title: '总订单',
                    dataIndex: 'amountUsdtCurrent',
                },
                {
                    title: '管理级别',
                    dataIndex: 'mgmt_level',
                    customRender: (v, row) => {
                        if (row && row.vip) return row.vip
                        const n = parseInt(v, 10) || 0
                        return 'W' + n
                    }
                },
                {
					title: '大区业绩',
					dataIndex: 'large_area_perf',
				},
				{
                    title: '小区业绩',
                    dataIndex: 'small_area_perf',
                },
                {
                    title: '团队业绩',
                    dataIndex: 'team_perf',
                },
                {
                    title: '直推人数',
                    dataIndex: 'invitee_count',
                },
                {
                    title: '上级地址',
                    dataIndex: 'myRecommendAddress',
                },
                {
                    title: '操作',
                    key: 'action',
                    fixed: 'right',
                    width: 220,
                    customRender: (v) => {
                        return (
                            <div>
                                <a-button-group>
                                    <a-button
                                        type="primary"
                                        onClick={() => {
                                            this.$router.push({ name: 'lookChildren', query: { userId: v.userId || v.id } })
                                        }}
                                    >
                                        查看下级
                                    </a-button>

                                    <a-dropdown>
                                        <a-button type="primary">
                                            更多
                                            <a-icon type="down" />
                                        </a-button>

                                        <a-menu slot="overlay">
                                            <a-menu-item onClick={() => this.add_account_balance(v.address)}>
                                                添加充值余额
                                            </a-menu-item>

                                            <a-menu-item onClick={() => this.recharge_to_reward(v.userId || v.id, v.address)}>
                                                充值钱包转入奖励钱包
                                            </a-menu-item>

                                            <a-menu-item onClick={() => this.vip_update(v.userId || v.id, v.vip || v.mgmt_level)}>
                                                设置级别(W0~W10)
                                            </a-menu-item>
                                        </a-menu>
                                    </a-dropdown>
                                </a-button-group>
                            </div>
                        )
                    },
                },
            ],
            searchData: {
                address: '',
            },
        }
    },
    methods: {
        getList() {
            this.loading = true
            Gai.user_list({
                page: this.current,
                ...this.searchData
            }).then((res) => {
                this.data = (res.users || []).map((value, key) => {
                    return { ...value, key }
                })
                this.loading = false
                this.total = parseInt(res.count || 0)
            }).catch(() => {
                this.loading = false
            })
        },
        vip_update(user_id, defaultValue) {
            let vip = String(defaultValue == null ? '0' : defaultValue).replace(/^W/i, '').replace(/^V/i, '')
            if (!vip) vip = '0'
            this.$confirm({
                title: `设置级别(W0~W10)`,
                content: (
                    <a-select style="width:240px" defaultValue={vip} placeholder="选择级别" onChange={(val) => {
                        vip = val;
                    }}>
                        <a-select-option value="0">W0（无级别）</a-select-option>
                        <a-select-option value="1">W1</a-select-option>
                        <a-select-option value="2">W2</a-select-option>
                        <a-select-option value="3">W3</a-select-option>
                        <a-select-option value="4">W4</a-select-option>
                        <a-select-option value="5">W5</a-select-option>
                        <a-select-option value="6">W6</a-select-option>
                        <a-select-option value="7">W7</a-select-option>
                        <a-select-option value="8">W8</a-select-option>
                        <a-select-option value="9">W9</a-select-option>
                        <a-select-option value="10">W10</a-select-option>
                    </a-select>
                ),
                centered: true,
                onOk: () => {
                    return new Promise((resolve, reject) => {
                        if (vip === undefined || vip === null || vip === '') {
                            this.$notification.warning({
                                message: '提示',
                                description: '请选择级别'
                            })
                            reject()
                            return;
                        }
                        Gai.vip_update({ user_id, vip: 'W' + vip }).then(() => {
                            this.$message.success('级别已更新')
                            resolve()
                            this.getList()
                        }).catch(() => {
                            reject()
                        })
                    })
                }
            })
        },
        add_account_balance(address) {
            let amount = ""
            this.$confirm({
                title: `添加充值余额 (usdt_recharge)`,
                content: (
                    <div>
                        <div style="margin-bottom:8px;color:#888;font-size:12px;">地址：{address}</div>
                        <a-input style="margin-top:8px;" placeholder="请输入增加的金额(USDT)" onInput={(val) => {
                            amount = val.target.value
                        }} />
                    </div>
                ),
                centered: true,
                onOk: () => {
                    const n = parseFloat(amount)
                    if (!amount || isNaN(n) || n <= 0) {
                        this.$message.warning('请填写大于 0 的金额')
                        return Promise.reject()
                    }
                    return Gai.admin_recharge({ address, amount }).then(() => {
                        this.$message.success('已添加到充值余额')
                        this.getList()
                    })
                }
            })
        },
        recharge_to_reward(user_id, address) {
            let amount = ""
            this.$confirm({
                title: `充值钱包转入奖励钱包`,
                content: (
                    <div>
                        <div style="margin-bottom:8px;color:#888;font-size:12px;">地址：{address}</div>
                        <div style="margin-bottom:8px;color:#888;font-size:12px;">仅划转 USDT，不产生直推奖；出局目标仍按 USDT 结算</div>
                        <a-input style="margin-top:8px;" placeholder="请输入划转金额(USDT)" onInput={(val) => {
                            amount = val.target.value
                        }} />
                    </div>
                ),
                centered: true,
                onOk: () => {
                    const n = parseFloat(amount)
                    if (!amount || isNaN(n) || n <= 0) {
                        this.$message.warning('请填写大于 0 的金额')
                        return Promise.reject()
                    }
                    return Gai.recharge_to_reward({ user_id, amount }).then(() => {
                        this.$message.success('已转入奖励钱包')
                        this.getList()
                    })
                }
            })
        },
    },
}
</script>

<style scoped lang="less">
.inputGroup {
    >div {
        margin-bottom: 20px;
    }
}
</style>
