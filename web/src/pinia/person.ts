import { defineStore } from "pinia";
import { ETH } from "@/tools/contract";
import request from "@/tools/request";
import lang from '@/i18n/index'
import fetchSign from './fetchSign'
import { getAixBalance, getAixProfile } from '@/api/aix'
import { showFailToast, showToast, showDialog, setToastDefaultOptions } from "vant";

setToastDefaultOptions({
  zIndex: 6001
})

let timeSwitch: any = null//定时获取用户信息
export default defineStore('person', {
  state: () => ({
    loadAccount: false,
    isLogin: false,
    userinfo: {
      status: 'ok',
      level: '0',
      locationNum: '0',
      total: '0',
      max: '0',
      min: '0',
      inviteUserAddress: '0x36fEa8A26AaD9Be34B29383D46FEaB42332389e6',
      buy: '0.00',
      amountGetSub: '0.00',
      amountGet: '0.0000',
      outNum: '0',
      location: '0.00',
      recommend: '0.00',
      recommendNum: 0,
      recommendTwo: '0.00',
      team: '0.00',
      all: '0.00',
      usdt: '0.00',
      withdrawRate: 0.06,
      withdrawMin: 10,
      raw: '0.00',
      withdrawRateTwo: 0,
      withdrawMinTwo: 0,
      notice: '',
      goods: [],
      one: '', // 国家
      two: '', // 省份
      three: '', // 城市
      four: '', // 区域
      five: '', // 详情
      six: '', // 收件人手机
      seven: '' // 收件人
    },
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
      aix_contract: ''
    } as Record<string, any>,
    urlCode: '',
    sign: '',
    address: '',
    inviteUserAddress: '',
    isOpened: false,
    showCodeInput: false //输入邀请码弹窗
  }),
  actions: {
    // 系统初始化
    async init() {
      // await this.inputInvitationCode()
      // this.isLogin = true;
      // return
      let url = window.location.href

      console.log(url, /invitecode/.test(url.toLowerCase()), url.replace(/^(.*)(-invitetdh-)(.*)(-invitetdh-)(.*)$/gi, '$3'))
      if (/invitecode/.test(url.toLowerCase())) {
        this.urlCode = url.replace(/^(.*)(-invitetdh-)(.*)(-invitetdh-)(.*)$/gi, '$3')
      }

      const account = await ETH.getAccount()
      const accountLac = localStorage.getItem('account')

      this.address = account
      const login = async (params: any): Promise<string> => {
        let res: any = await request.post('app_server/eth_authorize', params)
        if (res.status === '用户已锁定') throw new Error(res.status)
        if (res.status === '请输入推荐码' || res.status === '无效的推荐码') throw new Error(res.status)
        if (res.status !== 'ok' || !res.token) throw new Error(res.message || res.status || 'fail')
        return res.token
      }

      const getFreshSign = async () => {
        const sign = await fetchSign()
        this.sign = sign
        return sign
      }

      const mapAuthError = (message?: string) => {
        if (message === '用户已锁定') return lang('common.userLocked')
        if (message === '无效的推荐码') return lang('common.inviteCodeMustBeRegisteredWallet')
        if (message === '请输入推荐码') return lang('common.enterInviteCode')
        return message || lang('common.invalidInviteCode')
      }

      const checkLogin = async (sign: string, canRetryChallenge = true) => {
        try {
          // 判断是否直接进入系统
          await this.loginSuccess(await login({ address: ETH.account, code: '', sign, noMsg: true }))
        } catch (err: any) {
          if (err.message === '用户已锁定') {
            showFailToast(lang('common.userLocked'))
            return
          }
          if (canRetryChallenge && /签名挑战|challenge/i.test(err.message || '')) {
            const renewedSign = await getFreshSign()
            await checkLogin(renewedSign, false)
            return
          }
          if (err.message !== '请输入推荐码' && err.message !== '无效的推荐码') {
            showDialog({
              title: lang('common.error'),
              message: mapAuthError(err?.message),
              confirmButtonText: lang('common.confirm'),
            })
            return
          }
          // 根据推荐码进入系统
          try {
            const code = await this.inputInvitationCode()
            await this.loginSuccess(await login({ address: ETH.account, code, sign, loading: true }))
          } catch (error: any) {
            showDialog({
              title: lang('common.error'),
              message: mapAuthError(error?.message),
              confirmButtonText: lang('common.confirm'),
            }).then(() => {
              checkLogin(sign)
            })
          }
        }
      }

      /* 判断是否本地token */
      const storedToken = localStorage.getItem('token')
      const isSameAccount = Boolean(
        accountLac && account && accountLac.toLowerCase() === account.toLowerCase()
      )
      if (isSameAccount && storedToken) {
        try {
          await this.loginSuccess()
        } catch (error: any) {
          const message = error?.response?.data?.message || error?.message || ''
          const tokenInvalid = /登录过期|未登录|unauthorized|token/i.test(message)
          if (!tokenInvalid) {
            showFailToast(message || lang('common.operationFailed'))
            this.loadAccount = true
            return
          }
          // 仅在后端明确判定登录失效时，才重新请求钱包签名。
          localStorage.removeItem('token')
          localStorage.removeItem('account')
          await checkLogin(await getFreshSign())
        }
      } else {
        localStorage.removeItem('token')
        localStorage.removeItem('account')
        await checkLogin(await getFreshSign())
      }
      this.loadAccount = true
    },
    async inputInvitationCode() {
      return new Promise((resolve, reject) => {
        const urlParams = new URLSearchParams(window.location.search)
        let code = urlParams.get('code')
        code = code === null || code === 'null' ? '' : code.trim()
        const genesis = (import.meta as any).env?.VITE_GENESIS_ADDRESS || ''
        const defaultCode = this.urlCode || code || genesis

        showDialog({
          title: lang('common.inviteCode'),
          message: `
            <input
              id="inviteInput"
              type="text"
              value="${defaultCode}"
              placeholder="${lang('common.registeredWalletAddress')}"
              style="
                width: calc(100% - 16px);
                padding:8px;
                border:1px solid #ddd;
                border-radius:6px;
              "
            />
          `,
          confirmButtonText: lang('common.confirm'),
          allowHtml: true,
          zIndex: 6000,
          beforeClose(action: any) {
            if (action === 'confirm') {
              const value = (document.getElementById('inviteInput') as HTMLInputElement).value
              if (!value) {
                showToast({
                  message: lang('common.enterInviteCode'),
                  type: 'fail'
                })
                return false // ❌ 阻止关闭
              }
              resolve(value.trim())
            }
            return true
          }
        } as any)
      })
    },
    async codeError(sign: string) {
      const code = await this.inputInvitationCode()

      const login = async (params: any): Promise<string> => {
        let res: any = await request.post('app_server/eth_authorize', params)
        if (res.status === '无效的推荐码') throw new Error(res.status)
        return res.token
      }

      try {
        this.loginSuccess(await login({ address: ETH.account, code, sign, loading: true }))
      } catch (error: any) {
        // f7.dialog.alert(error?.response?.data?.message || '无效的邀请码', '错误', async () => {
        //   await this.codeError(sign)
        // });
      }
    },
    /* 获取用户信息 */
    async getUser() {
      const getData = async () => {
        let res: any = await request.get('app_server/user_info')
        this.userinfo = { ...this.userinfo, ...res }
      }
      clearInterval(timeSwitch)
      await getData()
      timeSwitch = setInterval(getData, 30000)
    },
    async refreshProfile() {
      try {
        const res: any = await getAixProfile()
        this.profile = { ...this.profile, ...res }
        return this.profile
      } catch (profileError) {
        try {
          const balance: any = await getAixBalance()
          this.profile = {
            ...this.profile,
            address: balance.address || this.address,
            usdt_recharge: balance.balance ?? this.userinfo.usdt ?? this.profile.usdt_recharge,
            usdt_reward: balance.released_balance ?? (this.userinfo as any).reward ?? this.profile.usdt_reward,
            aix_balance: balance.claimed_amount ?? balance.claimable_amount ?? this.profile.aix_balance,
            static_usdt_total: balance.static_usdt_total ?? this.profile.static_usdt_total,
            pending_amount: balance.pending_amount ?? this.profile.pending_amount,
            unexited_amount: balance.unexited_amount ?? this.profile.unexited_amount,
            total_nodes: balance.total_nodes ?? this.profile.total_nodes,
            next_release_at: balance.next_release_at ?? this.profile.next_release_at,
            server_time: balance.server_time ?? this.profile.server_time
          }
          return this.profile
        } catch (balanceError) {
          console.error('[refreshProfile]', profileError, balanceError)
          throw balanceError
        }
      }
    },
    /* 登录成功 */
    async loginSuccess(token?: string) {
      if (token) {
        localStorage.setItem('token', token)
        localStorage.setItem('account', ETH.account)
      }
      await this.getUser()
      // await this.recommend_update();
      this.isLogin = true
      location.hash = ''
      this.urlCode = ''
    },
    /* 更新推荐关系 */
    async recommend_update() {
      const sign = await fetchSign()

      if (this.urlCode && this.urlCode !== this.inviteUserAddress) {
        let res: any = await request.post('app_server/recommend_update', {
          code: this.urlCode,
          sign
        })
        this.inviteUserAddress = res.inviteUserAddress
      }
    },
    /* 自动退出系统 */
    outLogin() {
      localStorage.clear()
      this.$reset()
      this.init()
    }
  }
})
