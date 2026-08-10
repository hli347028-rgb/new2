<template>
    <PageView>
        <a-card class="cardCon" title="AIX 数据统计" :loading="loading">
            <a-row :gutter="0">
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="team" />
                        <div>
                            <a>注册人数:</a> {{ data.totalUserR || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="team" />
                        <div>
                            <a>报单人数:</a> {{ data.totalUser || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="team" />
                        <div>
                            <a>今日注册:</a> {{ data.todayUserR || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="team" />
                        <div>
                            <a>今日报单人数:</a> {{ data.todayUser || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="pay-circle" />
                        <div>
                            <a>报单总额(U):</a> {{ data.buyTotal || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="pay-circle" />
                        <div>
                            <a>今日报单(U):</a> {{ data.todayBuy || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="wallet" />
                        <div>
                            <a>USDT奖励合计:</a> {{ data.balanceUsdt || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="wallet" />
                        <div>
                            <a>USDT充值合计:</a> {{ data.usdtRechargeTotal || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="fund" />
                        <div>
                            <a>AIX余额合计:</a> {{ data.aixTotal || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="ordered-list" />
                        <div>
                            <a>活跃订单:</a> {{ data.orderActive || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="check-circle" />
                        <div>
                            <a>已出局订单:</a> {{ data.orderExited || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="rise" />
                        <div>
                            <a>今日静态(AIX):</a> {{ data.todayOne || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="rise" />
                        <div>
                            <a>今日动态(U):</a> {{ data.todayTwo || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="rise" />
                        <div>
                            <a>今日动静态:</a> {{ data.todayThree || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
                <a-col :xs="12" :md="8" :lg="6" :xl="6">
                    <a-card-grid>
                        <a-icon type="bar-chart" />
                        <div>
                            <a>奖励累计:</a> {{ data.totalReward || 0 }}
                        </div>
                    </a-card-grid>
                </a-col>
            </a-row>
        </a-card>
    </PageView>
</template>

<script type="text/jsx">
import Gai from '../../api/Gai'
export default {
    name: 'home',
    data() {
        return {
            loading: true,
            data: {},
        }
    },
    activated() {
        this.getList()
    },
    methods: {
        getList() {
            this.loading = true
            Gai.all().then(res => {
                this.data = res
                this.loading = false
            })
        },
    }
}
</script>

<style scoped lang="less">
.cardCon {
    /deep/ .ant-card-body {
        padding: 0 !important;
    }

    .ant-card-grid {
        padding: 24px 15px;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: space-between;

        i {
            font-size: 20px;
            color: #ffffff;
            padding: 10px;
            border-radius: 8px;
            background: #1890ff;
        }
    }
}

.inputGroup {
    >div {
        margin: 20px;
    }
}
</style>
