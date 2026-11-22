const TOKEN_KEY = 'auth_token'
const USER_KEY = 'auth_user'
const DEVICE_KEY = 'auth_device_id'

/** Backend user shape (subset the mini-program cares about). */
export interface AuthUser {
  id: number
  username: string
  nickname: string
  avatarUrl: string
  language: string
}

export function getToken(): string {
  return wx.getStorageSync(TOKEN_KEY) || ''
}

export function getUser(): AuthUser | null {
  return wx.getStorageSync(USER_KEY) || null
}

export function saveAuth(token: string, user: AuthUser) {
  wx.setStorageSync(TOKEN_KEY, token)
  wx.setStorageSync(USER_KEY, user)
}

export function clearAuth() {
  wx.removeStorageSync(TOKEN_KEY)
  wx.removeStorageSync(USER_KEY)
}

/** Stable per-install device id used by the backend's dev-mode wechat login. */
export function getDeviceId(): string {
  let id = wx.getStorageSync(DEVICE_KEY)
  if (!id) {
    id = 'mp-' + Date.now().toString(36) + '-' + Math.floor(Math.random() * 1e8).toString(36)
    wx.setStorageSync(DEVICE_KEY, id)
  }
  return id
}
