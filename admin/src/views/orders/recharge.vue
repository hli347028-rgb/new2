<template>
    <PageView>
        <a-card title="充值列表">
            <a-row :gutter="10" class="inputGroup">
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-input v-model="searchData.address" placeholder="账户" @keyup.enter="getListTwo" />
                </a-col>
                <a-col :xs="12" :md="6" :lg="5" :xl="4">
                    <a-select v-model="searchData.type" allowClear placeholder="充值类型" style="width:100%" @change="getListTwo">
                        <a-select-option value="">全部类型</a-select-option>
                        <a-select-option value="admin">后台充值</a-select-option>
                        <a-select-option value="usdt">USDT充值</a-select-option>
                        <a-select-option value="win">WIN充值</a-select-option>
                    </a-select>
                </a-col>
                <a-col :xs="12" :md="10" :lg="8" :xl="6">
                    <a-button-group>
                        <a-button type="primary" :loading="loading" @click="getListTwo">确定筛选</a-button>
                        <a-button type="primary" @click="openCredit">给用户充值</a-button>
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
    name: 'recharge',
    mixins: [listMixin],
    data() {
        return {
            columns: [
                {
                    title: '账户',
                    dataIndex: 'address',
                },
                {
                    title: '充值数量',
                    dataIndex: 'amount',
                    customRender: (v, row) => {
                        const asset = (row && row.asset) || (row && row.type === 'win' ? 'WIN' : 'USDT')
                        return `${v || 0} ${asset}`
                    }
                },
                {
                    title: '类型',
                    dataIndex: 'remark',
                    customRender: (v) => {
                        if (v === '后台充值' || v === 'USDT充值' || v === 'WIN充值') return v
                        if (v === '链上充值') return 'USDT充值'
                        return v || '-'
                    }
                },
                {
                    title: '交易订单',
                    dataIndex: 'txHash',
                },
                {
                    title: '创建时间',
                    dataIndex: 'createdAt',
                },
            ],
            searchData: {
                address: '',
                type: '',
            },
        }
    },
    methods: {
        getList() {
            this.loading = true
            const params = {
                page: this.current,
                pageSize: this.pageSize,
                address: this.searchData.address || '',
            }
            if (this.searchData.type) {
                params.type = this.searchData.type
            }
            Gai.record_list(params).then((res) => {
                const list = (res && (res.rewards || res.list || res.locations)) || []
                this.data = list.map((value, key) => {
                    return { ...value, key }
                })
                this.loading = false
                this.total = parseInt(res.count || 0)
            }).catch(() => {
                this.loading = false
            })
        },
        openCredit() {
            let address = this.searchData.address || ''
            let amount = ''
            this.$confirm({
                title: '给用户充值',
                content: (
                    <div style="margin-top:16px;">
                        <div style="margin-bottom:12px;">
                            <div style="margin-bottom:6px;color:#666;">用户地址</div>
                            <a-input
                                defaultValue={address}
                                placeholder="0x..."
                                onInput={(e) => { address = e.target.value }}
                            />
                        </div>
                        <div>
                            <div style="margin-bottom:6px;color:#666;">充值金额 (USDT)</div>
                            <a-input
                                placeholder="请输入充值金额"
                                onInput={(e) => { amount = e.target.value }}
                            />
                        </div>
                        <div style="margin-top:10px;color:#999;font-size:12px;">
                            充值后进入用户 usdt_recharge，可用于报单并触发直推（类型：后台充值）
                        </div>
                    </div>
                ),
                centered: true,
                okText: '确认充值',
                onOk: () => {
                    address = (address || '').trim()
                    amount = (amount || '').trim()
                    if (!address) {
                        this.$message.warning('请填写用户地址')
                        return Promise.reject()
                    }
                    if (!amount || Number(amount) <= 0) {
                        this.$message.warning('请填写正确的充值金额')
                        return Promise.reject()
                    }
                    return Gai.admin_recharge({ address, amount }).then(() => {
                        this.getList()
                    })
                },
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
