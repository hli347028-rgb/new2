<template>
    <PageView>
        <a-card title="每日结算">
            <a-alert
                type="info"
                show-icon
                style="margin-bottom: 16px"
                message="每日结算会执行：静态奖（金本位发 AIX）+ 管理奖（W1–W10 级差，USDT 进奖励账本）。系统默认在中国时区 0 点自动跑「昨日」结算；管理端可手动触发。"
            />
            <a-row :gutter="10" class="inputGroup" style="margin-bottom: 16px">
                <a-col :xs="24" :md="8" :lg="6" :xl="5">
                    <a-date-picker
                        v-model="settleDate"
                        format="YYYY-MM-DD"
                        style="width: 100%"
                        placeholder="结算日期"
                    />
                </a-col>
                <a-col :xs="24" :md="10" :lg="8" :xl="6">
                    <a-button-group>
                        <a-button type="primary" :loading="triggering" @click="triggerSettle">执行结算</a-button>
                        <a-button :loading="loading" @click="getListTwo">刷新列表</a-button>
                    </a-button-group>
                </a-col>
            </a-row>
            <a-table
                :loading="loading"
                :columns="columns"
                :dataSource="data"
                :pagination="{ total, pageSize, current }"
                @change="changePagination"
                bordered
                :scroll="{ x: true }"
            />
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import moment from 'moment'
import Gai from '../../api/Gai'
import listMixin from '../mixin/listMixin'

export default {
    name: 'settlement',
    mixins: [listMixin],
    data() {
        return {
            settleDate: undefined,
            triggering: false,
            columns: [
                {
                    title: 'ID',
                    dataIndex: 'id',
                    width: 80,
                },
                {
                    title: '结算日期',
                    dataIndex: 'settlementDate',
                },
                {
                    title: '状态',
                    dataIndex: 'status',
                    customRender: (v) => ({
                        running: '进行中',
                        completed: '已完成',
                        success: '已完成',
                        failed: '失败',
                    }[v] || v),
                },
                {
                    title: 'AIX价格',
                    dataIndex: 'aixPrice',
                },
                {
                    title: '静态合计',
                    dataIndex: 'staticAmount',
                },
                {
                    title: '管理奖合计',
                    dataIndex: 'mgmtAmount',
                },
                {
                    title: '本轮释放合计',
                    dataIndex: 'releaseTotal',
                },
                {
                    title: '开始时间',
                    dataIndex: 'startedAt',
                },
                {
                    title: '结束时间',
                    dataIndex: 'finishedAt',
                },
                {
                    title: '操作',
                    key: 'action',
                    fixed: 'right',
                    width: 120,
                    customRender: (v) => {
                        const disabled = v.status === 'running'
                        return <a-button type="primary" disabled={disabled} onClick={() => this.rerun(v.settlementDate)}>再结算</a-button>
                    },
                },
            ],
        }
    },
    methods: {
        dateStr() {
            if (!this.settleDate) return ''
            if (moment.isMoment(this.settleDate)) {
                return this.settleDate.format('YYYY-MM-DD')
            }
            return String(this.settleDate)
        },
        getList() {
            this.loading = true
            Gai.settlement_list({
                page: this.current,
                pageSize: this.pageSize,
            }).then((res) => {
                this.data = (res.list || []).map((value, key) => {
                    return { ...value, key }
                })
                this.total = parseInt(res.total || res.count || 0)
                if (!this.settleDate && res.defaultSettleDate) {
                    this.settleDate = moment(res.defaultSettleDate, 'YYYY-MM-DD')
                }
                this.loading = false
            }).catch(() => {
                this.loading = false
            })
        },
        triggerSettle() {
            const date = this.dateStr()
            if (!date) {
                this.$message.warning('请选择结算日期')
                return
            }
            this.$confirm({
                title: '执行每日结算',
                content: `确定对 ${date} 执行每日结算吗？将发放静态 AIX 与管理奖。`,
                centered: true,
                onOk: () => {
                    this.triggering = true
                    return Gai.settlement_trigger({
                        settlement_date: date,
                    }).then(() => {
                        this.getList()
                    }).finally(() => {
                        this.triggering = false
                    })
                },
            })
        },
        rerun(date) {
            this.settleDate = date ? moment(date, 'YYYY-MM-DD') : undefined
            this.triggerSettle()
        },
    },
}
</script>

<style scoped lang="less">
.inputGroup {
    > div {
        margin-bottom: 12px;
    }
}
</style>
