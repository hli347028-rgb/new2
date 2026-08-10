import { defineStore } from 'pinia'
import { ETH } from '@/tools/contract'
import { fetchChallenge, login as apiLogin, getAixProfile, getBalance, errMsg, clearAuth } from '@/api/aix'
import { showFailToast, showToast, showDialog, setToastDefaultOptions } from 'vant'

setToastDefaultOptions({ zIndex: 6001 })

let pollTimer: any = null

export default defineStore('person', {
  state: () => ({
    loadAccount: false,
    isLogin: false,
    address: '',
    urlCode: '',
    profile: {
      address: '',
      usdt_recharge: '0',
      usdt_reward: '0',
      aix_balance: '0',
      static_usdt_total: '0',
      pending_amount: '0',
      unexited_amount: '0',
      total_nodes: 0,
      mgmt_level: 0,
      small_area_perf: '0',
      team_perf: '0',
      next_release_at: 0,
      server_time: 0,
      aix_contract: '',
    } as Record<string, any>,
  }),
  actions: {
    parseInviteCodeFromUrl() {
      const extract = (raw: string) => {
        const value = (raw || '').trim()
        if (!value) return ''
        if (/^0x[a-fA-F0-9]{40}$/i.test(value)) return value
        return value
      }
      const hash = window.location.hash || ''
      const hashQuery = hash.includes('?') ? hash.slice(hash.indexOf('?') + 1) : ''
      if (hashQuery) {
        const params = new URLSearchParams(hashQuery)
        const fromHash = extract(params.get('inviteCode') || params.get('code') || '')
        if (fromHash) return fromHash
      }
      const searchParams = new URLSearchParams(window.location.search)
      return extract(searchParams.get('inviteCode') || searchParams.get('code') || '')
    },

    async init() {
      this.urlCode = this.parseInviteCodeFromUrl()
      const account = await ETH.getAccount()
      const accountLac = localStorage.getItem('account')
      this.address = account

      const doLogin = async (inviteCode: string) => {
        const challenge = await fetchChallenge(account)
        if (!challenge?.message) {
          throw new Error('获取签名挑战失败')
        }
        const signature = await ETH.signMessage(challenge.message)
        const res = await apiLogin(account, signature, inviteCode)
        if (!res?.token) {
          throw new Error('登录未返回 token')
        }
        return res.token as string
      }

      const checkLogin = async () => {
        try {
          const token = await doLogin(this.urlCode || '')
          await this.loginSuccess(token)
        } catch (err: any) {
          const msg = errMsg(err, '登录失败')
          if (/邀请|invite|推荐/i.test(msg)) {
            try {
              const code = await this.inputInvitationCode()
              const token = await doLogin(String(code))
              await this.loginSuccess(token)
            } catch (e2: any) {
              showDialog({ title: '错误', message: errMsg(e2, '无效推荐码') }).then(() => checkLogin())
            }
            return
          }
          showFailToast(msg)
        }
      }

      const hasToken = !!localStorage.getItem('token')
      if (accountLac === account && hasToken) {
        try {
          await this.loginSuccess()
        } catch {
          clearAuth()
          await checkLogin()
        }
      } else {
        clearAuth()
        await checkLogin()
      }
      this.loadAccount = true
    },

    async inputInvitationCode() {
      return new Promise((resolve, reject) => {
        showDialog({
          title: '邀请码',
          message: `
            <input id="inviteInput" type="text" value="${this.urlCode || ''}"
              placeholder="请输入邀请地址"
              style="width:calc(100% - 16px);padding:8px;border:1px solid #ddd;border-radius:6px;" />
          `,
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          showCancelButton: true,
          allowHtml: true,
          zIndex: 6000,
          beforeClose(action: any) {
            if (action === 'confirm') {
              const value = (document.getElementById('inviteInput') as HTMLInputElement)?.value?.trim()
              if (!value) {
                showToast({ message: '请填写邀请码', type: 'fail' })
                return false
              }
              resolve(value)
            } else {
              reject(new Error('已取消'))
            }
            return true
          },
        } as any)
      })
    },

    applyBalanceFallback(bal: any) {
      this.profile = {
        ...this.profile,
        address: bal.address || this.address,
        usdt_recharge: bal.balance ?? this.profile.usdt_recharge,
        usdt_reward: bal.released_balance ?? this.profile.usdt_reward,
        aix_balance: bal.claimed_amount ?? bal.claimable_amount ?? this.profile.aix_balance,
        static_usdt_total: bal.static_usdt_total ?? this.profile.static_usdt_total ?? '0',
        pending_amount: bal.pending_amount ?? this.profile.pending_amount,
        unexited_amount: bal.unexited_amount ?? this.profile.unexited_amount,
        total_nodes: bal.total_nodes ?? this.profile.total_nodes,
        next_release_at: bal.next_release_at ?? this.profile.next_release_at,
        server_time: bal.server_time ?? this.profile.server_time,
      }
      if (bal.address) this.address = bal.address
    },

    async refreshProfile(throwOnError = false) {
      const tok = localStorage.getItem('token')
      if (!tok) {
        const err = new Error('未登录')
        if (throwOnError) throw err
        return
      }
      try {
        const res = await getAixProfile()
        this.profile = { ...this.profile, ...res }
        if (res.address) this.address = res.address
      } catch (err: any) {
        // 回退 balance（proto 路由走 Middleware）
        try {
          const bal = await getBalance()
          this.applyBalanceFallback(bal)
        } catch (err2) {
          console.error('[refreshProfile]', err, err2)
          if (throwOnError) throw err2 || err
        }
      }
    },

    async loginSuccess(token?: string) {
      if (token) {
        localStorage.setItem('token', token)
        localStorage.setItem('account', ETH.account || this.address)
      }
      if (!localStorage.getItem('token')) {
        throw new Error('登录 token 缺失')
      }
      await this.refreshProfile(true)
      this.isLogin = true
      if (pollTimer) clearInterval(pollTimer)
      pollTimer = setInterval(() => this.refreshProfile(false), 30000)
    },

    outLogin() {
      if (pollTimer) clearInterval(pollTimer)
      clearAuth()
      this.$reset()
      this.init()
    },
  },
})
