<template>
<div class='withdraw-page'>
  <Header />
  <div class="page-main">
    <div class="withdraw-info">
      <p class="withdraw-balance">{{ $t('recharge.balance') }}: {{ displayAmount(rechargeBalance) }}</p>
      <p class="withdraw-balance">{{ $t('recharge.winBalance') }}: {{ displayAmount(winBalance) }}</p>
      <!-- <p v-if="winPrice > 0" class="withdraw-balance withdraw-price">{{ $t('recharge.winPrice') }}: {{ winPrice }} USDT</p> -->
      <div class="withdraw-actions">
        <button class="withdraw-btn" @click="showRecharge"><van-icon name="balance-pay" />{{ $t('recharge.recharge') }}</button>
        <button class="withdraw-btn" @click="router.push('/transfer')"><van-icon name="exchange" />{{ $t('transfer.title') }}</button>
      </div>
    </div>
    <div class="withdraw-tab">
      <div class="section-title-wrap">
        <div class="title-bar"></div>
        <h3 class="section-title">{{ $t('recharge.rechargeRecord') }}</h3>
      </div>
      <ul class="withdraw-tab-title">
        <li :class="{ active: recordTab === 'usdt' }" @click="switchRecordTab('usdt')">USDT</li>
        <li :class="{ active: recordTab === 'win' }" @click="switchRecordTab('win')">WIN</li>
      </ul>
      <div class="withdrawal-list">
        <div class="withdrawal-list-content">
          <div class="table">
            <div class="table-header" v-if="currentRecords.length > 0">
              <div class="table-row">
                <div class="table-cell">{{ recordTab === 'win' ? $t('recharge.winRecordIndex') : $t('recharge.date') }}</div>
                <div class="table-cell">{{ $t('recharge.amount') }}</div>
              </div>
            </div>
            <div class="table-body">
              <div class="table-row" v-for="(item, index) in currentRecords" :key="item.id ?? index">
                <div class="table-cell">{{ recordTab === 'win' ? item.index : item.createdAt }}</div>
                <div class="table-cell">{{ item.amount }} {{ recordTab === 'win' ? 'WIN' : 'USDT' }}</div>
              </div>
            </div>
          </div>
          <div class="empty" v-if="currentRecords.length === 0">
            <img :src="emptyImage" />
            <div class="empty-text">{{ $t('common.noData') }}</div>
          </div>
          <Pagination
            v-if="recordTab === 'usdt' && usdtRecords.length > 0 && allPageCount > 1"
            v-model="page"
            :page-count="allPageCount"
            mode="simple"
            @change="getUsdtRecords"
          />
        </div>
      </div>
    </div>
  </div>
  <RechargeDialog :getBalance="getBalance" :usdtBalance="usdtBalance" :onChange="handleRechargeChange" ref="rechargeDialogRef" />
</div>
</template>
<script setup>
import userPerson from "@/pinia/person";
import { useRouter } from 'vue-router'
import { Contract, ETH } from "@/tools/contract";
import { ref, computed } from 'vue'
import request from "@/tools/request";
import { closeToast, showLoadingToast } from "vant";
import RechargeDialog from "./subpage/components/rechargeDialog.vue";
import emptyImage from '../assets/images/custom-empty-image.png'
import { Pagination } from "vant"
import Header from '@/components/Header.vue'
import { useI18n } from 'vue-i18n'
import { fetchWinRechargeRecords } from '@/tools/winRecharge'
import { displayDecimal } from '@/tools/decimal'

const USDT = import.meta.env.VITE_USDT ? new Contract(import.meta.env.VITE_USDT, 'ERC20') : null
const BUY_USDT = new Contract(import.meta.env.VITE_BUY_USDT || import.meta.env.VITE_BUY, 'BUY')
const BUY = BUY_USDT // 兼容：USDT 授权检查用 USDT 充值合约

const router = useRouter()
const { t: $t } = useI18n()
const person = userPerson();
const userinfo = $computed(() => person.userinfo);
const profile = $computed(() => person.profile);
const rechargeBalance = $computed(() => String(profile.usdt_recharge || userinfo.usdt || '0'))
const winBalance = $computed(() => String(profile.win_balance || '0'))
const winPrice = $computed(() => Number(profile.win_price || 0))
const displayAmount = (value) => displayDecimal(value)
const rechargeDialogRef = ref(null)
const recordTab = ref('usdt')
let usdtRecords = $ref([])
let winRecords = $ref([])
const page = $ref(1)
let allPageCount = $ref(1)
let usdtBalance = $ref("0");
let usdtApproved = $ref(false);

const currentRecords = computed(() => recordTab.value === 'win' ? winRecords : usdtRecords)

const switchRecordTab = async (tab) => {
  if (recordTab.value === tab) return
  recordTab.value = tab
  if (tab === 'win') await getWinRecords()
}

const getUsdtApproved = async () => {
    if (!USDT) return false
    let res = await USDT.call("allowance", [ETH.account, BUY.address]);
    usdtApproved = Number(res) > 0;
    closeToast()
    return usdtApproved
}

const getBalance = async () => {
  await ETH.getAccount()
  const res = await ETH.getUSDTBalance()
  usdtBalance = res;
}

const getData = async () => {
    await Promise.allSettled([
      person.refreshProfile?.(),
      ...(import.meta.env.VITE_USDT ? [getBalance(), getUsdtApproved()] : []),
    ])
}

getData()

