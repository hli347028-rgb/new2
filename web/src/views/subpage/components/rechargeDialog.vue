<template>
  <a-modal forceRender :maskClosable="false" v-model:open="isOpen" :footer="null" centered destroyOnClose :title="null" @ok="handleOk">
    <div class='withdraw-dialog'>
      <div class="dialog-main">
        <div class="dialog-title">{{ lang('当前余额') }}：{{ usdtBalance }}</div>
        <a-input-number style="width: 100%" v-model:value="amount" :min="10" size="large" :placeholder="lang('请输入数量')" />
        <div class="dialog-info">
        <p><QuestionCircleOutlined style="margin-right: 5px" />{{ lang('最低充值金额') }}: 10</p>
        <p></p>
        </div>
      </div>
      <a-button class="withdraw-btn" :disabled="loading" size="large" @click="handleWithdrawal" type="primary">{{lang('充值')}}</a-button>
    </div>
  </a-modal>
</template>
<script setup>
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import userPerson from "@/pinia/person";
import fetchSign from '@/pinia/fetchSign'
import { Contract, ETH } from "@/tools/contract";
import { showToast, showLoadingToast, closeToast, showDialog } from 'vant'
import lang from '@/i18n/index'
const BUY = new Contract(import.meta.env.VITE_BUY, "BUY");

const person = userPerson();
const userinfo = $computed(() => person.userinfo);
// const sign = $computed(() => person.sign);
const isOpen = $ref(false)
const amount = $ref(null)
const type = $ref('USDT')
const loading = $ref(false)


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

const amountRound = (num) => {
  return Math.round(num * 100) / 100
}

const open = (t) => {
  type = t
  isOpen = true

}

const transferUsdt = async (count) => {
  // person.userinfo.status = 'running'
  BUY.send("buy", [count],).then(() => {
    console.log('充值成功')
    loading = false;
    closeToast()
    showDialog({
      title: lang('提示'),
      message: lang(`充值成功`),
      theme: 'round-button',
      confirmButtonColor: "#242738",
      confirmButtonText: lang('我知道了！'),
    }).then(async () => {
      person.getUser();
      props.onChange()
      await props.getBalance()
      isOpen = false
    })
  }).catch((error) => {
    console.log(error)
    loading = false
  });
}

const handleWithdrawal = async () => {
  if (loading) return

  showLoadingToast({
      message: lang('充值中'), duration: 0, overlay: true, overlayStyle: {
          background: "transparent"
      }
  });
  loading = true

  if (amount < 5) {
    loading = false
    return showToast(lang('充值金额不能小于5'))
  }
  try {
    transferUsdt(amount)
  } catch (error) {
    loading = false
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
    background: $gradient-gold;
    border: none;
    color: $text-inverse;
    
    &:hover:not(:disabled) {
      background: linear-gradient(135deg, $brand-gold-light 0%, $brand-gold 100%);
    }
    
    &:disabled {
      opacity: 0.5;
    }
  }
}
</style>