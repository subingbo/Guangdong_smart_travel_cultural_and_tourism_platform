// Agent API: run lifecycle, SSE streaming and plan confirmation.
import { api, API_BASE, ApiError, getToken, type Coordinate, type Page } from './client';

export type AgentRunStatus =
  | 'queued'
  | 'running'
  | 'canceling'
  | 'succeeded'
  | 'failed'
  | 'canceled'
  | 'interrupted';

export interface AgentUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

export interface AgentPlanStop {
  placeId: number;
  name: string;
  nameEn?: string;
  lat: number;
  lng: number;
  source: string;
  arrivalTime: string;
  leaveTime: string;
  stayMinutes: number;
  transportMode: string;
}

export interface AgentPlanSegment {
  fromIndex: number;
  toIndex: number;
  transportMode: string;
  distanceMeters: number;
  durationMinutes: number;
  estimated: boolean;
  polyline?: Coordinate[];
}

export interface AgentPlan {
  version: number;
  plannerVersion: string;
  title: string;
  summary?: string;
  windowStart: string;
  windowEnd: string;
  transportMode: string;
  routeConfidence: 'AMAP' | 'ESTIMATED' | 'NONE';
  stops: AgentPlanStop[];
  segments: AgentPlanSegment[];
  totalTravelMinutes: number;
  totalDistanceMeters: number;
  score: number;
  suggestions?: string[];
  createdAt: string;
  status: 'draft' | 'confirmed';
}

export interface AgentStep {
  id: number;
  runId: number;
  stepNo: number;
  type: string;
  name: string;
  arguments?: string;
  result?: string;
  latencyMs: number;
  error?: string;
  createdAt: string;
}

export interface AgentSnapshot {
  runId: number;
  conversationId: number;
  status: AgentRunStatus;
  output: string;
  plan?: AgentPlan;
  planVersion: number;
  stepCount: number;
  usage: AgentUsage;
  error?: { code: string; message: string };
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  seq: number;
  steps?: AgentStep[];
}

export interface AgentMessage {
  id: number;
  sessionId: number;
  runId?: number;
  role: 'user' | 'assistant';
  content: string;
  status?: string;
  createdAt: string;
}

export interface AgentConversation {
  id: number;
  title: string;
  plan?: AgentPlan;
  planVersion: number;
  lastRunId?: number;
  messages: AgentMessage[];
  createdAt: string;
}

export interface AgentEvent {
  type: string;
  seq: number;
  runId?: number;
  data: unknown;
}

export interface ConfirmPlanResult {
  routeId: number;
  routeName: string;
  totalDistance: string;
  routeConfidence: string;
}

export function newRequestId(): string {
  const cryptoObj = typeof globalThis !== 'undefined' ? globalThis.crypto : undefined;
  if (cryptoObj && typeof cryptoObj.randomUUID === 'function') {
    return cryptoObj.randomUUID();
  }
  if (cryptoObj && typeof cryptoObj.getRandomValues === 'function') {
    const bytes = cryptoObj.getRandomValues(new Uint8Array(16));
    const hex = Array.from(bytes as Uint8Array, (b: number) => b.toString(16).padStart(2, '0')).join('');
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
  }
  return `req-${Date.now().toString(36)}`;
}

export function createRun(body: {
  conversationId?: number;
  message: string;
  clientRequestId?: string;
}): Promise<AgentSnapshot> {
  return api.post<AgentSnapshot>('/api/agent/runs', body);
}

export function getRun(runId: number, includeSteps = true): Promise<AgentSnapshot> {
  return api.get<AgentSnapshot>(`/api/agent/runs/${runId}${includeSteps ? '' : '?steps=0'}`);
}

export function cancelRun(runId: number): Promise<AgentSnapshot> {
  return api.post<AgentSnapshot>(`/api/agent/runs/${runId}/cancel`);
}

export function confirmPlan(
  runId: number,
  version: number,
  idempotencyKey = newRequestId(),
): Promise<ConfirmPlanResult> {
  return api.post<ConfirmPlanResult>(`/api/agent/runs/${runId}/plan/confirm`, {
    version,
    idempotencyKey,
  });
}

export function listConversations(page = 1, pageSize = 20): Promise<Page<AgentConversation>> {
  return api.get<Page<AgentConversation>>(
    `/api/agent/conversations?page=${page}&pageSize=${pageSize}`,
  );
}

export function getConversation(id: number): Promise<AgentConversation> {
  return api.get<AgentConversation>(`/api/agent/conversations/${id}`);
}

// Streams a run over fetch; EventSource cannot send the bearer token.
export async function streamRun(
  runId: number,
  onEvent: (event: AgentEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const token = getToken();
  const res = await fetch(`${API_BASE}/api/agent/runs/${runId}/events`, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    signal,
  });
  if (!res.ok) {
    let message = `Stream failed (${res.status})`;
    try {
      const body = (await res.json()) as { message?: string };
      if (body.message) message = body.message;
    } catch {
      // keep the status message
    }
    throw new ApiError(res.status, message);
  }
  if (!res.body) {
    throw new Error('Streaming is not supported by this browser');
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    buffer = buffer.replace(/\r\n/g, '\n');
    let boundary = buffer.indexOf('\n\n');
    while (boundary >= 0) {
      const frame = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);
      const event = parseFrame(frame);
      if (event) onEvent(event);
      boundary = buffer.indexOf('\n\n');
    }
  }
  if (buffer.trim()) {
    const event = parseFrame(buffer);
    if (event) onEvent(event);
  }
}

function parseFrame(frame: string): AgentEvent | null {
  let eventName = '';
  let data = '';
  let id = 0;
  for (const rawLine of frame.split('\n')) {
    const line = rawLine.replace(/\r$/, '');
    if (!line || line.startsWith(':')) continue;
    if (line.startsWith('event:')) eventName = line.slice(6).trim();
    else if (line.startsWith('data:')) data += line.slice(5).trim();
    else if (line.startsWith('id:')) id = Number(line.slice(3).trim()) || 0;
  }
  if (!data) return null;
  try {
    const parsed = JSON.parse(data) as AgentEvent;
    return {
      type: parsed.type || eventName || 'message',
      seq: parsed.seq ?? id,
      runId: parsed.runId,
      data: parsed.data,
    };
  } catch {
    return null;
  }
}
