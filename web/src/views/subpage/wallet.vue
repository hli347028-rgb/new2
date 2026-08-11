<template>
<div class='page'>
  <van-nav-bar
    :title="$t('wallet.title')"
    left-arrow
    :border="false"
    fixed
    @click-left="handleBack"
  />
  <div class="page-main">
    <div class="usdt-price" @click="router.push('/withdrawal')">
      <div class="price-list">
        <div class="price-item">
          <p>{{ $t('wallet.rechargeBalance') }}</p>
          <p>{{ userinfo.usdt || 0 }}</p>
        </div>
        <div class="price-item">
          <p>{{ $t('wallet.rewardBalance') }}</p>
          <p>{{ userinfo.reward || 0 }}</p>
        </div>
        <div class="price-item">
          <p>{{ $t('wallet.withdrawableAix') }}</p>
          <p>{{ userinfo.amountGet || 0 }}<van-icon style="margin-left: 5px;" name="arrow" /></p>
        </div>
      </div>
    </div>
    <ul class="wallet-tab">
      <li :class="tab === 1 ? 'active' : ''" @click="tab = 1">{{ $t('wallet.myIncome') }}</li>
      <li :class="tab === 2 ? 'active' : ''" @click="tab = 2">{{ $t('wallet.referralRelation') }}</li>
    </ul>
    <div class="tab-content" v-if="tab === 1">
      <div class="pledge">
        <div class="pledge-info">
          <div class="pledge-item">
            <p>{{ $t('wallet.exitRemaining') }}</p>
            <p>{{ userinfo.amountGetSub || 0 }}</p>
          </div>
          <div class="pledge-item">
            <p>{{ $t('wallet.earnedIncome') }}</p>
            <p>{{ userinfo.all || 0 }}</p>
          </div>
        </div>
      </div>
      <div class="pledge-frame">
        <div class="pledge-frame-item">
          <p>{{ $t('wallet.staticIncome') }}</p>
          <p>{{ userinfo.location || 0 }}</p>
        </div>
        <div class="pledge-frame-item">
          <p>{{ $t('wallet.directReferralReward') }}</p>
          <p>{{ userinfo.recommend || 0 }}</p>
        </div>
        <div class="pledge-frame-item">
          <p>{{ $t('wallet.managementReward') }}</p>
          <p>{{ userinfo.team || 0 }}</p>
        </div>
        <div class="pledge-frame-item">
          <p>{{ $t('wallet.totalIncome') }}</p>
          <p>{{ userinfo.all || 0 }}</p>
        </div>
      </div>
      <van-tabs v-model:active="active" scrollable :ellipsis="false" @change="onChangeTab">
        <van-tab v-for="value in menuType" :title="value[1]" :name="value[0]" :key="value[0]" />
      </van-tabs>
      <div class="records-panel" :aria-busy="rewardLoading">
        <div v-if="rewardLoading" class="records-loading">
          <van-loading type="spinner" color="#1597E5" />
        </div>
        <van-empty v-else-if="rewardList.length === 0" :description="active === '1' ? $t('wallet.noSubscribeRecords') : $t('wallet.noIncomeOfType')" :image="emptyImage" />
        <div class="income-list" v-else>
          <div class="income-list-main">
            <div class="income-list-item" v-for="(item, index) in rewardList" :key="item.id || `${active}-${page}-${index}`">
              <div class="income-list-item-info">
                <p>
                  <template v-if="active === '1'">
                    {{ item.name || $t('community.subscribe') }}
                    <span v-if="item.exited != null" style="margin-left:6px;opacity:.7">
                      {{ item.exited ? $t('node.statusExited') : $t('node.statusActive') }}
                    </span>
                  </template>
                  <template v-else>
                    {{ item.name || $t('wallet.income') }}
                    <span v-if="active === '3' && item.num" style="margin-left:6px;opacity:.7">
                      {{ $t('community.generationNum', { num: item.num }) }}
                    </span>
                  </template>
                </p>
                <p class="income-list-item-time">
                  <span>{{ item.createdAt }}</span>
                  <span class="income-list-item-money">{{ item.reward }}</span>
                </p>
                <p v-if="active === '1' && item.progressAcc != null && item.progressTarget" style="font-size: 12px; opacity: .7;">
                  {{ $t('wallet.exitProgress') }} {{ item.progressAcc }} / {{ item.progressTarget }}
                </p>
                <p v-if="item.address" style="font-size: 12px; opacity: .7;">{{ formatShortAddr(item.address) }}</p>
              </div>
            </div>
            <Pagination
              v-if="allPageCount > 1"
              v-model="page"
              :page-count="allPageCount"
              mode="simple"
              @change="getRewardList"
            />
          </div>
        </div>
      </div>
    </div>
    <div class="tab-content" v-if="tab === 2">
      <van-empty v-if="treeData.length === 0" :description="$t('wallet.noDirectReferral')" :image="emptyImage" />
      <a-tree
        v-else
        v-model:expandedKeys="expandedKeys"
        v-model:selectedKeys="selectedKeys"
        :load-data="onLoadData"
        :tree-data="treeData"
      />
    </div>
  </div>
