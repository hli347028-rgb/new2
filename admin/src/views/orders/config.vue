<template>
    <PageView>
        <a-card title="AIX 配置项">
            <a-table :loading="loading" :columns="columns" :dataSource="data" :pagination="false" bordered
                :scroll="{ x: true }">
            </a-table>
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import Gai from '../../api/Gai'
import listMixin from '../mixin/listMixin'
export default {
    name: 'config',
    mixins: [listMixin],
    data() {
        return {
            columns: [
                {
                    title: '名称',
                    dataIndex: 'name',
                },
                {
                    title: '值',
                    dataIndex: 'value',
                },
                {
                    title: '操作',
                    key: 'action',
                    fixed: 'right',
                    width: 110,
                    customRender: (v) => {
                        return <a-button type="primary" onClick={() => {
                            this.config_update(v.id);
                        }}>修改</a-button>
                    },
                },
            ],
        }
    },
    methods: {
        getList() {
            this.loading = true
            Gai.config().then((res) => {
                const list = (res && res.config) ? res.config : []
                this.data = list.map((value, key) => {
                    return { ...value, key }
                })
                this.loading = false
            }).catch(() => {
                this.loading = false
            })
        },
        config_update(id) {
            const row = (this.data || []).find((item) => item.id === id)
            let value = row && row.value != null ? String(row.value) : ""
            this.$confirm({
                title: `修改${row && row.name ? ` - ${row.name}` : ''}`,
                content: (
                    <div>
                        <div style="margin-bottom:8px;color:#888;font-size:12px;">
                            静态利率填百分数如 0.5；直推/W 收益系数填小数如 0.2 表示 20%；出局倍数默认 4
                        </div>
                        <a-input style="margin-top:8px;" defaultValue={value} placeholder="请输入" onInput={(val) => {
                            value = val.target.value
                        }} />
                    </div>
                ),
                centered: true,
                onOk: () => {
                    return new Promise((resolve, reject) => {
                        Gai.config_update({ id, value }).then(() => {
                            resolve()
                            this.getList()
                        }).catch(() => {
                            reject()
                        })
                    })
                }
            })
        }
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
