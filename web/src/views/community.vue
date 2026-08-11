<template>
  <div class="community-page">
    <Header />
    <div class="container">

      <div class="my-level">
        <span class="level-label">{{ $t('community.communityLevel') }}</span>
        <span class="level-value">{{ levelLabel }}</span>
      </div>
      <div class="info-card">
        <div class="info-box">
          <div class="info-title">{{ $t('community.superiorAddress') }}</div>
          <div class="info-address">{{ formatAddress(userinfo.inviteUserAddress) || '-' }}</div>
        </div>
        <div class="info-box">
          <div class="info-title">{{ $t('community.myInviteLink') }}</div>
          <div class="info-link">
            <span>{{ inviteUrl }}</span>
            <i class="copy-button" @click="copyToClipboard(inviteUrl)"></i>
          </div>
        </div>
      </div>

      <div class="performance-list">
        <div class="performance-info">
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.recommendNum) }}</p>
            <p>{{ $t('community.directReferralCount') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.buy) }}</p>
            <p>{{ $t('community.activeSubscription') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.total) }}</p>
            <p>{{ $t('community.teamTotalPerformance') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.max) }}</p>
            <p>{{ $t('community.regionalPerformance') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.min) }}</p>
            <p>{{ $t('community.smallAreaPerformance') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.recommend) }}</p>
            <p>{{ $t('community.directReferralReward') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.location) }}</p>
            <p>{{ $t('community.staticIncomeTotal') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.team) }}</p>
            <p>{{ $t('community.managementReward') }}</p>
          </div>
          <!-- 
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.teamTwo) }}</p>
            <p>{{ $t('community.peerDividendTotal') }}</p>
          </div> -->
          <div class="performance-info-item">
            <p>{{ formatNum(userinfo.all) }}</p>
            <p>{{ $t('community.incomeTotal') }}</p>
          </div>
        </div>
        <div class="performance-share-title">{{ $t('community.directInviteData') }}</div>
        <div class="performance-share-list">
          <Tree
            v-if="treeData.length > 0"
            v-model:expandedKeys="expandedKeys"
            v-model:selectedKeys="selectedKeys"
            :load-data="onLoadData"
            :tree-data="treeData"
          />
          <div v-else class="empty-state"><p>{{ $t('common.noData') }}</p></div>
        </div>
      </div>

      <h3 class="list-title">{{ $t('community.generationRewardRecord') }}</h3>

      <div class="table-card">
        <div class="table-header">
          <span>{{ $t('community.amount') }}</span>
          <span>{{ $t('community.source') }}</span>
          <span>{{ $t('community.time') }}</span>
        </div>
        <div class="income-list" v-if="rewardList.length > 0">
          <div class="income-list-item" v-for="(item, index) in rewardList" :key="index">
            <div class="income-list-item-info">
              <p>{{ item.reward }} U<span v-if="item.num"> · {{ $t('community.generationNum', { num: item.num }) }}</span></p>
              <p>{{ item.createdAt }}</p>
            </div>
            <div class="income-list-item-money">{{ formatAddr(item.address) }}</div>
          </div>
          <Pagination
            v-model="page"
            :page-count="allPageCount"
            mode="simple"
            @change="getRewardList"
          />
        </div>
        <div class="empty-state" v-else>
          <p>{{ $t('common.noData') }}</p>
        </div>
      </div>

      <div class="safe-bottom"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import { computed, onMounted } from 'vue'
import { showToast } from 'vant'
import copy from 'copy-to-clipboard'
import request from '@/tools/request'
import { Pagination } from 'vant'
import { Tree } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'

const person = userPerson()
const { t: $t } = useI18n()
const userinfo = computed(() => person.userinfo)
const address = computed(() => person.address)

// 使用 ?code= 便于本地登录弹窗预填（与 eth_authorize 邀请码一致）
const inviteUrl = computed(() => {
  const addr = person.address || ''
  if (!addr) return ''
  return `${window.location.origin}/?code=${addr}`
})

const levelLabel = computed(() => {
  const lv = Number(userinfo.value?.level || 0)
  return lv > 0 ? `V${lv}` : 'V0'
})

let rewardList = $ref<any[]>([])
let page = $ref(1)
let allPageCount = $ref(1)
let active = $ref('3') // reqType=3 代数奖励

// performance-list 相关变量
const expandedKeys = $ref([])
const selectedKeys = $ref([])
let treeData = $ref([])

const formatAddress = (value: string) => {
  if (!value) return ''
  const frontSix = value.slice(0, 6)
  const backSix = value.slice(-4)
  const middle = '...'
  return frontSix + middle + backSix
}

const formatAddr = (value: string) => formatAddress(value) || '-'

const formatNum = (value: any) => Number(value || 0).toFixed(2)

