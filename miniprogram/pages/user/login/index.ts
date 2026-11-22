import { getLanguage, getTranslations } from '../../../utils/i18n'
import { login, register, wechatLogin } from '../../../services/api'
import { getDeviceId, saveAuth } from '../../../utils/auth'

Page({
  data: {
    t: {} as any,
    mode: 'login' as 'login' | 'register',
    username: '',
    password: '',
    nickname: '',
    submitting: false,
  },

  onLoad() {
    const lang = getLanguage()
    this.setData({ t: getTranslations(lang) })
    wx.setNavigationBarTitle({ title: lang === 'en' ? 'Sign in' : '登录' })
  },

  onModeSwitch(e: any) {
    this.setData({ mode: e.currentTarget.dataset.mode })
  },

  onUsername(e: any) {
    this.setData({ username: e.detail.value })
  },

  onPassword(e: any) {
    this.setData({ password: e.detail.value })
  },

  onNickname(e: any) {
    this.setData({ nickname: e.detail.value })
  },

  async onSubmit() {
    if (this.data.submitting) return
    const { mode, username, password, nickname } = this.data
    if (!username || !password) {
      wx.showToast({ title: this.data.t.pleaseFillForm || '请输入账号和密码', icon: 'none' })
      return
    }
    this.setData({ submitting: true })
    try {
      let resp: any
      if (mode === 'login') {
        resp = await login(username, password)
      } else {
        resp = await register(username, password, nickname)
      }
      saveAuth(resp.token, resp.user)
      wx.showToast({ title: mode === 'login' ? '登录成功' : '注册成功', icon: 'success' })
      setTimeout(() => wx.navigateBack({ fail: () => wx.switchTab({ url: '/pages/index/index' }) }), 800)
    } catch (err: any) {
      // request() already toasts the message
    } finally {
      this.setData({ submitting: false })
    }
  },

  async onWechatLogin() {
    if (this.data.submitting) return
    this.setData({ submitting: true })
    try {
      const code = await new Promise<string>((resolve, reject) => {
        wx.login({
          success: (res) => resolve(res.code),
          fail: (err) => reject(new Error(err.errMsg || 'wx.login failed')),
        })
      })
      const resp = await wechatLogin(code, getDeviceId())
      saveAuth(resp.token, resp.user)
      wx.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => wx.navigateBack({ fail: () => wx.switchTab({ url: '/pages/index/index' }) }), 800)
    } catch (err: any) {
      // request() already toasts the message
    } finally {
      this.setData({ submitting: false })
    }
  },
})
