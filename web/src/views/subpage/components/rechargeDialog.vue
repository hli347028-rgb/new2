<template>
  <a-modal :maskClosable="false" v-model:open="isOpen" :footer="null" centered destroyOnClose :title="null" @ok="handleOk">
    <div class='withdraw-dialog'>
      <div class="dialog-main">
        <div class="dialog-title">{{ $t('recharge.currentRechargeBalance') }}: {{ usdtBalance }}</div>
        <a-input-number autofocus style="width: 100%" v-model:value="amount" :min="MIN_RECHARGE_AMOUNT" size="large" :placeholder="$t('recharge.enterAmount')" />
        <div class="dialog-info">
        <p><QuestionCircleOutlined style="margin-right: 5px" />{{ $t('recharge.minRechargeAmount') }}: {{ MIN_RECHARGE_AMOUNT }}</p>
        <p></p>
        </div>
      </div>
      <a-button class="withdraw-btn" :disabled="loading" size="large" @click="handleWithdrawal" type="primary">{{ $t('recharge.recharge') }}</a-button>
    </div>
  </a-modal>
</template>
<script setup>
import { ref } from 'vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import userPerson from "@/pinia/person";
import { Contract, ETH } from "@/tools/contract";
import { showToast, showLoadingToast, closeToast, showDialog } from 'vant'
import { useI18n } from 'vue-i18n'
const BUY = new Contract(import.meta.env.VITE_BUY, "BUY");
const USDT = new Contract(import.meta.env.VITE_USDT, "ERC20");

const person = userPerson();
const { t: $t } = useI18n()
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
    title: $t('common.prompt'),
    message: $t('recharge.success'),
    theme: 'round-button',
    confirmButtonColor: "#0A1724",
    confirmButtonText: $t('common.gotIt'),
  })

  await person.getUser()
  props.onChange()
  await props.getBalance()
  isOpen.value = false
}

const handleWithdrawal = async () => {
  if (loading.value) return

  if (!(Number(amount.value) >= MIN_RECHARGE_AMOUNT)) {
    showToast($t('recharge.minimumError', { amount: MIN_RECHARGE_AMOUNT }))
    return
  }

  loading.value = true
  showLoadingToast({
    message: $t('recharge.processing'),
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
    showToast($t('common.operationFailed'))
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