const copyToClipboard = (text: string) => {
  copy(text)
  showToast($t('common.copiedToClipboard'))
}

const onLoadData = (treeNode: any) => {
  return new Promise<void>(async (resolve) => {
    if (treeNode.dataRef.children) {
      resolve()
      return
    }

    const res: any = await request.get(`app_server/recommend_list?address=${treeNode.dataRef.address}`)

    treeNode.dataRef.children = (res.recommends || []).map((item: any, index: number) => {
      const hasChildren = item.hasChildren != null ? !!item.hasChildren : Number(item.countLow || 0) > 0
      const tag = item.activated === false ? $t('community.inactive') : `${$t('community.subscribe')}: ${item.amount}`
      return {
        title: `${formatAddress(item.address)}（${tag}）`,
        key: `${treeNode.eventKey}-${index}`,
        amount: item.amount,
        address: item.address,
        isLeaf: !hasChildren
      }
    })
    treeData = [...treeData]
    resolve()
  })
}

const getUserArea = async () => {
  if (!address.value) return
  const res: any = await request.get(`app_server/recommend_list?address=${address.value}`)
  treeData = (res.recommends || []).map((item: any, index: number) => {
    const hasChildren = item.hasChildren != null ? !!item.hasChildren : Number(item.countLow || 0) > 0
    const tag = item.activated === false ? $t('community.inactive') : `${$t('community.subscribe')}: ${item.amount}`
    return {
      title: `${formatAddress(item.address)}（${tag}）`,
      key: index,
      address: item.address,
      isLeaf: !hasChildren
    }
  })
}

const getRewardList = async (pageNum: number = 1) => {
  const res: any = await request.get("app_server/reward_list", {
    params: {
      page: pageNum,
      reqType: active
    }
  })

  allPageCount = Math.ceil((res.count || 0) / 10) || 1
  rewardList = res.list || []
}

