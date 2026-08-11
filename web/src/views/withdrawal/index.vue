<template>
  <div class="withdrawal-page">
    <Header />

    <div class="content">
      <div class="page-header">
        <h1 class="page-title">{{ $t('withdraw.title') }}</h1>
        <p class="page-subtitle">AIX {{ $t('withdraw.title') }}</p>
        <p class="page-balance">{{ $t('withdraw.availableBalance') }}: {{ aixBalance }} AIX</p>
      </div>

      <div class="withdraw-form">
        <div class="form-hint-row">
          <p class="form-hint">{{ $t('withdraw.amount') }}</p>
          <button type="button" class="all-btn" @click="handleAllAmount()">
            {{ $t('withdraw.all') }}
          </button>
        </div>
        <div class="form-row">
          <input
            class="form-input"
            v-model="amountAix"
            @input="checkAixAmount"
            type="text"
            inputmode="decimal"
            :placeholder="$t('withdraw.enterAmount')"
          />
          <button
            class="subscribe-btn custom-btn"
            :disabled="!amountAix || loading"
            @click="handleWithdrawal()"
          >
            {{ loading ? $t('withdraw.processing') : $t('withdraw.confirm') }}
          </button>
        </div>
        <div class="form-info">
          <p>{{ $t('withdraw.fee') }}: 0 AIX</p>
        </div>
      </div>

      <div class="record-section">
        <div class="section-title-wrap">
          <div class="title-bar"></div>
          <h3 class="section-title">{{ $t('withdraw.details') }}</h3>
        </div>

        <div class="table-card">
          <div class="table-header table-header-3">
            <span>{{ $t('node.amount') }}</span>
            <span>{{ $t('withdraw.received') }}</span>
            <span>{{ $t('withdraw.status') }}</span>
          </div>
          <div class="order-list" v-for="(item, index) in amountList" :key="index">
            <div class="table-row table-row-3">
              <span>{{ item.amount }}</span>
              <span>{{ item.relAmount || '-' }}</span>
              <span class="status-cell">
                {{ withdrawStatusText(item.status) }}
                <small v-if="item.tx_hash" class="tx-hint">{{ String(item.tx_hash).slice(0, 10) }}…</small>
                <small class="muted">{{ item.createdAt }}</small>
              </span>
            </div>
          </div>
          <div class="empty-state" v-if="amountList.length === 0">
            <p>{{ $t('withdraw.noRecords') }}</p>
          </div>
          <div class="pagination-wrapper" v-if="amountList.length > 0 && allPageCount > 1">
            <Pagination
              v-model="allPage"
              :page-count="allPageCount"
              mode="simple"
              @change="getAmountList"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/Header.vue'
import userPerson from '@/pinia/person'
import request from '@/tools/request'
import { Pagination, showToast } from 'vant'
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n()
const person = userPerson()
const aixBalance = computed(() => String(person.profile?.aix_balance || '0'))

const allPage = ref(1)
const allPageCount = ref(1)
const amountAix = ref('')
const loading = ref(false)
const amountList = ref<any[]>([])

const withdrawStatusText = (status: string) => {
  switch (status) {
    case 'pending': return $t('withdraw.statusPendingReview')
    case 'rewarded':
    case 'approved': return $t('withdraw.statusPaying')
    case 'doing': return $t('withdraw.statusPaying')
    case 'pass': return $t('withdraw.statusReceived')
    case 'rejected': return $t('withdraw.statusRejected')
    case 'cancelled': return $t('withdraw.statusCancelled')
    default: return status || '-'
  }
}

const getAmountList = async (page: number = 1) => {
  try {
    const res: any = await request.get('app_server/withdraw_list', {
      params: { page }
    })
    allPageCount.value = Math.max(1, Math.ceil(Number(res.count || 0) / 10))
    amountList.value = res.list || []
    allPage.value = page
  } catch {
    amountList.value = []
  }
}

const handleAllAmount = () => {
  const max = Number(aixBalance.value)
  if (!Number.isFinite(max) || max <= 0) {
    showToast({ message: $t('withdraw.insufficientBalance'), position: 'center' })
    return
  }
  amountAix.value = String(max)
}

const checkAixAmount = (e: any) => {
  // 仅规范化数字格式，不按可提余额钳制输入（超额在提交时由前端/后端校验）
  let raw = String(e?.target?.value ?? amountAix.value ?? '')
  raw = raw.replace(/[^\d.]/g, '')
  const parts = raw.split('.')
  if (parts.length > 2) {
    raw = parts[0] + '.' + parts.slice(1).join('')
  }
  if (parts[1] != null && parts[1].length > 8) {
    raw = parts[0] + '.' + parts[1].slice(0, 8)
  }
  amountAix.value = raw
}

