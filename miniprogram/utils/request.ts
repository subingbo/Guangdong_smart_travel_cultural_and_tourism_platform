import { BASE_URL } from '../config/env'
import { clearAuth, getToken } from './auth'

const LOGIN_PAGE = '/pages/user/login/index'

interface Envelope<T> {
  success: boolean
  data: T
  message: string
}

export interface ApiError extends Error {
  statusCode: number
}

export interface RequestOptions<T = any> {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: T
  /** Skip the global error toast (the page shows its own). */
  silent?: boolean
}

/**
 * Core request helper. Wraps wx.request as a Promise, injects the bearer
 * token, unwraps the {success, data, message} envelope and toasts failures.
 */
export function request<T = any>(options: RequestOptions): Promise<T> {
  const header: Record<string, string> = { 'content-type': 'application/json' }
  const token = getToken()
  if (token) {
    header.Authorization = 'Bearer ' + token
  }
  return new Promise<T>((resolve, reject) => {
    wx.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header,
      timeout: 60000,
      success: (res) => {
        const statusCode = res.statusCode || 0
        const body = (res.data || {}) as Envelope<T>
        if (statusCode === 401) {
          clearAuth()
          wx.showToast({ title: '请先登录', icon: 'none' })
          // Avoid stacking login pages when the 401 came from the login page.
          const pages = getCurrentPages()
          const current = pages[pages.length - 1] as any
          if (!current || current.route !== 'pages/user/login/index') {
            wx.navigateTo({ url: LOGIN_PAGE })
          }
          reject(makeError('unauthorized', 401))
          return
        }
        if (statusCode < 200 || statusCode >= 300 || !body.success) {
          const message = body.message || '请求失败 (' + statusCode + ')'
          if (!options.silent) {
            wx.showToast({ title: message, icon: 'none' })
          }
          reject(makeError(message, statusCode))
          return
        }
        resolve(body.data)
      },
      fail: (err) => {
        const message = '网络错误，请确认后端已启动 (' + (err.errMsg || 'request:fail') + ')'
        if (!options.silent) {
          wx.showToast({ title: message, icon: 'none', duration: 3000 })
        }
        reject(makeError(message, 0))
      },
    })
  })
}

function makeError(message: string, statusCode: number): ApiError {
  const err = new Error(message) as ApiError
  err.statusCode = statusCode
  return err
}

export function gotoLogin() {
  wx.navigateTo({ url: LOGIN_PAGE })
}