onMounted(() => {
  person.getUser?.()
  getRewardList()
  getUserArea()
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.community-page {
  min-height: 100vh;
  padding-top: 64px;
}

.container {
  padding: 0 15px;
}

.my-level {
  margin-top: 10px;
  color: $brand-primary;
  font-size: 18px;
  display: flex;
  align-items: center;
  gap: 10px;

  .level-label {
    color: rgba(255, 255, 255, 0.6);
    font-size: 16px;
    font-weight: 400;
  }

  .level-value {
    color: $brand-primary;
    font-size: 20px;
    font-weight: 600;
  }
}

.info-card {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 13px;
}

.info-box {
  width: 100%;
  min-height: 80px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 18px;
  padding: 15px 15px 20px 15px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 10px;

  .info-title {
    color: $brand-primary;
    font-size: 16px;
  }

  .info-address {
    word-break: break-all;
    color: $text-primary;
    font-size: 14px;
  }

  .info-link {
    display: flex;
    gap: 20px;
    align-items: center;

    span {
      word-break: break-all;
      color: $text-primary;
      font-size: 14px;
    }

    .copy-button {
      flex-shrink: 0;
      width: 32px;
      height: 32px;
      background: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAC6UlEQVR4AexXTWgTQRR+M7tbYpvQWlLw0EMOCh4VLJpGIUcPHhSTEEWQoIWIesjfPVdr2ngpVSz0UolNCz2IUPBgwGztQdCD4MGD3lQaMNJAI4k7vllYTLK7ye5mBQ8O82bn571vvryZeTOh0Ce9zIfF9UIoWV4IfUBhDmVvoxBaLs/PHDGaypQAn3zP194iBJbR8DiK0+xnBJIgjLzeLAR1OKYEvo217gCwC05nNbALKCC8WLt3erpzzJQAJXC9U9GVOmHTI4JQ7MQyJRDLyidjGZk4FQUgSABWOidT64RESouhU2odC1MCODZUjmfk3WhGniMUrvUCiQA3tD6VQHkp7F1fCD3GXX6A4nS3a3b75cLshrbW0ZT8BD2xpk3Iv4yx8/zLha7mwx742XqGSjexw4MybPYCunlEFF9pJH4BLHWDkgA/ZbyPHhprJYFBGNxPAUmQ7nPYg0npHf7ANq9rUh8HL69TSv+sB+9wUwhhEe7hRKLSBCANMEh8D6hMDMbc6BInxqHvsnICbkzkGOM/ASMPbLeZNGU3AoJH8hEGD+2uhY6AorDc1WylZhcodrvS8Deku2iHOx5Li5n26lEGGCl7e621P4N9W6qDFkmxvBg8ulUMT9gR/uAY9bZXEa/vscPxrqwnwKMiox9bSuu7HcEHxxcMPPEudAsNPQELRm6q/JsE+MVBCKnbEoCuy8aql4w8sC1SaSqarh62I41JyUcAHoDNpCNAmZK6lKrUbeJAAm88/76UQ7vh4gBQ6vh2dCUOKAyK5fnZE3ZiANctFc8ERn3tEnrAShz4qnlZtwQIcBYE8tZODOC6giJ8IsAuov2g3BQo488/Vc+IgDrgfsFq+FbcxP8b5y6ndp5r+NiGmtb4G9/6D1A3JT7Rj8XS1WgkLb/pnIfig/RRZ4fL9d1Enr8HzVFpNCuvEEaemqs4Hmng1T43yFrdA9Fs9QpjcAuV8SRhOVxu8rVuM2Umntt5PwjqNwAAAP//ec0etAAAAAZJREFUAwA1x8ykU4MciwAAAABJRU5ErkJggg==') no-repeat;
      background-size: 20px 20px;
      cursor: pointer;
      transition: opacity 0.3s ease;

      &:hover {
        opacity: 0.8;
      }
    }
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  margin-top: 15px;

  .stats-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 10px 0;
    border: 1px solid rgba(21, 151, 229, 0.25);
    border-radius: 6px;
    background: rgba(8, 19, 30, 0.4);
    backdrop-filter: blur(6px);

    &.relative {
      position: relative;
    }

    .stats-num {
      font-size: 18px;
      font-weight: bold;
      color: $brand-primary;
    }

    .stats-label {
      margin-top: 7px;
      font-size: 13px;
      color: #fff;
      text-align: center;
    }
  }
}

.list-title {
  margin: 30px 0 0 0;
  font-size: 15px;
  color: $text-muted;
  text-align: center;
  font-weight: normal;
}

.table-card {
  margin-top: 15px;
  min-height: 300px;
  overflow: hidden;
  border: 1px solid $border-color;
  border-radius: 11px;
  background: rgba(8, 19, 30, 0.6);
  backdrop-filter: blur(10px);
  padding: 11px 0;

  .table-header {
    display: flex;
    align-items: center;
    background: #030A11;
    padding: 8px 0;
    margin: -11px 0 0;

    span {
      flex: 1;
      text-align: center;
      font-size: 10px;
      color: $text-muted;
    }
  }

  .income-list {
    padding: 10px;

    .income-list-item {
      width: 100%;
      box-sizing: border-box;
      padding: 10px;
      border-bottom: 1px solid $border-light;
      display: flex;
      align-items: center;

      &:last-child {
        border-bottom: none;
      }

      .income-list-item-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 4px;

        p {
          color: $text-muted;
          font-size: 12px;

          &:last-child {
            color: $text-muted;
            font-size: 10px;
          }
        }
      }

      .income-list-item-money {
        flex-shrink: 0;
        width: 80px;
        text-align: right;
        color: $brand-primary;
        font-size: 14px;
        font-weight: 500;
      }
    }
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 250px;

    p {
      margin-top: 8px;
      font-size: 12px;
      color: $text-muted;
    }
  }
}

 .performance-list {
    margin-top: 20px;
    min-height: 200px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 18px;
    padding: 16px;
    box-sizing: border-box;
    margin-bottom: 25px;
    border: 1px solid rgba(21, 151, 229, 0.2);

    .performance-info {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 10px;
      margin-bottom: 20px;

      .performance-info-item {
        min-height: 66px;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 12px 8px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        box-sizing: border-box;

        p {
          color: #fff;
          margin: 0;
          text-align: center;

          &:first-child {
            font-size: 18px;
            font-weight: 600;
            color: $brand-primary;
            word-break: break-all;
          }

          &:last-child {
            font-size: 12px;
            color: rgba(255, 255, 255, 0.8);
          }
        }
      }
    }

    .performance-share-title {
      margin-bottom: 15px;
      color: #fff;
      font-size: 14px;
      font-weight: 500;
    }

    .performance-share-list {
      // width: 100%;
      display: flex;
      min-height: 100px;
      background: rgba(0, 0, 0, 0.2);
      border-radius: 12px;
      padding: 10px;
      overflow-x: auto;

      :deep(.ant-tree) {
        background: transparent;
        color: #fff;
        width: 100%;

        .ant-tree-treenode {
          padding: 4px 0;

          .ant-tree-node-content-wrapper {
            color: #fff;
            &:hover {
              background: rgba(21, 151, 229, 0.1);
            }
          }

          .ant-tree-switcher {
            color: #fff;
          }

          .ant-tree-node-title {
            color: #fff;
          }
        }

        .ant-tree-switcher_open,
        .ant-tree-switcher_close {
          color: $brand-primary;
        }
      }
    }
  }

.safe-bottom {
  height: 50px;
}
</style>
