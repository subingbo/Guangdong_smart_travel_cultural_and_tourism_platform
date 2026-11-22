// Typed API client: unified envelope + JWT header, per docs/api/openapi.yaml.

export const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

// Uploads are stored as portable relative paths (e.g. /uploads/blog/...).
// In dev the Vite proxy serves them; when the API lives on another origin,
// VITE_API_BASE_URL / VITE_ASSET_BASE_URL prefixes them for the browser.
export const ASSET_BASE = import.meta.env.VITE_ASSET_BASE_URL || API_BASE;
export const FALLBACK_IMAGE = '/placeholder-cover.svg';

export function assetUrl(path?: string | null): string {
  if (!path) return '';
  if (/^(?:https?:)?\/\//i.test(path) || /^(?:data|blob):/i.test(path)) return path;
  const base = ASSET_BASE.replace(/\/$/, '');
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return base ? `${base}${normalized}` : normalized;
}

const TOKEN_KEY = 'token';
const LANG_KEY = 'language';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

export function getLanguage(): 'en' | 'zh' {
  return (localStorage.getItem(LANG_KEY) as 'en' | 'zh') || 'en';
}

export function setLanguage(lang: 'en' | 'zh') {
  localStorage.setItem(LANG_KEY, lang);
}

interface Envelope<T> {
  success: boolean;
  data: T;
  message: string;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  };
  const token = getToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const body = (await res.json()) as Envelope<T>;
  if (!res.ok || !body.success) {
    throw new ApiError(res.status, body.message || `Request failed (${res.status})`);
  }
  return body.data;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown, headers?: Record<string, string>) =>
    request<T>(path, {
      method: 'POST',
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  del: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};

// Uploads one image and returns its public path (e.g. /uploads/blog/2026/11/x.jpg).
export async function uploadImage(file: File, kind: 'blog' | 'avatar' = 'blog'): Promise<string> {
  const form = new FormData();
  form.append('file', file);
  const token = getToken();
  const res = await fetch(`${API_BASE}/api/upload?kind=${kind}`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    body: form,
  });
  const body = (await res.json()) as Envelope<{ url: string }>;
  if (!res.ok || !body.success) {
    throw new ApiError(res.status, body.message || 'Upload failed');
  }
  return body.data.url;
}

// API model types

export interface User {
  id: number;
  username: string;
  nickname: string;
  avatarUrl: string;
  language: 'en' | 'zh';
  role: 'user' | 'merchant' | 'admin';
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ScenicImage {
  id: number;
  imageUrl: string;
  titleCn: string;
  titleEn: string;
  location: string;
  tag: string;
  descriptionCn: string;
  descriptionEn: string;
}

export interface Place {
  id: number;
  nameCn: string;
  nameEn: string;
  lat: number;
  lng: number;
  tags: string[];
  indoor: boolean;
  duration: number;
  descriptionCn: string;
  descriptionEn: string;
}

export interface Author {
  id: number;
  nickname: string;
  avatarUrl: string;
}

export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface Blog {
  id: number;
  userId: number;
  author?: Author;
  title: string;
  content: string;
  images: string[];
  likes: number;
  likedByMe: boolean;
  commentCount: number;
  createdAt: string;
  updatedAt?: string;
}

export interface Comment {
  id: number;
  blogId: number;
  userId: number;
  author?: Author;
  content: string;
  createdAt: string;
}

export interface UserStats {
  posts: number;
  followers: number;
  following: number;
  likesReceived: number;
}

export interface PublicUser {
  id: number;
  username: string;
  nickname: string;
  avatarUrl: string;
  createdAt: string;
}

export interface UserProfile {
  user: PublicUser;
  stats: UserStats;
  isFollowing: boolean;
}

export interface NotificationItem {
  id: number;
  type: 'like' | 'comment' | 'follow';
  isRead: boolean;
  createdAt: string;
  actor?: Author;
  blogId?: number;
  blogTitle?: string;
  commentContent?: string;
}

export interface NotificationPage extends Page<NotificationItem> {
  unread: number;
}

export interface MyStats {
  favorites: number;
  routes: number;
  posts: number;
  followers: number;
  following: number;
  likesReceived: number;
  unreadNotifications: number;
}

export interface Waypoint {
  name: string;
  nameEn: string;
  lat: number;
  lng: number;
  stayTime: string;
  description: string;
}

export interface Coordinate {
  lat: number;
  lng: number;
}

export interface Route {
  id: number;
  routeName: string;
  durationDays: number;
  totalDistance: string;
  mode: string;
  pace: string;
  interests: string[];
  waypoints: Waypoint[];
  polyline: Coordinate[];
  createdAt: string;
}

export interface GuideMessage {
  id: number;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

export interface GuideSession {
  id: number;
  title: string;
  messages: GuideMessage[];
}

// Shop types

export interface Product {
  id: number;
  nameCn: string;
  nameEn: string;
  descriptionCn: string;
  descriptionEn: string;
  priceCents: number;
  category: string;
  imageUrl: string;
  stock: number;
  soldCount: number;
  rating: number;
  reviewCount: number;
  active: boolean;
}

export interface ProductReview {
  id: number;
  productId: number;
  userId: number;
  rating: number;
  content: string;
  createdAt: string;
  author?: Author;
}

export interface Activity {
  id: number;
  titleCn: string;
  titleEn: string;
  descriptionCn: string;
  descriptionEn: string;
  location: string;
  startAt: string;
  endAt: string;
  imageUrl: string;
  organizer: string;
  tag: string;
  status: string;
}

export interface CartItem {
  id: number;
  productId: number;
  quantity: number;
  nameCn: string;
  nameEn: string;
  priceCents: number;
  imageUrl: string;
  stock: number;
  category: string;
}

export interface CartView {
  items: CartItem[];
  totalCents: number;
  count: number;
}

export interface OrderItem {
  productId: number;
  nameCn: string;
  nameEn: string;
  priceCents: number;
  quantity: number;
  imageUrl: string;
}

export interface Order {
  id: number;
  orderNo: string;
  totalCents: number;
  status: 'pending' | 'paid' | 'completed' | 'canceled';
  createdAt: string;
  updatedAt: string;
  items: OrderItem[];
}

export interface Knowledge {
  id: number;
  titleCn: string;
  titleEn: string;
  contentCn: string;
  contentEn: string;
  tags: string[];
}
