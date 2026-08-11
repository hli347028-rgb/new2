<template>
<div class='withdraw-page'>
  <Header />
  <div class="page-main">
    <div class="withdraw-info">
      <p class="withdraw-balance">充值余额: {{ userinfo.usdt || 0.0000 }}</p>
      <!-- <p class="withdraw-balance">BRC20: {{ userinfo.rawNew || 0.0000 }}</p> -->
      <div class="withdraw-actions">
        <button class="withdraw-btn" @click="showWithdraw"><van-icon name="balance-pay" />{{ $t('recharge.recharge') }}</button>
        <button class="withdraw-btn" @click="router.push('/transfer')"><van-icon name="exchange" />划转</button>
      </div>
    </div>
    <div class="withdraw-tab">
      <ul class="withdraw-tab-title">
        <li class="active">{{ $t('recharge.rechargeRecord') }}</li>
      </ul>
      <div class="withdrawal-list">
        <div class="withdrawal-list-content">
          <div class="table">
            <div class="table-header" v-if="amountList.length > 0">
              <div class="table-row">
                <div class="table-cell">{{ $t('recharge.date') }}</div>
                <div class="table-cell">{{ $t('recharge.amount') }}</div>
              </div>
            </div>
            <div class="table-body">
              <div class="table-row" v-for="(item, index) in amountList" :key="index">
                <div class="table-cell">{{item.createdAt}}</div>
                <div class="table-cell">{{item.amount}}</div>
              </div>
            </div>
          </div>
          <div class="empty" v-if="amountList.length === 0">
            <img :src="emptyImage" />
            <div class="empty-text">{{ $t('common.noData') }}</div>
          </div>
          <Pagination
            v-if="amountList.length > 0"
            v-model="page"
            :page-count="allPageCount"
            mode="simple"
            @change="getAmountList"
          />
        </div>
      </div>
    </div>
  </div>
  <RechargeDialog :getBalance="getBalance" :usdtBalance="usdtBalance" :onChange="getAmountList" ref="rechargeDialogRef" />
</div>
</template>
<script setup>
import userPerson from "@/pinia/person";
import { useRouter, useRoute } from 'vue-router'
import { Contract, ETH } from "@/tools/contract";
import { ref, computed } from 'vue'
import request from "@/tools/request";
import { showLoadingToast, closeToast, showFailToast, showDialog, closeDialog, showSuccessToast, showToast } from "vant";
import RechargeDialog from "./subpage/components/rechargeDialog.vue";
import emptyImage from '../assets/images/custom-empty-image.png'
import { Pagination } from "vant"
import lang from '@/i18n/index'
import Header from '@/components/Header.vue'

const USDT = new Contract(import.meta.env.VITE_USDT, "ERC20");
const BUY = new Contract(import.meta.env.VITE_BUY, "BUY");
const route = useRoute()

const router = useRouter()
const person = userPerson();
const userinfo = $computed(() => person.userinfo);
const rechargeDialogRef = ref(null)
let amountList = $ref([])
const page = $ref(1)
let allPageCount = $ref(1)
let usdtBalance = $ref("0");
let usdtApproved = $ref(false);

/* 获取授权 */
const getUsdtApproved = async () => {
    let res = await USDT.call("allowance", [ETH.account, BUY.address]);
    console.log('getUsdtApproved', Number(res))
    usdtApproved = Number(res) > 0;
    closeToast()

    return usdtApproved
}

const usdtApprove = async () => {
    showLoadingToast({
        message: lang('授权中'), duration: 0, overlay: true, overlayStyle: {
            background: "transparent"
        }
    });
    await USDT.send("approve", [
        BUY.address,
        "115792089237316195423570985008687907853269984665640564039457584007913129639935"
    ]).then(getUsdtApproved).catch(() => closeToast());
}

/* 获取代币余额 */
const getBalance = async () => {
  await ETH.getAccount()
  const res = await ETH.getUSDTBalance()
  usdtBalance = res;
}

const getData = async () => {
    // showLoadingToast({
    //     message: lang('授权中'), duration: 0, overlay: true, overlayStyle: {
    //         background: "transparent"
    //     }
    // });
    await getBalance();
    await getUsdtApproved();
    // closeToast();
}

getData()

function showWithdraw() {
  rechargeDialogRef.value?.open('USDT')
}


const getAmountList = async (page = 1) => {
    await request.get("app_server/deposit_list", {
      params: {
        page
      }
    }).then((res) => {
      allPageCount = Math.ceil(res.count / 10);
      amountList = res.list
    })
}

getAmountList()

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
      background: url('@/assets/images/boxbg3.png') no-repeat center;
      background-size: cover;
      box-sizing: border-box;
      padding:30px 20px;
      display: flex;
      margin-bottom: 20px;
      flex-direction: column;
      gap: 10px;
      align-items: flex-start;
      justify-content: space-around;
      .withdraw-balance {
        font-size: 16px;
        color: $text-primary;
      }
      .withdraw-name {
        font-size: 16px;
        color: $text-primary;
      }
      .withdraw-actions {
        display: flex;
        align-items: center;
        gap: 10px;
        // margin-top: 10px;
      }
      .withdraw-btn {
        padding: 0 20px;
        display: inline-flex;
        height: 36px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid $border-color;
        border-radius: 22px;
        display: flex;
        gap: 5px;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        color: $brand-primary;
        transition: $transition-base;

        &:hover {
          background: $gradient-primary;
          color: $text-inverse;
          border-color: transparent;
        }

        i {
          font-size: 20px;
        }
      }
    }
    .withdraw-tab {
      flex: 1 1 auto;
      min-height: 0;
      display: flex;
      flex-direction: column;

      .withdraw-tab-title {
        flex: 0 0 44px;
        height: 44px;
        display: flex;
        gap: 20px;
        margin-top: 20px;
        border-top-left-radius: 20px;
        border-top-right-radius: 20px;
        background: linear-gradient(0deg, rgba(5, 5, 5, 0), $bg-card);
        display: flex;
        align-items: center;
        justify-content: flex-start;
        padding: 0 20px;
        li {
          font-size: 14px;
          color: $text-primary;
          position: relative;
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
    .withdraw-tab-content {
      padding: 50px 0;
    }
    .withdrawal-list {
      flex: 1 1 auto;
      min-height: 0;
      padding: 20px;
      box-sizing: border-box;
      overflow-y: auto;
      overscroll-behavior: contain;
      -webkit-overflow-scrolling: touch;

      .withdrawal-list-title {
        font-size: 16px;
      }
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
              &:nth-child(2) {
                text-align: center;
              }
              &:nth-child(3) {
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
              p {
                height: 16px;
                padding: 5px 0;
                margin: 0;
              }
              &:nth-child(2) {
                text-align: center;
              }
              &:nth-child(3) {
                text-align: center;
              }
            }
          }
        }
      }
    }
    .recharge {
      min-height: 200px;
      background: rgba(8, 19, 30, .8);
      border: 1px solid #666;
      border-radius: 18px;
      padding: 15px;
      .recharge-info {
        .recharge-info-list {
          li {
            height: 32px;
            line-height: 32px;
            border-bottom: 1px solid #EEE;
            &::last-child {
              border-bottom: 0;
            }
          }
        }
      }
    }
  }
</style>