function showRecharge() {
  rechargeDialogRef.value?.open()
}

const getUsdtRecords = async (pageNum = 1) => {
    await request.get("app_server/deposit_list", {
      params: { page: pageNum }
    }).then((res) => {
      allPageCount = Math.max(1, Math.ceil(Number(res.count || 0) / 10));
      usdtRecords = res.list || []
      page = pageNum
    })
}

const getWinRecords = async () => {
  try {
    await ETH.getAccount()
    winRecords = await fetchWinRechargeRecords(ETH.account)
  } catch {
    winRecords = []
  }
}

const handleRechargeChange = async (pageNum = 1) => {
  await Promise.allSettled([
    person.refreshProfile?.(),
    getBalance(),
    getUsdtRecords(pageNum),
    getWinRecords(),
  ])
}

getUsdtRecords()

</script>
<style lang='scss' scoped>
@use '@/style/variables.scss' as *;

  .withdraw-page {
    height: 100vh;
    height: 100dvh;
    overflow: hidden;
  }

  .page-main {
    height: 100%;
    min-height: 0;
    padding: 100px 15px 0 15px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;

    .withdraw-info {
      flex: 0 0 auto;
      width: 100%;
      aspect-ratio: 694 / 310;
      height: auto;
      min-height: 0;
      overflow: hidden;
      background: url('@/assets/images/boxbg3.png') no-repeat center / 100% 100%;
      box-sizing: border-box;
      padding: 22px 20px;
      display: flex;
      margin-bottom: 20px;
      flex-direction: column;
      gap: 8px;
      align-items: flex-start;
      justify-content: center;
      .withdraw-balance {
        margin: 0;
        line-height: 1.35;
        font-size: 14px;
        color: $text-primary;
      }
      .withdraw-price {
        font-size: 12px;
        color: $text-muted;
      }
      .withdraw-actions {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 8px;
        margin-top: 4px;
      }
      .withdraw-btn {
        padding: 0 14px;
        display: inline-flex;
        height: 34px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid $border-color;
        border-radius: 22px;
        gap: 5px;
        align-items: center;
        justify-content: center;
        font-size: 13px;
        color: $brand-primary;
        transition: $transition-base;

        &:hover {
          background: $gradient-primary;
          color: $text-inverse;
          border-color: transparent;
        }

        i {
          font-size: 18px;
        }
      }
    }
    .withdraw-tab {
      flex: 1 1 auto;
      min-height: 0;
      display: flex;
      flex-direction: column;

      .section-title-wrap {
        flex: 0 0 auto;
        position: relative;
        margin-bottom: 8px;
        margin-left: 10px;
        display: flex;
        align-items: center;

        .title-bar {
          position: absolute;
          left: -10px;
          top: 50%;
          width: 4px;
          height: 16px;
          border-radius: 2px;
          background: linear-gradient(180deg, #1597E5 0%, #075FB8 100%);
          transform: translateY(-50%);
        }

        .section-title {
          margin: 0 0 0 8px;
          font-size: 16px;
          font-weight: bold;
          color: #fff;
        }
      }

      .withdraw-tab-title {
        flex: 0 0 44px;
        height: 44px;
        display: flex;
        gap: 20px;
        margin-top: 0;
        border-top-left-radius: 20px;
        border-top-right-radius: 20px;
        background: linear-gradient(0deg, rgba(5, 5, 5, 0), $bg-card);
        align-items: center;
        justify-content: flex-start;
        padding: 0 20px;
        li {
          font-size: 14px;
          color: $text-primary;
          position: relative;
          cursor: pointer;
          &.active {
            &::before {
              width: 25px;
              height: 2px;
              background: $brand-primary;
              content: '';
              display: block;
              position: absolute;
              bottom: -10px;
              left: 50%;
              transform: translateX(-50%);
              border-radius: 10px;
            }
          }
        }
      }
    }
    .empty {
      text-align: center;
      margin: 50px 0;
      img {
        width: 50px;
        height: 50px;
        margin: 0 auto;
        opacity: 0.3;
      }
      .empty-text {
        margin-top: 15px;
        font-size: 14px;
        color: rgba(255, 255, 255, 0.3);
      }
    }
    .withdrawal-list {
      flex: 1 1 auto;
      min-height: 0;
      padding: 20px;
      box-sizing: border-box;
      overflow-y: auto;
      overscroll-behavior: contain;
      -webkit-overflow-scrolling: touch;

      .withdrawal-list-content {
        min-height: 200px;
        .table {
          display: table;
          width: 100%;
          border-collapse: collapse;
          .table-header {
            display: table-header-group;
            font-weight: bold;
            border-bottom: 0.5px dashed $border-color;
            .table-row {
              width: 100%;
              display: table-row;
            }
            .table-cell {
              display: table-cell;
              padding: 10px;
              text-align: left;
              box-sizing: border-box;
              font-weight: 500;
              color: $text-muted;
              &:nth-child(n+2) {
                text-align: center;
              }
            }
          }
          .table-body {
            display: table-row-group;
            overflow-y: auto;
            .table-row {
              width: 100%;
              display: table-row;
              border-bottom: 1px solid $border-light;
              &:last-child {
                border-bottom: none;
              }
            }
            .table-cell {
              display: table-cell;
              text-align: left;
              box-sizing: border-box;
              padding: 10px;
              color: $text-primary;
              &:nth-child(n+2) {
                text-align: center;
              }
            }
          }
        }
      }
    }
  }
</style>
