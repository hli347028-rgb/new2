<template>
  <a-modal :maskClosable="false" v-model:open="isOpen" :footer="null" centered destroyOnClose :title="null" @ok="handleOk">
    <div class='withdraw-dialog'>
      <div class="dialog-main">
        <div class="dialog-title">当前充值余额:{{ usdtBalance }}</div>
        <a-input-number autofocus style="width: 100%" v-model:value="amount" :min="MIN_RECHARGE_AMOUNT" size="large" :placeholder="lang('请输入数量')" />
        <div class="dialog-info">
        <p><QuestionCircleOutlined style="margin-right: 5px" />{{ lang('最低充值金额') }}: {{ MIN_RECHARGE_AMOUNT }}</p>
        <p></p>
        </div>
      </div>
      <a-button class="withdraw-btn" :disabled="loading" size="large" @click="handleWithdrawal" type="primary">{{lang('充值')}}</a-button>
    </div>
  </a-modal>
</template>
<script setup>
import { ref } from 'vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import userPerson from "@/pinia/person";
import { Contract, ETH } from "@/tools/contract";
import { showToast, showLoadingToast, closeToast, showDialog } from 'vant'
import lang from '@/i18n/index'
const BUY = new Contract(import.meta.env.VITE_BUY, "BUY");
const USDT = new Contract(import.meta.env.VITE_USDT, "ERC20");

const person = userPerson();
const isOpen = ref(false)
const amount = ref(null)
const loading = ref(false)
const MIN_RECHARGE_AMOUNT = 100


const props = defineProps({
  getBalance: {
    type: Function,
    required: true
  },
  onChange: {
    type: Function,
    required: true
  },
  usdtBalance: {
    type: String
  }
})

const open = () => {
  amount.value = null
  isOpen.value = true
}

const transferUsdt = async (count) => {
  // 分账合约：buy(num) — num 为 USDT 整数金额
  await BUY.send("buy", [count])
  closeToast()

  await showDialog({
    title: lang('提示'),
    message: lang('充值成功'),
    theme: 'round-button',
    confirmButtonColor: "#0A1724",
    confirmButtonText: lang('我知道了！'),
  })

  await person.getUser()
  props.onChange()
  await props.getBalance()
  isOpen.value = false
}

const handleWithdrawal = async () => {
  if (loading.value) return

  if (!(Number(amount.value) >= MIN_RECHARGE_AMOUNT)) {
    showToast(lang('充值金额不能小于5'))
    return
  }

  loading.value = true
  showLoadingToast({
    message: lang('充值中'),
    duration: 0,
    overlay: true,
    overlayStyle: { background: "transparent" }
  })

  try {
    await ETH.getAccount()
    // 先确保授权
    const allowance = await USDT.call("allowance", [ETH.account, BUY.address]);
    if (!(Number(allowance) > 0)) {
      await USDT.send("approve", [
        BUY.address,
        "115792089237316195423570985008687907853269984665640564039457584007913129639935"
      ])
    }
    await transferUsdt(amount.value)
  } catch (error) {
    console.error('充值失败:', error)
    showToast(lang('操作失败'))
  } finally {
    loading.value = false
    closeToast()
  }
}

const handleOk = () => {
  console.log('WithdrawDialog handleOk')
}

defineExpose({
  open
})
</script>
<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.withdraw-dialog {
  display: flex;
  flex-direction: column;
  align-items: center;
  
  .dialog-title {
    width: 100%;
    font-size: 14px;
    font-weight: 500;
    color: $text-primary;
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
      color: $text-muted;
    }
  }
  
  .withdraw-btn {
    width: 160px;
    background: $gradient-primary;
    border: none;
    color: $text-inverse;
    
    &:hover:not(:disabled) {
      background: linear-gradient(135deg, $brand-primary-light 0%, $brand-primary 100%);
    }
    
    &:disabled {
      opacity: 0.5;
    }
  }
}
</style>