</div>
</template>
<script setup>
import userPerson from "@/pinia/person";
import { useRouter } from 'vue-router'
import { onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import emptyImage from '../../assets/images/custom-empty-image.png'
import request from "@/tools/request";
import { Pagination } from "vant"

const { t: $t, locale } = useI18n()
const router = useRouter()
const person = userPerson();
const userinfo = $computed(() => person.userinfo);
const address = $computed(() => person.address);

let active = $ref('1')
let page = $ref(1);
let allPageCount = $ref(1);
let rewardList = $ref([]);
let rewardLoading = $ref(false)
let rewardRequestId = 0
const tab = $ref(1)

const menuType = computed(() => [
  ['1', $t('wallet.subscribeRecords')],
  ['2', $t('wallet.staticIncome')],
  ['3', $t('wallet.directReferralReward')],
  ['5', $t('wallet.managementReward')],
])
const expandedKeys = $ref([]);
const selectedKeys = $ref([]);
let treeData = $ref([]);

const formatShortAddr = (value) => {
  if (!value) return ''
  return `${value.slice(0, 6)}...${value.slice(-4)}`
}

const getRewardList = async (pageNum = 1, reqType = active) => {
  const requestId = ++rewardRequestId
  const requestedType = String(reqType)
  rewardLoading = true
  try {
    if (requestedType === '1') {
      const res = await request.get("app_server/order_list", {
        params: { page: pageNum }
      });
      if (requestId !== rewardRequestId || requestedType !== active) return
      const count = Number(res?.count || 0)
      allPageCount = Math.max(1, Math.ceil(count / 10));
      rewardList = (res?.list || []).map((item) => {
        const exited = String(item.status) === '2'
        const acc = item.accumulated ?? item.amountGet ?? '0'
        const target = item.exit_target ?? item.amountMax ?? ''
        return {
          name: $t('community.subscribe'),
          exited,
          createdAt: item.createdAt || item.created_at || '',
          reward: item.amount || '0',
          progressAcc: target ? acc : null,
          progressTarget: target || '',
          address: person.address || address || '',
        }
      })
      page = pageNum
      return
    }
    const res = await request.get("app_server/reward_list", {
      params: {
        page: pageNum,
        reqType: requestedType
      }
    });
    if (requestId !== rewardRequestId || requestedType !== active) return
    const count = Number(res?.count || 0)
    allPageCount = Math.max(1, Math.ceil(count / 10));
    rewardList = res?.list || []
    page = pageNum
  } catch {
    if (requestId !== rewardRequestId || requestedType !== active) return
    rewardList = []
    allPageCount = 1
  } finally {
    if (requestId === rewardRequestId) rewardLoading = false
  }
}

const formatAddress = (value) => {
  if (!value) return ''
  const frontSix = value.slice(0, 6);
  const backSix = value.slice(-4);
  const middle = '...';
  return frontSix + middle + backSix;
}

const buildTreeTag = (item) => {
  return item.activated === false
    ? $t('community.inactive')
    : `${$t('community.subscribe')}:${item.amount}`
}

const onLoadData = (treeNode) => {
  return new Promise(async (resolve) => {
    if (treeNode.dataRef.children) {
      resolve();
      return;
    }

    try {
      const res = await request.get(`app_server/recommend_list?address=${treeNode.dataRef.address}`);
      treeNode.dataRef.children = (res.recommends || []).map((item, index) => {
        const hasChildren = item.hasChildren != null ? !!item.hasChildren : Number(item.countLow || 0) > 0
        const tag = buildTreeTag(item)
        return {
          title: `${formatAddress(item.address)}（${tag}）`,
          key: `${treeNode.eventKey}-${index}`,
          amount: item.amount,
          address: item.address,
          isLeaf: !hasChildren
        }
      })
      treeData = [...treeData];
    } catch {
      treeNode.dataRef.children = []
    }
    resolve();
  });
};

const getUserArea = async () => {
  const addr = person.address || address
  if (!addr) {
    treeData = []
    return
  }
  try {
    const res = await request.get(`app_server/recommend_list?address=${addr}`);
    treeData = (res.recommends || []).map((item, index) => {
      const hasChildren = item.hasChildren != null ? !!item.hasChildren : Number(item.countLow || 0) > 0
      const tag = buildTreeTag(item)
      return {
        title: `${formatAddress(item.address)}（${tag}）`,
        key: index,
        address: item.address,
        isLeaf: !hasChildren
      }
    })
  } catch {
    treeData = []
  }
}

watch(locale, () => {
  getRewardList(page, active)
  getUserArea()
})

onMounted(async () => {
  await person.getUser?.()
  getRewardList(1)
  getUserArea()
})

const onChangeTab = (name) => {
  active = String(name)
  page = 1
  allPageCount = 1
  rewardList = []
  getRewardList(1, active)
}

const handleBack = () => {
  router.back()
}

</script>
<style lang='less' scoped>
  .page {
    min-height: 100vh;
    box-sizing: border-box;
    padding: 50px 15px 20px 15px;
    background: url('../../assets/images/a3.png') no-repeat;
    background-size: 100% auto;
    .page-main {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    .usdt-price {
      width: 100%;
      height: 111px;
      background-image: url(@/assets/images/a1.png);
      background-repeat: no-repeat;
      background-size: 100% 111px;
      padding: 20px 20px 0 20px;
      display: flex;
      box-sizing: border-box;
      font-size: 15px;
      .price-list {
        width: 60%;
        display: flex;
        flex: 1;
        flex-direction: column;
        justify-content: center;
        gap: 6px;

        .price-item {
          display: flex;
          flex-direction: row;
          align-items: center;
          gap: 10px;

          p {
            margin: 0;
          }
        }
      }
    }
    .wallet-tab {
      width: 100%;
      min-height: 45px;
      background: #0B1824;
      border: 1px solid #183247;
      border-radius: 31px;
      padding: 5px;
      box-sizing: border-box;
      display: flex;
      li {
        flex: 1 0 0;
        display: flex;
        align-items: center;
        justify-content: center;
        &.active {
          height: 45px;
          background: #087BC1;
          border-radius: 26px;
          font-weight: 500;
        }
      }
    }
    .ispay-price {
      width: 100%;
      height: 111px;
      background-image: url(@/assets/images/a2.png);
      background-repeat: no-repeat;
      background-size: 100% 111px;
      padding: 25px 20px;
      box-sizing: border-box;
      display: flex;
      flex-direction: column;
      gap: 20px;
      font-size: 18px;
    }
    .tab-menu {
      display: flex;
      height: 55px;
      background: #0B1824;
      border: 1px solid #183247;
      border-radius: 31px;
      padding: 5px;
      box-sizing: border-box;
      .tab-item {
        flex: 1;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .active-tab {
        background: #087BC1;
        border-radius: 32px;
        font-weight: 500;
      }
    }
    .tab-content {
      display: flex;
      flex-direction: column;
      gap: 20px;
      :deep(.van-tabs__wrap) {
        overflow: visible;
      }
      :deep(.van-tabs__nav) {
        overflow-x: auto;
        overflow-y: hidden;
        -webkit-overflow-scrolling: touch;
      }
      :deep(.van-tab) {
        flex: none;
        padding: 0 14px;
        white-space: nowrap;
      }
      :deep(.van-tab__text) {
        overflow: visible;
        white-space: nowrap;
      }
      :deep(.van-tabs__content) {
        display: none;
      }
      .records-panel {
        height: clamp(260px, 45vh, 420px);
        position: relative;
        overflow-y: auto;
        overscroll-behavior: contain;
        -webkit-overflow-scrolling: touch;
      }
      .records-loading {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .records-panel > :deep(.van-empty) {
        min-height: 100%;
        box-sizing: border-box;
      }
      .pledge {
        background: #0B1824;
        border-radius: 18px;
        overflow: hidden;
        .pledge-total {
          display: flex;
          justify-content: space-between;
          margin: 15px;
          height: 57px;
          background: hsla(0, 0%, 100%, .1);
          border-radius: 12px;
          padding: 0 15px;
          box-sizing: border-box;
          align-items: center;
          span {
            &:nth-child(2) {
              color: rgb(21, 151, 229);
              font-weight: 500;
            }
          }
        }
        .pledge-info {
          display: flex;
          margin: 20px 15px;
          .pledge-item {
            flex: 1;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 6px;
            p {
              &:nth-child(2) {
                font-weight: 500;
                font-size: 18px;
              }
            }
          }
        }
        .pledge-count {
          border-top: 1px solid rgba(255, 255, 255, 0.1);
          background: url('@/assets/images/xian.png') no-repeat;
          background-size: 100% auto;
          padding: 60px 0 20px 0;
          text-align: center;
          span {
            color: rgb(21, 151, 229);
            font-weight: 500;
            margin-left: 10px;
          }
        }
      }
      .pledge-frame {
        height: 220px;
        background: url('@/assets/images/boxbg1.png') no-repeat;
        background-size: 100% 220px;
        box-sizing: border-box;
        display: flex;
        flex-wrap: wrap;
        .pledge-frame-item {
          width: 50%;
          height: 110px;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-direction: column;
          gap: 10px;
          p {
            &:nth-child(2) {
              font-size: 18px;
              font-weight: 500;
            }
          }
        }
      }
      .pledge-earnings {
        min-height: 88px;
        background: #0B1824;
        border-radius: 18px;
        display: flex;
        .pledge-earnings-item {
          height: 88px;
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 6px;
          p {
            &:nth-child(2) {
              font-size: 16px;
              font-weight: 500;
            }
          }
        }
      }
      .pledge-give {
        height: 58px;
        background-image: url(@/assets/images/btnbg.png);
        background-repeat: no-repeat;
        background-size: 100% 58px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        box-sizing: border-box;
        padding: 0 20px;
        span {
          &:nth-child(2) {
            font-size: 16px;
            font-weight: 500;
          }
        }
      }
    }
    .income-box {
      display: flex;
      flex-direction: column;
      .income-main {
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 20px;
        align-items: center;
        position: relative;
        padding-bottom: 20px;
        &::after {
          content: "";
          position: absolute;
          z-index: 1;
          bottom: 0;
          left: 0;
          width: 100%;
          height: 0.02564rem;
          background: linear-gradient(90deg,rgba(179,179,179,0) 0%,rgba(255,255,255,.6) 50.45%,rgba(179,179,179,0) 100%);
        }
        p {
          &:nth-child(1) {
            font-size: 14px;
            color: #CCC;
          }
          &:nth-child(2) {
            font-size: 26px;
            color: #FFF;
          }
        }
      }
      .income-footer {
        display: flex;
        flex-wrap: wrap;
        padding-top: 20px;
        gap: 20px 0;
        .income-footer-item {
          width: 25%;
          flex-grow: 1;
          flex-shrink: 0;
          align-items: center;
          display: flex;
          flex-direction: column;
          justify-content: flex-start;
          align-items: center;
          gap: 5px;
          p {
            &:nth-child(1) {
              font-size: 12px;
              color: #CCC;
            }
            &:nth-child(2) {
              font-size: 12px;
              color: #FFF;
              display: flex;
              gap: 4px;
              align-items: center;
            }
          }
        }
      }
    }
    .income-list {
      min-height: 100%;
      overflow: hidden;
      .list-menu-select {
        width: 100%;
        height: 40px;
        background: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAQAAAD9CzEMAAAAtUlEQVR42u2VUQqCQBRF3y4FoUAoCooCoUAQnJ21gfl0PzdIbOhS+UbfT/HO53g9B0QYcRzHcX4SBPSICJIJOuV7uGGgzdK3GOinpxEjjVrfYCRqPlHiqtJfkAgieYl6cl0j0QmhSZy/Lk+sn5M4flwdWD83sX+72LF+SWIrBDasX5qoXp5UrLdIrJ+nK9ZbJcrHScn/vWWiQMF64wTrTRP2ek6w3j7BevtERETwa9lxHOfvuAOAC4GPzKVVpAAAAABJRU5ErkJggg==') no-repeat right center;
        background-size: 18px 18px;
      }
      .income-list-header {
        width: 100%;
        height: 40px;
        line-height: 40px;
        overflow-x: auto;
        padding-bottom: 10px;
        &::-webkit-scrollbar {
          height: 0;
        }
        .header-list {
          height: 40px;
          line-height: 40px;
          padding: 0 5px;
          display: flex;
          gap: 10px;
          li {
            display: flex;
            align-items: center;
            white-space: nowrap;
            padding: 0 15px;
            border-radius: 6px;
            &.active {
              background: #0D1B2A;
            }
          }
        }
      }
      .income-list-main {
        display: flex;
        flex-direction: column;
        background: #0D1B2A;
        padding: 10px;
        .income-list-item {
          width: 100%;
          box-sizing:border-box;
          -moz-box-sizing:border-box;
          -webkit-box-sizing:border-box;
          padding: 10px;
          border-bottom: 1px solid #0A1724;
          &:nth-child(2n) {
            background: #0D1B2A;
          }
          .income-list-item-info {
            width: 100%;
            display: flex;
            flex-direction: column;
            gap: 4px;
            p {
              width: 100%;
              color: #CCC;
            }
            .income-list-item-time {
              font-size: 12px;
              display: flex;
              align-items: center;
              justify-content: space-between;
              gap: 8px;
            }
            .income-list-item-money {
              flex-shrink: 0;
              color: #CCC;
              font-size: 15px;
              font-weight: 500;
            }
          }
        }
      }
    }
  }
</style>
