<template>
    <PageView>
        <a-spin :spinning="loading">
            <a-card title="查看下级">
                <!-- <div slot="extra" style="display:flex;align-items:center">
                    <a>伞下提现总数：{{ getAll }}</a>
                </div> -->
                <treeDownList :list="list"></treeDownList>
            </a-card>
        </a-spin>
    </PageView>
</template>

<script>
import Gai from '../../api/Gai'
export default {
    name: 'lookChildren',
    data() {
        return {
            loading: false,
            list: [],
            userId: null,
            address: undefined,
        }
    },
    activated() {
        this.list = []
        this.address = undefined;
        this.userId = this.$route.query.userId;
        this.getList()
    },
    computed:{
        getAll(){
            let all = this.list.length === 0 ? 0 : eval(this.list.map(v=>Number(v.recommendAllAmount) + Number(v.amount)).join("+"));
            return Number(all.toFixed(4));
        }
    },
    methods: {
        getList() {
            this.list = [];
            Gai.user_recommend({ userId: this.userId, address: this.address }).then(res => {
                this.list = res.users.map(v => {
                    v.child = []
                    return v
                })
            })
        },
    }
}
</script>

<style scoped lang="less"></style>