const handleWithdrawal = async () => {
  if (loading.value) return
  const amount = Number(amountAix.value)
  const maxBal = Number(aixBalance.value)
  if (!Number.isFinite(amount) || amount <= 0) {
    showToast({ message: $t('withdraw.enterAmount'), position: 'center' })
    return
  }
  if (!Number.isFinite(maxBal) || maxBal <= 0) {
    showToast({
      message: $t('withdraw.insufficientHint'),
      position: 'center',
    })
    return
  }
  if (amount > maxBal) {
    showToast({ message: $t('withdraw.insufficientBalance'), position: 'center' })
    return
  }
  loading.value = true
  try {
    const res: any = await request.post('app_server/withdraw', {
      amount: String(amount),
      nosuccess: true,
    })
    if (res.status === 'ok') {
      showToast({
        message: $t('withdraw.submittedProcessing'),
        position: 'center',
        duration: 2000,
      })
      amountAix.value = ''
      await person.refreshProfile?.()
      await getAmountList(1)
    } else {
      showToast({
        message: res.message || $t('common.operationFailed'),
        position: 'center',
        duration: 2000,
      })
    }
  } catch (e: any) {
    const msg = e?.response?.data?.message || (typeof e === 'string' ? e : '') || $t('common.operationFailed')
    showToast({ message: msg, position: 'center', duration: 2000 })
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    person.getUser?.(),
    person.refreshProfile?.(),
    getAmountList(),
  ])
})
</script>

<style lang="scss" scoped>
@use '@/style/variables.scss' as *;

.withdrawal-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #030A11 0%, #0D1B2A 100%);
}

.content {
  padding: 90px 20px 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 20px;

  .page-title {
    font-size: 16px;
    font-weight: bold;
    color: #fff;
    margin-bottom: 8px;
  }

  .page-subtitle {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.6);
  }

  .page-balance {
    font-size: 14px;
    color: $brand-primary;
    margin-top: 8px;
  }
}

.withdraw-form {
  margin-bottom: 40px;

  .form-hint-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }

  .form-hint {
    margin: 0;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
  }

  .all-btn {
    padding: 0;
    border: none;
    background: transparent;
    color: $brand-primary;
    font-size: 13px;
    cursor: pointer;
  }

  .form-row {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .form-input {
    flex: 1;
    height: 44px;
    padding: 0 14px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.25);
    color: #fff;
    font-size: 15px;
    outline: none;
    caret-color: $brand-primary;
    -webkit-text-fill-color: #fff;

    &::placeholder {
      color: rgba(255, 255, 255, 0.4);
      -webkit-text-fill-color: rgba(255, 255, 255, 0.4);
    }

    &:focus {
      border-color: $brand-primary;
    }
  }

  .form-info {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;

    p {
      margin: 0;
      font-size: 12px;
      color: rgba(255, 255, 255, 0.5);
    }
  }
}

.subscribe-btn {
  padding: 8px 20px;
  background: $gradient-primary;
  color: $text-inverse;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 400;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $brand-primary-light 0%, $brand-primary 100%);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(21, 151, 229, 0.3);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.5);
  }

  &.custom-btn {
    flex-shrink: 0;
    min-width: 120px;
    width: auto;
    height: 44px;
    padding: 0 20px;
  }
}

.record-section {
  margin-top: 20px;

  .section-title-wrap {
    position: relative;
    margin-bottom: 10px;
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
}

.table-card {
  margin-top: 10px;
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

  .order-list {
    .table-row {
      display: flex;
      align-items: center;
      padding: 12px 0;
      border-bottom: 1px solid $border-light;

      &:last-child {
        border-bottom: none;
      }

      span {
        flex: 1;
        text-align: center;
        font-size: 14px;
        color: $text-primary;
      }
    }
  }

  .status-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;

    .tx-hint {
      font-size: 10px;
      color: $text-muted;
    }
  }

  .cancel-btn {
    padding: 4px 10px;
    border: 1px solid rgba(255, 255, 255, 0.25);
    border-radius: 6px;
    background: transparent;
    color: #fff;
    font-size: 12px;
    cursor: pointer;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  .muted {
    font-size: 11px;
    color: $text-muted;
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

  .pagination-wrapper {
    padding: 16px 0;
    display: flex;
    justify-content: center;
  }
}
</style>
