<template>
    <PageView>
        <a-card title="AIX→WIN 兑换记录">
            <a-row :gutter="10" class="inputGroup">
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-input v-model="searchData.address" placeholder="钱包地址" @keyup.enter="getListTwo" />
                </a-col>
                <a-col :xs="12" :md="6" :lg="6" :xl="4">
                    <a-button-group>
                        <a-button type="primary" :loading="loading" @click="getListTwo">确定筛选</a-button>
                    </a-button-group>
                </a-col>
            </a-row>
            <a-table :loading="loading" :columns="columns" :dataSource="data"
                :pagination="{ total, pageSize, showSizeChanger, current }" @change="changePagination" bordered
                :scroll="{ x: true }">
            </a-table>
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import Gai from '../../api/Gai'
import listMixin from '../mixin/listMixin'

export default {
    name: 'exchangeList',
    mixins: [listMixin],
    data() {
        return {
            columns: [
                {
                    title: 'ID',
                    dataIndex: 'id',
                },
                {
                    title: '地址',
                    dataIndex: 'address',
                },
                {
                    title: '兑换来源',
                    dataIndex: 'fromAsset',
                    customRender: (v) => v || 'AIX'
                },
                {
                    title: '支付数量',
                    dataIndex: 'fromAmount',
                },
                {
                    title: '获得币种',
                    dataIndex: 'toAsset',
                    customRender: (v) => v || 'WIN'
                },
                {
                    title: '到账数量',
                    dataIndex: 'toAmount',
                },
                {
                    title: '手续费',
                    dataIndex: 'feeAmount',
                    customRender: (v) => v || '0'
                },
                {
                    title: '手续费率',
                    dataIndex: 'feeRate',
                    customRender: (v) => {
                        if (!v) return '-'
                        try {
                            const pct = parseFloat(v) * 100
                            return isNaN(pct) ? v : pct.toFixed(2) + '%'
                        } catch (e) { return v }
                    }
                },
                {
                    title: '兑换价格',
                    dataIndex: 'exchangePrice',
                },
                {
                    title: '状态',
                    dataIndex: 'status',
                    customRender: (v) => {
                        if (v === 1 || v === '1') return <a-tag color="green">成功</a-tag>
                        if (v === 0 || v === '0') return <a-tag color="orange">处理中</a-tag>
                        return <a-tag color="red">{v}</a-tag>
                    }
                },
                {
                    title: '备注',
                    dataIndex: 'remark',
                    customRender: (v) => v || '-'
                },
                {
                    title: '时间',
                    dataIndex: 'createdAt',
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
            Gai.exchange_list({
                page: this.current,
                ...this.searchData
            }).then((res) => {
                this.data = (res.list || []).map((value, key) => {
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
