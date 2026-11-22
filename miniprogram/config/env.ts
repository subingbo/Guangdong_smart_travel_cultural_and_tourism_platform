// Local development backend base URL. For production, replace with the
// deployed HTTPS domain (must be whitelisted in the WeChat MP console).
export const BASE_URL = 'http://127.0.0.1:8081'

/** Prefix for image paths returned by the backend (e.g. /uploads/...). */
export function fileUrl(path: string): string {
  if (!path) return ''
  if (/^https?:\/\//.test(path)) return path
  return BASE_URL + path
}
