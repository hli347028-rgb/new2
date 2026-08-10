<template>
    <PageView>
        <a-card title="订单奖励">
            <a-row :gutter="10" class="inputGroup">
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-input v-model="searchData.address" placeholder="账户" @keyup.enter="getListTwo" />
                </a-col>
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-select allowClear v-model="searchData.type" style="width:100%" placeholder="类型"
                        @change="getListTwo">
                        <a-select-option v-for="item in typeOptions" :key="item.value" :value="item.value">
                            {{ item.label }}
                        </a-select-option>
                    </a-select>
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

const typeLabel = {
    static_aix: '静态奖(AIX)',
    dynamic_usdt: '直推奖(USDT)',
    mgmt: '管理奖(USDT)',
    exit_accel: '出局加速',
    transfer_in: '转入',
    transfer_out: '转出',
}

const typeOptions = Object.keys(typeLabel).map((value) => ({
    value,
    label: typeLabel[value],
}))

export default {
    name: 'ordersList',
    mixins: [listMixin],
    data() {
        return {
            typeOptions,
            columns: [
                {
                    title: '时间',
                    dataIndex: 'createdAt',
                },
                {
                    title: '类型',
                    dataIndex: 'type',
                    customRender: (v) => typeLabel[v] || v || '-',
                },
                {
                    title: '资产',
                    dataIndex: 'asset',
                },
                {
                    title: '金额',
                    dataIndex: 'amount',
                },
                {
                    title: '地址',
                    dataIndex: 'address',
                    customRender: (v) => v || '-',
                },
                {
                    title: '来源地址',
                    dataIndex: 'addressTwo',
                    customRender: (v) => v || '-',
                },
                {
                    title: '结算日',
                    dataIndex: 'settlementDate',
                    customRender: (v) => v || '-',
                },
            ],
            searchData: {
                address: '',
                type: undefined,
            },
        }
    },
    methods: {
        getList() {
            this.loading = true
            const params = {
                page: this.current,
                pageSize: this.pageSize,
                address: this.searchData.address,
            }
            if (this.searchData.type) {
                params.type = this.searchData.type
            }
            Gai.reward_list(params).then((res) => {
                const list = (res && (res.rewards || res.list)) || []
                this.data = list.map((value, key) => {
                    return { ...value, key }
                })
                this.loading = false
                this.total = parseInt(res.count || 0)
            }).catch(() => {
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
