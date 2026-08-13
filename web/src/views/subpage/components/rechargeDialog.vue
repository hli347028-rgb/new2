<template>
  <a-modal
    :maskClosable="false"
    v-model:open="isOpen"
    :footer="null"
    centered
    destroyOnClose
    :title="null"
    :width="360"
    wrap-class-name="recharge-modal"
    :body-style="{ padding: '0' }"
  >
    <div class="withdraw-dialog">
      <h3 class="dialog-heading">{{ $t('recharge.recharge') }}</h3>

      <a-radio-group v-model:value="assetType" button-style="solid" class="asset-tabs" @change="resetAmount">
        <a-radio-button value="usdt">USDT</a-radio-button>
        <a-radio-button value="win">WIN</a-radio-button>
      </a-radio-group>

      <div class="dialog-main">
        <template v-if="assetType === 'usdt'">
          <p class="dialog-subtitle dialog-subtitle--placeholder" aria-hidden="true">&nbsp;</p>
          <a-input-number
            autofocus
            style="width: 100%"
            v-model:value="amount"
            :min="MIN_USDT_RECHARGE"
            size="large"
            :placeholder="$t('recharge.enterAmount')"
          />
          <div class="dialog-info">
            <p><QuestionCircleOutlined style="margin-right: 5px" />{{ $t('recharge.minRechargeAmount') }}: {{ MIN_USDT_RECHARGE }} USDT</p>
          </div>
        </template>

        <template v-else>
          <p class="dialog-subtitle">{{ $t('recharge.winPrice') }}: {{ winPrice > 0 ? winPrice : '-' }} USDT</p>
          <a-input-number
            autofocus
            style="width: 100%"
            v-model:value="amount"
            :min="MIN_WIN_RECHARGE"
            size="large"
            :placeholder="$t('recharge.enterAmount')"
          />
          <div class="dialog-info">
            <p>{{ $t('recharge.minWinRecharge') }}</p>
          </div>
        </template>
      </div>

      <a-button class="withdraw-btn" :disabled="loading" size="large" @click="handleSubmit" type="primary">
        {{ $t('recharge.recharge') }}
      </a-button>
    </div>
  </a-modal>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ethers } from 'ethers'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import userPerson from '@/pinia/person'
import { Contract, ETH } from '@/tools/contract'
import { showToast, showLoadingToast, closeToast, showDialog } from 'vant'
import { useI18n } from 'vue-i18n'
import { confirmWinRecharge, createWinRecharge, errMsg } from '@/api/aix'
import { displayDecimal, isPositiveDecimal } from '@/tools/decimal'

const BUY = new Contract(import.meta.env.VITE_BUY, 'BUY')
const USDT = new Contract(import.meta.env.VITE_USDT, 'ERC20')
const MIN_USDT_RECHARGE = 10
const MIN_WIN_RECHARGE = 1

const person = userPerson()
const { t: $t } = useI18n()
const isOpen = ref(false)
const assetType = ref('usdt')
const amount = ref(null)
const loading = ref(false)

const props = defineProps({
  getBalance: {
    type: Function,
    required: true,
  },
  onChange: {
    type: Function,
    required: true,
  },
  usdtBalance: {
    type: String,
  },
})

const winBalance = computed(() => String(person.profile?.win_balance || '0'))
const winPrice = computed(() => Number(person.profile?.win_price || 0))
const displayAmount = (value) => displayDecimal(value)

const open = async () => {
  assetType.value = 'usdt'
  amount.value = null
  await person.refreshProfile?.()
  isOpen.value = true
}

const resetAmount = () => {
  amount.value = null
}

async function transferWin(contractAddress, to, amountText, decimals) {
  const WIN = new Contract(contractAddress, 'ERC20')
  const amountWei = ethers.utils.parseUnits(amountText, decimals)
  const tx = await WIN.getInsance().transfer(to, amountWei)
  const receipt = await ETH.provider.waitForTransaction(tx.hash)
  if (receipt.status !== 1) throw new Error($t('common.operationFailed'))
  return String(tx.hash)
}

const transferUsdt = async (count) => {
  await BUY.send('buy', [count])
  closeToast()

  await showDialog({
    title: $t('common.prompt'),
    message: $t('recharge.success'),
    theme: 'round-button',
    confirmButtonColor: '#0A1724',
    confirmButtonText: $t('common.gotIt'),
  })

  await person.getUser()
  await props.onChange?.()
  await props.getBalance()
  isOpen.value = false
}

