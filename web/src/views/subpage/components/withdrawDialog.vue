<template>
  <a-modal :maskClosable="false" v-model:open="isOpen" :footer="null" centered destroyOnClose :title="null" @ok="handleOk">
    <div class='withdraw-dialog'>
      <div class="dialog-title">{{ type === 'USDT' ? lang('当前USDT余额') : lang('当前ISPAY余额') }}：{{type === 'USDT' ? userinfo.amountGet : type === 'NEWISPAY' ? userinfo.ispayAmount : userinfo.raw}}</div>
      <div class="dialog-main">
        <a-input-number
          autofocus
          style="width: 100%"
          v-model:value="amount"
          :min="0"
          :precision="2"
          size="large"
          :placeholder="lang('请输入数量')"
        />
        <div class="dialog-info">
        <p>{{lang('手续费')}}：{{ amountRound(Number(userinfo.withdrawRate) * (amount || 0)) }}</p>
        </div>
      </div>
      <a-button class="withdraw-btn" :disabled="loading" size="large" @click="handleWithdrawal" type="primary">{{lang('提现')}}</a-button>
    </div>
  </a-modal>
</template>
<script setup>
import userPerson from "@/pinia/person";
import fetchSign from '@/pinia/fetchSign'
import { showToast } from 'vant'
import request from "@/tools/request";
import lang from '@/i18n/index'

const person = userPerson();
const userinfo = $computed(() => person.userinfo);
// const sign = $computed(() => person.sign);
const isOpen = $ref(false)
const amount = $ref(null)
const type = $ref('USDT')
const loading = $ref(false)

const props = defineProps({
  onChange: {
    type: Function,
    required: true
  }
})

const amountRound = (num) => {
  return Math.round(num * 100) / 100
}

const balanceMax = $computed(() => {
  const raw = type === 'USDT'
    ? userinfo.amountGet
    : type === 'NEWISPAY'
      ? userinfo.ispayAmount
      : userinfo.raw
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : 0
})

const open = (t) => {
  console.log('WithdrawDialog open', t)
  type = t
  amount = null
  isOpen = true
}

const handleWithdrawal = async () => {
  console.log('WithdrawDialog handleWithdrawal', amount)
  if (loading) return
  loading = true
  const amt = Number(amount)
  if (!Number.isFinite(amt) || amt <= 0) {
    loading = false
    return showToast(lang('recharge.enterAmount'))
  }
  if (balanceMax <= 0) {
    loading = false
    return showToast(lang('withdraw.insufficientHint'))
  }
  if (amt > balanceMax) {
    loading = false
    return showToast(lang('withdraw.amountExceedsBalance'))
  }

  if (type === 'ISPAY') {
    loading = false
    return showToast(lang('withdraw.ispayNotSupported'))
  }

  const sign = await fetchSign()

  await request.post("app_server/withdraw", {
    amount: String(amt),
    sign: sign,
    coinType: type === 'NEWISPAY' ? 3 : undefined,
    nosuccess: true,
  }).then((res) => {
    loading = false

    if (res.status === 'ok') {
      isOpen = false
      amount = null
      props.onChange()
      showToast({
        message: lang('withdraw.submittedProcessing'),
        position: 'center',
        duration: 2000,
      });
    } else {
      showToast({
        message: res.status || lang('withdraw.failed'),
        position: 'center',
        duration: 2000,
      });
		}
  }).catch((err) => {
    loading = false
    showToast({
      message: lang('withdraw.failed'),
      position: 'center',
      duration: 2000,
    });
  })
}

const handleOk = () => {
  console.log('WithdrawDialog handleOk')
}

defineExpose({
  open
})
</script>
<style lang="less" scoped>
.withdraw-dialog {
  display: flex;
  flex-direction: column;
  align-items: center;
  .dialog-title {
    width: 100%;
    font-size: 14px;
    font-weight: 500;;
  }
  .dialog-main {
    width: 100%;
    margin: 20px 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
    .dialog-info {
      display: flex;
      justify-content: space-between;
    }
    p {
      text-align: right;
      color: #999;
    }
  }
  .withdraw-btn {
    width: 160px;
  }
}
</style>
