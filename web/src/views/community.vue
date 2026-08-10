<template>
  <div class="community-page">
    <Header />
    <div class="container">

      <div class="my-level">
        <span class="level-label">{{ $t('community.level') }}</span>
        <span class="level-value">{{ displayLevelText }}</span>
      </div>
      <div class="info-card">
        <div class="info-box">
          <div class="info-title">{{ $t('community.superiorAddress') }}</div>
          <div class="info-address">{{ formatAddress(userinfo.inviteUserAddress) }}</div>
        </div>
        <div class="info-box">
          <div class="info-title">{{ $t('community.myInviteLink') }}</div>
          <div class="info-link">
            <span>{{ inviteUrl }}</span>
            <i class="copy-button" @click="copyToClipboard(inviteUrl)"></i>
          </div>
        </div>

        <!-- <div class="stats-grid">
          <div class="stats-item">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.evangelists') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.activeUsers') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.communitySize') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0.00 U</span>
            <span class="stats-label">{{ $t('community.totalIncome') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0 U</span>
            <span class="stats-label">{{ $t('community.communityPerformance') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.teamIdoTotal') }}</span>
          </div>
          <div class="stats-item">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.personalRedeemed') }}</span>
          </div>
          <div class="stats-item relative">
            <span class="stats-num">0</span>
            <span class="stats-label">{{ $t('community.teamRedeemed') }}</span>
          </div>
        </div> -->
      </div>

      <div class="performance-list">
        <div class="performance-info">
          <div class="performance-info-item">
            <p>{{ areaTotal }}</p>
            <p>{{ $t('community.totalPerformance') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ areaMax }}</p>
            <p>{{ $t('community.regionalPerformance') }}</p>
          </div>
          <div class="performance-info-item">
            <p>{{ areaMin }}</p>
            <p>{{ $t('community.smallPerformance') }}</p>
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
        </div>
      </div>

      <h3 class="list-title">{{ $t('community.rewardRecord') }}</h3>

      <div class="table-card">
        <div class="table-row table-header">
          <span>{{ $t('community.amount') }}</span>
          <span>{{ $t('community.generation') }}</span>
          <span>{{ $t('community.reward') }}</span>
          <span>{{ $t('community.time') }}</span>
        </div>
        <div class="income-list" v-if="rewardList.length > 0">
          <div class="table-row income-list-item" v-for="(item, index) in rewardList" :key="index">
            <span>{{ item.amount }}</span>
            <span>{{ item.num || '0' }}</span>
            <span class="col-reward">{{ item.reward }}</span>
            <span class="col-time">{{ item.createdAt }}</span>
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
import { computed, onMounted, ref } from 'vue'
import { showToast } from 'vant'
import copy from 'copy-to-clipboard'
import request from '@/tools/request'
import { Pagination } from 'vant'
import { Tree } from 'ant-design-vue'
import lang from '@/i18n/index'

const person = userPerson()
const userinfo = computed(() => person.userinfo)
const address = computed(() => person.address)

const inviteUrl = computed(() => `https://${window.location.host}/#/?inviteCode=-inviteTdh-${person.address}-inviteTdh-`)

let rewardList = $ref<any[]>([])
let page = $ref(1)
let allPageCount = $ref(1)

// performance-list 相关变量
const expandedKeys = $ref([])
const selectedKeys = $ref([])
let treeData = $ref([])
// 总业绩=各直推线之和；大区=最高线；小区=其余之和；等级按小区
const areaTotal = ref('0')
const areaMax = ref('0')
const areaMin = ref('0')

const displayLevelText = computed(() => {
  const raw = String(userinfo.value.communityLevel || userinfo.value.level || '').trim().toUpperCase()
  if (!raw || raw === '0') return '暂无'
  if (raw.startsWith('V')) return raw
  if (/^[1-9]$/.test(raw)) return `V${raw}`
  return raw
})

const fmtPerf = (n: number) => {
  if (!Number.isFinite(n) || n <= 0) return '0'
  return Number.isInteger(n) ? String(n) : n.toFixed(4).replace(/\.?0+$/, '')
}

const updateAreaStats = (recommends: any[]) => {
  const profileTotal = Number(userinfo.value.total)
  const profileMax = Number(userinfo.value.max)
  const profileMin = Number(userinfo.value.min)
  if (Number.isFinite(profileTotal) && profileTotal > 0) {
    areaTotal.value = fmtPerf(profileTotal)
    areaMax.value = fmtPerf(profileMax)
    areaMin.value = fmtPerf(profileMin)
    return
  }
  const amounts = (recommends || []).map((item) => Number(item.amount) || 0)
  const total = amounts.reduce((sum, n) => sum + n, 0)
  const max = amounts.length ? Math.max(...amounts) : 0
  const min = Math.max(0, total - max)
  areaTotal.value = fmtPerf(total)
  areaMax.value = fmtPerf(max)
  areaMin.value = fmtPerf(min)
}

const formatAddress = (value: string) => {
  if (!value) return ''
  const frontSix = value.slice(0, 6)
  const backSix = value.slice(-4)
  const middle = '...'
  return frontSix + middle + backSix
}

const copyToClipboard = (text: string) => {
  copy(text)
  showToast(lang('common.copiedToClipboard'))
}

const treeTitle = (item: any) => {
  const level = item.level ? ` | ${lang('community.level')}:${item.level}` : ''
  return `${formatAddress(item.address)}---(${lang('common.quantity')}:${item.amount})${level}`
}

const onLoadData = (treeNode: any) => {
  return new Promise<void>(async (resolve) => {
    if (treeNode.dataRef.children) {
      resolve()
      return
    }

    const res: any = await request.get(`app_server/recommend_list?address=${treeNode.dataRef.address}`)

    setTimeout(() => {
      treeNode.dataRef.children = res.recommends.map((item: any, index: number) => {
        return {
          title: treeTitle(item),
          key: `${treeNode.eventKey}-${index}`,
          amount: item.amount,
          level: item.level,
          address: item.address,
          isLeaf: false
        }
      })
      treeData = [...treeData]
      resolve()
    }, 1000)
  })
}

const getUserArea = async () => {
  const res: any = await request.get(`app_server/recommend_list?address=${address.value}`)
  const recommends = res.recommends || []
  updateAreaStats(recommends)
  treeData = recommends.map((item: any, index: number) => {
    return {
      title: treeTitle(item),
      key: index,
      amount: item.amount,
      level: item.level,
      address: item.address,
      isLeaf: item.countLow === 0
    }
  })
}

/** 我的团队奖励：走 adapter 统一汇总（静态+代数+社区） */
const getRewardList = async (pageNum: number = 1) => {
  const pageNo = Number(pageNum) > 0 ? Number(pageNum) : 1
  try {
    const res: any = await request.get('app_server/reward_list', {
      params: { page: pageNo, reqType: 'team' },
    })
    allPageCount = Math.max(1, Math.ceil((res.count || 0) / 10) || 1)
    rewardList = res.list || []
  } catch {
    rewardList = []
    allPageCount = 1
  }
}

onMounted(async () => {
  try {
    await person.getUser()
  } catch (e) {
    console.error('[community] getUser failed', e)
  }
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
  color: $brand-gold;
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
    color: $brand-gold;
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
    color: $brand-gold;
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
    border: 1px solid rgba(212, 175, 55, 0.25);
    border-radius: 6px;
    background: rgba(20, 20, 20, 0.4);
    backdrop-filter: blur(6px);

    &.relative {
      position: relative;
    }

    .stats-num {
      font-size: 18px;
      font-weight: bold;
      color: $brand-gold;
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
  background: rgba(20, 20, 20, 0.6);
  backdrop-filter: blur(10px);
  padding: 11px 0;

  .table-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1.1fr 1.6fr;
    align-items: center;
    column-gap: 4px;
    padding: 10px 12px;
    box-sizing: border-box;

    > span {
      text-align: center;
      word-break: break-all;
      line-height: 1.3;
    }
  }

  .table-header {
    background: #0A0A0A;
    margin: -11px 0 0;
    padding: 8px 12px;

    > span {
      font-size: 10px;
      color: $text-muted;
    }
  }

  .income-list {
    padding: 0;

    .income-list-item {
      border-bottom: 1px solid $border-light;

      &:last-child {
        border-bottom: none;
      }

      > span {
        font-size: 12px;
        color: $text-muted;
      }

      .col-reward {
        color: $brand-gold;
        font-size: 14px;
        font-weight: 500;
      }

      .col-time {
        font-size: 10px;
        color: $text-muted;
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
    border: 1px solid rgba(212, 175, 55, 0.2);

    .performance-info {
      display: flex;
      justify-content: space-between;
      background: rgba(0, 0, 0, 0.3);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 12px;
      margin-bottom: 20px;
      padding: 10px;

      .performance-info-item {
        height: 66px;
        flex: 1 0 0;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 10px;

        p {
          color: #fff;
          margin: 0;

          &:first-child {
            font-size: 24px;
            font-weight: 600;
            color: $brand-gold;
          }

          &:last-child {
            font-size: 12px;
            color: rgba(255, 255, 255, 0.6);
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
              background: rgba(212, 175, 55, 0.1);
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
          color: $brand-gold;
        }
      }
    }
  }

.safe-bottom {
  height: 50px;
}
</style>
