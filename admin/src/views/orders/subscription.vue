<template>
    <PageView>
        <a-card :title="`订单列表（共 ${total} 条）`">
            <a-row :gutter="10" class="inputGroup">
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-input v-model="searchData.address" placeholder="用户地址" allowClear @keyup.enter="getListTwo" />
                </a-col>
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-select allowClear v-model="searchData.status" style="width:100%" placeholder="状态"
                        @change="getListTwo">
                        <a-select-option value="active">进行中</a-select-option>
                        <a-select-option value="exited">已出局</a-select-option>
                    </a-select>
                </a-col>
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-button-group>
                        <a-button type="primary" :loading="loading" @click="getListTwo">确定筛选</a-button>
                    </a-button-group>
                </a-col>
            </a-row>
            <a-table
                rowKey="id"
                :loading="loading"
                :columns="columns"
                :dataSource="data"
                :pagination="{ total, pageSize, current, showSizeChanger: true, pageSizeOptions: ['20', '50', '100', '200'] }"
                @change="changePagination"
                bordered
                :scroll="{ x: true }"
            />
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import Gai from '../../api/Gai'
import listMixin from '../mixin/listMixin'

const statusText = {
    active: '进行中',
    exited: '已出局',
    completed: '已出局',
}

const fundSourceText = {
    recharge: '充值账本',
    reward: '奖励账本',
}

export default {
    name: 'subscription',
    mixins: [listMixin],
    data() {
        return {
            columns: [
                {
                    title: '时间',
                    dataIndex: 'createdAt',
                },
                {
                    title: '报单本金',
                    dataIndex: 'amount',
                },
                {
                    title: '出局倍数',
                    dataIndex: 'exitAmount',
                },
                {
                    title: '出局目标',
                    dataIndex: 'money',
                },
                {
                    title: '已获收益',
                    dataIndex: 'amountGet',
                },
                {
                    title: '剩余额度',
                    dataIndex: 'amountLast',
                },
                {
                    title: '资金来源',
                    dataIndex: 'fund_source',
                    customRender: (v) => fundSourceText[v] || v || '-',
                },
                {
                    title: '用户地址',
                    dataIndex: 'address',
                },
                {
                    title: '状态',
                    dataIndex: 'status',
                    customRender: (v) => statusText[v] || v || '-',
                },
            ],
            searchData: {
                address: '',
                status: undefined,
            },
            pageSize: 50,
        }
    },
    mounted() {
        this.getList()
    },
    methods: {
        getList() {
            this.loading = true
            const params = {
                page: this.current || 1,
                pageSize: this.pageSize || 50,
            }
            const address = (this.searchData.address || '').trim()
            if (address) params.address = address
            const status = this.searchData.status
            if (status) params.status = status

            Gai.buy_list(params).then((res) => {
                const list = (res && res.rewards) ? res.rewards : []
                this.data = list.map((value, key) => ({
                    ...value,
                    id: value.id != null ? value.id : key,
                    key: value.id != null ? value.id : key,
                }))
                this.total = parseInt((res && res.count) || 0, 10) || 0
            }).catch(() => {
                this.data = []
                this.total = 0
            }).finally(() => {
                this.loading = false
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
