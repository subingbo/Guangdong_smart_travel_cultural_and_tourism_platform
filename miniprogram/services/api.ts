import { request } from '../utils/request'
import { uploadImage } from '../utils/upload'

// --- shared types ----------------------------------------------------------

export interface Scenic {
  id: number
  imageUrl: string
  titleCn: string
  titleEn: string
  location: string
  tag: string
  descriptionCn: string
  descriptionEn: string
}

export interface Place {
  id: number
  nameCn: string
  nameEn: string
  lat: number
  lng: number
  tags: string[]
  indoor: boolean
  duration: number
  descriptionCn: string
  descriptionEn: string
}

export interface Waypoint {
  name: string
  nameEn: string
  lat: number
  lng: number
  stayTime: string
  description: string
}

export interface Route {
  id: number
  routeName: string
  durationDays: number
  totalDistance: string
  mode: string
  pace: string
  interests: string[]
  waypoints: Waypoint[]
  polyline: { lat: number; lng: number }[]
}

export interface BlogAuthor {
  id: number
  nickname: string
  avatarUrl: string
}

export interface Blog {
  id: number
  userId: number
  author: BlogAuthor | null
  title: string
  content: string
  images: string[]
  likes: number
  likedByMe: boolean
  commentCount: number
  createdAt: string
}

export interface UserProfile {
  user: {
    id: number
    username: string
    nickname: string
    avatarUrl: string
    createdAt: string
  }
  stats: {
    posts: number
    followers: number
    following: number
    likesReceived: number
  }
  isFollowing: boolean
}

// --- auth ------------------------------------------------------------------

export interface AuthResponse {
  token: string
  user: {
    id: number
    username: string
    nickname: string
    avatarUrl: string
    language: string
  }
}

export function login(username: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>({ url: '/api/auth/login', method: 'POST', data: { username, password } })
}

export function register(username: string, password: string, nickname: string): Promise<AuthResponse> {
  return request<AuthResponse>({ url: '/api/auth/register', method: 'POST', data: { username, password, nickname } })
}

export function wechatLogin(code: string, deviceId: string): Promise<AuthResponse> {
  return request<AuthResponse>({
    url: '/api/auth/wechat-login',
    method: 'POST',
    data: { code, deviceId },
  })
}

// --- catalog ----------------------------------------------------------------

export function listScenics(): Promise<Scenic[]> {
  return request<Scenic[]>({ url: '/api/scenics' })
}

export function listPlaces(tag = ''): Promise<Place[]> {
  return request<Place[]>({ url: '/api/places' + (tag ? '?tag=' + encodeURIComponent(tag) : '') })
}

// --- routes -----------------------------------------------------------------

export interface RouteGenRequest {
  days: number
  interests: string[]
  pace: string
  mode: string
}

export function generateRoute(req: RouteGenRequest): Promise<Route> {
  return request<Route>({ url: '/api/routes/generate', method: 'POST', data: req })
}

export function getRoute(id: number | string): Promise<Route> {
  return request<Route>({ url: '/api/routes/' + id })
}

export function listRoutes(): Promise<Route[]> {
  return request<Route[]>({ url: '/api/routes' })
}

// --- blogs ------------------------------------------------------------------

export interface BlogFeed {
  items: Blog[]
  total: number
  page: number
  pageSize: number
}

export function listBlogs(page = 1, pageSize = 10, q = ''): Promise<BlogFeed> {
  const query = '?page=' + page + '&pageSize=' + pageSize + (q ? '&q=' + encodeURIComponent(q) : '')
  return request<BlogFeed>({ url: '/api/blogs' + query })
}

export function myBlogs(page = 1, pageSize = 20): Promise<BlogFeed> {
  return request<BlogFeed>({ url: '/api/me/blogs?page=' + page + '&pageSize=' + pageSize })
}

export function draftBlog(topic: string): Promise<{ title: string; content: string }> {
  return request<{ title: string; content: string }>({ url: '/api/blogs/draft', method: 'POST', data: { topic } })
}

export function publishBlog(title: string, content: string, images: string[]): Promise<Blog> {
  return request<Blog>({ url: '/api/blogs', method: 'POST', data: { title, content, images } })
}

export function likeBlog(id: number): Promise<any> {
  return request({ url: '/api/blogs/' + id + '/like', method: 'POST' })
}

export function unlikeBlog(id: number): Promise<any> {
  return request({ url: '/api/blogs/' + id + '/like', method: 'DELETE' })
}

// --- guide ------------------------------------------------------------------

export function guideChat(sessionId: number, question: string): Promise<{ sessionId: number; reply: string }> {
  return request({ url: '/api/guide/chat', method: 'POST', data: { sessionId, question } })
}

export function guideSession(id: number): Promise<any> {
  return request({ url: '/api/guide/sessions/' + id })
}

// --- me / favorites ---------------------------------------------------------

export function getMe(): Promise<UserProfile> {
  return request<UserProfile>({ url: '/api/me' })
}

export function updateMe(data: { nickname?: string; avatarUrl?: string; language?: string }): Promise<UserProfile> {
  return request<UserProfile>({ url: '/api/me', method: 'PUT', data })
}

export function listFavorites(): Promise<Place[]> {
  return request<Place[]>({ url: '/api/favorites' })
}

export function getMyStats(): Promise<{ posts: number; followers: number; following: number; likesReceived: number }> {
  return request({ url: '/api/me/stats' })
}

// --- merchant AI ------------------------------------------------------------

export interface CopyResponse { content: string }

export function aiCopy(topic: string, type: string, tone: string, language: string): Promise<CopyResponse> {
  return request<CopyResponse>({ url: '/api/ai/copy', method: 'POST', data: { topic, type, tone, language } })
}

export interface PosterTask {
  taskId: string
  status: 'pending' | 'running' | 'succeeded' | 'failed'
  imageUrl?: string
  message?: string
}

export function createPoster(prompt: string, style: string, language: string): Promise<{ taskId: string }> {
  return request({ url: '/api/ai/poster', method: 'POST', data: { prompt, style, language } })
}

export function getPosterTask(taskId: string): Promise<PosterTask> {
  return request<PosterTask>({ url: '/api/ai/poster/' + taskId })
}

export interface TranslatedDish {
  name: string
  translation: string
  notes?: string
}

export function menuOCR(image: string, language: string): Promise<{ text: string }> {
  return request({ url: '/api/ai/menu-ocr', method: 'POST', data: { image, language } })
}

export function menuTranslate(image: string, language: string): Promise<{ items: TranslatedDish[] }> {
  return request({ url: '/api/ai/menu-translate', method: 'POST', data: { image, language } })
}

export function dishExplain(name: string, kind: 'card' | 'explain', language: string): Promise<{ title: string; content: string; tags?: string[] }> {
  return request({ url: '/api/ai/dish-explain', method: 'POST', data: { name, kind, language } })
}

// --- upload -----------------------------------------------------------------

export function uploadBlogImage(filePath: string): Promise<string> {
  return uploadImage(filePath, 'blog')
}

export function uploadAvatar(filePath: string): Promise<string> {
  return uploadImage(filePath, 'avatar')
}

export { request } from '../utils/request'
export { uploadImage } from '../utils/upload'
