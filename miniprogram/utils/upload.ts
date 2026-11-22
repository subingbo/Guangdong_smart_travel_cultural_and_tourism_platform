import { BASE_URL } from '../config/env'
import { getToken } from './auth'

/**
 * Upload one image file with wx.uploadFile (multipart) and return the backend
 * URL of the stored file (e.g. /uploads/blog/2026/11/xxx.jpg).
 */
export function uploadImage(filePath: string, kind: 'blog' | 'avatar' = 'blog'): Promise<string> {
  const header: Record<string, string> = {}
  const token = getToken()
  if (token) {
    header.Authorization = 'Bearer ' + token
  }
  return new Promise<string>((resolve, reject) => {
    wx.uploadFile({
      url: BASE_URL + '/api/upload?kind=' + kind,
      filePath,
      name: 'file',
      header,
      timeout: 60000,
      success: async (res) => {
        if (res.statusCode === 401) {
          reject(new Error('请先登录'))
          return
        }
        let body: any = null
        try {
          body = JSON.parse(res.data)
        } catch (err) {
          // fall through to the generic error below
        }
        if (res.statusCode >= 200 && res.statusCode < 300 && body && body.success) {
          resolve(body.data.url as string)
          return
        }
        const message = (body && body.message) || '上传失败 (' + res.statusCode + ')'
        wx.showToast({ title: message, icon: 'none' })
        reject(new Error(message))
      },
      fail: (err) => {
        const message = '上传失败：' + (err.errMsg || 'network error')
        wx.showToast({ title: message, icon: 'none', duration: 3000 })
        reject(new Error(message))
      },
    })
  })
}

/** Read a local image as a compressed base64 payload for AI vision calls. */
export function readImageBase64(filePath: string, quality = 60): Promise<string> {
  return new Promise((resolve, reject) => {
    wx.compressImage({
      src: filePath,
      quality,
      success: (res) => {
        const fs = wx.getFileSystemManager()
        fs.readFile({
          filePath: res.tempFilePath,
          encoding: 'base64',
          success: (data) => resolve(data.data as string),
          fail: (err) => reject(new Error(err.errMsg || 'read file failed')),
        })
      },
      fail: () => {
        // compressImage can fail on some formats; retry with the original.
        const fs = wx.getFileSystemManager()
        fs.readFile({
          filePath,
          encoding: 'base64',
          success: (data) => resolve(data.data as string),
          fail: (err) => reject(new Error(err.errMsg || 'read file failed')),
        })
      },
    })
  })
}
