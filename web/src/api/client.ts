/** Typed API client: unified envelope + JWT header, per docs/api/openapi.yaml. */

export const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

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
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body === undefined ? undefined : JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  del: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};

// ---------- API model types ----------

export interface User {
  id: number;
  username: string;
  nickname: string;
  avatarUrl: string;
  language: 'en' | 'zh';
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

export interface Blog {
  id: number;
  userId: number;
  title: string;
  content: string;
  images: string[];
  likes: number;
  likedByMe: boolean;
  createdAt: string;
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