const submitUsdtRecharge = async () => {
  if (!(Number(amount.value) >= MIN_USDT_RECHARGE)) {
    showToast($t('recharge.minimumError', { amount: MIN_USDT_RECHARGE }))
    return
  }

  await ETH.getAccount()
  const allowance = await USDT.call('allowance', [ETH.account, BUY.address])
  if (!(Number(allowance) > 0)) {
    await USDT.send('approve', [
      BUY.address,
      '115792089237316195423570985008687907853269984665640564039457584007913129639935',
    ])
  }
  await transferUsdt(amount.value)
}

const submitWinRecharge = async () => {
  const value = String(amount.value ?? '')
  if (!isPositiveDecimal(value) || Number(value) < MIN_WIN_RECHARGE) {
    showToast($t('recharge.minWinRechargeError'))
    return
  }

  await ETH.getAccount()
  const order = await createWinRecharge(value)
  let txHash = ''

  if (order.win_contract && order.deposit_address && !order.dev_mode) {
    txHash = await transferWin(
      order.win_contract,
      order.deposit_address,
      value,
      Number(order.win_decimals || 18),
    )
  }

  const signature = await ETH.signMessage(order.message)
  await confirmWinRecharge(order.recharge_id, txHash, signature)

  closeToast()
  await showDialog({
    title: $t('common.prompt'),
    message: $t('recharge.winRechargeSuccess'),
    theme: 'round-button',
    confirmButtonColor: '#0A1724',
    confirmButtonText: $t('common.gotIt'),
  })

  await Promise.allSettled([
    person.refreshProfile?.(),
    props.onChange?.(),
  ])
  isOpen.value = false
}

const handleSubmit = async () => {
  if (loading.value) return

  loading.value = true
  showLoadingToast({
    message: $t('recharge.processing'),
    duration: 0,
    overlay: true,
    overlayStyle: { background: 'transparent' },
  })

  try {
    if (assetType.value === 'usdt') {
      await submitUsdtRecharge()
    } else {
      await submitWinRecharge()
    }
  } catch (error) {
    console.error('充值失败:', error)
    showToast(errMsg(error, $t('common.operationFailed')))
  } finally {
    loading.value = false
    closeToast()
  }
}

defineExpose({ open })
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.withdraw-dialog {
  height: 260px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

.dialog-heading {
  margin: 0 0 16px;
  padding-bottom: 14px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.2;
  color: #fff;
  letter-spacing: 0.02em;
  background: linear-gradient(90deg, #fff 0%, $brand-primary-light 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.asset-tabs {
  width: 100%;
  display: flex;
  margin-bottom: 16px;

  :deep(.ant-radio-button-wrapper) {
    flex: 1;
    height: 38px;
    line-height: 36px;
    text-align: center;
    font-size: 14px;
    border-color: rgba(21, 151, 229, 0.25);
    background: rgba(3, 10, 17, 0.45);
    color: $text-muted;

    &::before {
      display: none;
    }

    &:first-child {
      border-radius: 10px 0 0 10px;
    }

    &:last-child {
      border-radius: 0 10px 10px 0;
    }
  }

  :deep(.ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)) {
    background: $gradient-primary;
    border-color: transparent;
    color: $text-inverse;
    font-weight: 600;
  }
}

.dialog-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.balance-line {
  min-height: 44px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(21, 151, 229, 0.08);
  border: 1px solid rgba(21, 151, 229, 0.14);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  span {
    font-size: 12px;
    color: $text-muted;
  }

  strong {
    font-size: 15px;
    font-weight: 600;
    color: $brand-primary-light;
    word-break: break-all;
    text-align: right;
  }
}

.dialog-subtitle {
  margin: 0;
  min-height: 18px;
  font-size: 12px;
  line-height: 18px;
  color: $text-muted;
}

.dialog-subtitle--placeholder {
  visibility: hidden;
}

.dialog-info {
  margin-top: auto;
  display: flex;
  justify-content: flex-end;

  p {
    margin: 0;
    text-align: right;
    color: $text-muted;
    font-size: 12px;
  }
}

.withdraw-btn {
  width: 100%;
  height: 44px;
  margin-top: 16px;
  background: $gradient-primary;
  border: none;
  color: $text-inverse;
  border-radius: 22px;
  font-size: 15px;
  font-weight: 600;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $brand-primary-light 0%, $brand-primary 100%);
  }

  &:disabled {
    opacity: 0.5;
  }
}
</style>

<style lang="scss">
.recharge-modal .ant-modal-content {
  border-radius: 16px;
  overflow: hidden;
}
</style>
