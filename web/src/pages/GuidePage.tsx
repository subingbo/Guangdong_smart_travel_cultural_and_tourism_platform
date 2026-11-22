import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Input, Select, Spin, message } from 'antd';
import { SendOutlined, StopOutlined, PlusOutlined } from '@ant-design/icons';
import { api, ApiError } from '../api/client';
import {
  cancelRun,
  confirmPlan,
  createRun,
  getConversation,
  getRun,
  listConversations,
  newRequestId,
  streamRun,
  type AgentConversation,
  type AgentEvent,
  type AgentPlan,
} from '../api/agent';
import { useAuth } from '../context/AuthContext';
import PlanCard from '../components/PlanCard';

const CONVERSATION_KEY = 'agentConversationId';

interface ChatItem {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  plan?: AgentPlan;
  planVersion?: number;
  runId?: number;
  status?: 'streaming' | 'done' | 'error' | 'canceled';
  error?: string;
}

export default function GuidePage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuth();
  const [items, setItems] = useState<ChatItem[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [conversationId, setConversationId] = useState<number | null>(null);
  const [conversations, setConversations] = useState<AgentConversation[]>([]);
  const [toolStatus, setToolStatus] = useState('');
  const [activeRunId, setActiveRunId] = useState<number | null>(null);
  const [confirmingRunId, setConfirmingRunId] = useState<number | null>(null);
  const [confirmedRoutes, setConfirmedRoutes] = useState<Record<number, number>>({});
  const bottomRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const lastSeqRef = useRef(0);
  const stoppedRef = useRef(false);
  const loadedRef = useRef(false);
  const language = (i18n.language === 'zh' ? 'zh' : 'en') as 'zh' | 'en';

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [items, sending, toolStatus]);

  useEffect(() => {
    if (loadedRef.current || !user) return;
    loadedRef.current = true;
    void refreshConversations();
    const stored = Number(localStorage.getItem(CONVERSATION_KEY));
    if (stored > 0) {
      void loadConversation(stored).catch(() => {
        localStorage.removeItem(CONVERSATION_KEY);
      });
    }
  }, [user]);

  const refreshConversations = async () => {
    try {
      const page = await listConversations();
      setConversations(page.items || []);
    } catch {
      // The conversation list is optional; chat still works without it.
    }
  };

  const loadConversation = async (id: number) => {
    const conv = await getConversation(id);
    setConversationId(conv.id);
    localStorage.setItem(CONVERSATION_KEY, String(conv.id));
    const next: ChatItem[] = [];
    for (const m of conv.messages || []) {
      if (m.role === 'user') {
        next.push({ id: `m-${m.id}`, role: 'user', content: m.content, status: 'done' });
      } else if (m.content) {
        next.push({ id: `m-${m.id}`, role: 'assistant', content: m.content, status: 'done' });
      }
    }
    // Reconnect to a run that was still streaming when the page was closed.
    const active = (conv.messages || []).find((m) => m.status === 'streaming' && m.runId);
    if (conv.plan && !active) {
      next.push({
        id: `plan-${conv.id}-${conv.planVersion}`,
        role: 'assistant',
        content: '',
        plan: conv.plan,
        planVersion: conv.planVersion,
        status: 'done',
      });
    }
    setItems(next);

    if (active?.runId) {
      void resumeRun(active.runId);
    }
  };

  const newChat = () => {
    abortRef.current?.abort();
    stoppedRef.current = true;
    if (activeRunId) {
      void cancelRun(activeRunId).catch(() => {
        // The run may already be terminal.
      });
    }
    setItems([]);
    setConversationId(null);
    setActiveRunId(null);
    setToolStatus('');
    setSending(false);
    localStorage.removeItem(CONVERSATION_KEY);
  };

  const patchItem = (id: string, updater: (item: ChatItem) => ChatItem) => {
    setItems((prev) => prev.map((item) => (item.id === id ? updater(item) : item)));
  };

  const finishStreaming = (assistantId: string) => {
    patchItem(assistantId, (item) =>
      item.status === 'streaming'
        ? { ...item, status: stoppedRef.current ? 'canceled' : 'done' }
        : item,
    );
    setSending(false);
    setActiveRunId(null);
    setToolStatus('');
    abortRef.current = null;
  };

  const send = async () => {
    const question = input.trim();
    if (!question || sending) return;
    if (!user) {
      message.warning(t('profile.loginToView'));
      return;
    }
    setInput('');
    const assistantId = newRequestId();
    setItems((prev) => [
      ...prev,
      { id: newRequestId(), role: 'user', content: question, status: 'done' },
      { id: assistantId, role: 'assistant', content: '', status: 'streaming' },
    ]);
    setSending(true);
    setToolStatus(t('guide.thinking'));
    stoppedRef.current = false;
    const controller = new AbortController();
    abortRef.current = controller;

    try {
      const snapshot = await createRun({
        conversationId: conversationId ?? undefined,
        message: question,
        clientRequestId: newRequestId(),
      });
      setConversationId(snapshot.conversationId);
      localStorage.setItem(CONVERSATION_KEY, String(snapshot.conversationId));
      setActiveRunId(snapshot.runId);
      patchItem(assistantId, (item) => ({
        ...item,
        runId: snapshot.runId,
        content: snapshot.output || item.content,
        plan: snapshot.plan || item.plan,
        planVersion: snapshot.planVersion || item.planVersion,
      }));

      if (stoppedRef.current) {
        await cancelRun(snapshot.runId);
        return;
      }

      lastSeqRef.current = 0;
      await streamRun(
        snapshot.runId,
        (event) => handleEvent(assistantId, event),
        controller.signal,
      );
    } catch (e) {
      if ((e as Error).name === 'AbortError') {
        patchItem(assistantId, (item) => ({ ...item, status: 'canceled' }));
      } else if (e instanceof ApiError && e.status === 503) {
        await fallbackLegacy(assistantId, question);
      } else {
        const text = (e as Error).message || t('guide.failed');
        patchItem(assistantId, (item) => ({
          ...item,
          status: 'error',
          error: text,
          content: item.content || '',
        }));
        message.error(text);
      }
    } finally {
      finishStreaming(assistantId);
      void refreshConversations();
    }
  };

  const resumeRun = async (runId: number) => {
    if (sending) return;
    try {
      const snapshot = await getRun(runId, false);
      if (!['queued', 'running', 'canceling'].includes(snapshot.status)) {
        return;
      }
      const assistantId = `resume-${runId}`;
      setItems((prev) => {
        if (prev.some((item) => item.id === assistantId)) return prev;
        return [
          ...prev,
          {
            id: assistantId,
            role: 'assistant',
            content: snapshot.output,
            plan: snapshot.plan,
            planVersion: snapshot.planVersion,
            runId,
            status: 'streaming',
          },
        ];
      });
      setSending(true);
      setActiveRunId(runId);
      setToolStatus(t('guide.thinking'));
      stoppedRef.current = false;
      const controller = new AbortController();
      abortRef.current = controller;
      lastSeqRef.current = 0;
      try {
        await streamRun(runId, (event) => handleEvent(assistantId, event), controller.signal);
      } catch (e) {
        if ((e as Error).name !== 'AbortError') {
          patchItem(assistantId, (item) => ({
            ...item,
            status: 'error',
            error: (e as Error).message || t('guide.failed'),
          }));
        }
      } finally {
        finishStreaming(assistantId);
      }
    } catch {
      // The run finished while the page was loading.
    }
  };

  const handleEvent = (assistantId: string, event: AgentEvent) => {
    if (event.seq > 0 && event.seq <= lastSeqRef.current) return;
    if (event.seq > lastSeqRef.current) lastSeqRef.current = event.seq;
    const data = event.data as {
      content?: string;
      name?: string;
      ok?: boolean;
      plan?: AgentPlan;
      version?: number;
      message?: string;
      output?: string;
      status?: string;
      planVersion?: number;
      error?: { message?: string };
    };

    switch (event.type) {
      case 'snapshot': {
        const terminal =
          !!data.status &&
          ['succeeded', 'failed', 'canceled', 'interrupted'].includes(data.status);
        patchItem(assistantId, (item) => ({
          ...item,
          content: data.output || item.content,
          plan: data.plan || item.plan,
          planVersion: data.planVersion || item.planVersion,
          status: terminal
            ? data.status === 'succeeded'
              ? 'done'
              : 'error'
            : item.status,
          error:
            terminal && data.status !== 'succeeded'
              ? data.error?.message || t('guide.failed')
              : item.error,
        }));
        break;
      }
      case 'delta':
        if (data.content) {
          setToolStatus('');
          patchItem(assistantId, (item) => ({ ...item, content: item.content + data.content }));
        }
        break;
      case 'thinking':
        setToolStatus(t('guide.thinking'));
        break;
      case 'tool_call':
        setToolStatus(`${t('guide.querying')} ${data.name || ''}`);
        break;
      case 'tool_result':
        setToolStatus(data.ok ? t('guide.toolDone') : t('guide.toolFailed'));
        break;
      case 'plan':
        if (data.plan) {
          patchItem(assistantId, (item) => ({
            ...item,
            plan: data.plan,
            planVersion: data.version || data.plan?.version,
          }));
        }
        break;
      case 'done':
        setToolStatus('');
        patchItem(assistantId, (item) => ({ ...item, status: 'done' }));
        break;
      case 'error':
        setToolStatus('');
        patchItem(assistantId, (item) => ({
          ...item,
          status: 'error',
          error: data.message || t('guide.failed'),
        }));
        break;
      default:
        break;
    }
  };

  const fallbackLegacy = async (assistantId: string, question: string) => {
    try {
      const resp = await api.post<{ sessionId: number; reply: string }>('/api/guide/chat', {
        sessionId: conversationId ?? undefined,
        question,
      });
      setConversationId(resp.sessionId);
      localStorage.setItem(CONVERSATION_KEY, String(resp.sessionId));
      patchItem(assistantId, (item) => ({ ...item, content: resp.reply, status: 'done' }));
    } catch (e) {
      const text = (e as Error).message || t('guide.failed');
      patchItem(assistantId, (item) => ({ ...item, status: 'error', error: text }));
      message.error(text);
    }
  };

  const stop = async () => {
    stoppedRef.current = true;
    abortRef.current?.abort();
    if (activeRunId) {
      try {
        await cancelRun(activeRunId);
      } catch {
        // The stream may already be closed.
      }
    }
    setSending(false);
    setToolStatus('');
  };

  const confirm = async (item: ChatItem) => {
    if (!item.runId || !item.plan) return;
    setConfirmingRunId(item.runId);
    try {
      const result = await confirmPlan(item.runId, item.plan.version);
      setConfirmedRoutes((prev) => ({ ...prev, [item.runId as number]: result.routeId }));
      message.success(t('guide.planConfirmed'));
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setConfirmingRunId(null);
    }
  };

  return (
    <div className="guide-shell">
      <div className="section-title">{t('guide.title')}</div>
      <div className="agent-toolbar">
        <Select
          size="small"
          allowClear
          placeholder={t('guide.history')}
          value={conversationId ?? undefined}
          onChange={(value) => {
            if (value) void loadConversation(value);
          }}
          onClear={newChat}
          options={conversations.map((c) => ({ value: c.id, label: c.title }))}
          style={{ minWidth: 220, flex: 1 }}
        />
        <Button size="small" icon={<PlusOutlined />} onClick={newChat}>
          {t('guide.newChat')}
        </Button>
      </div>

      <div className="chat-box">
        <div className="bubble assistant">{t('guide.welcome')}</div>
        {items.map((item) => (
          <div key={item.id} className={`agent-message ${item.role}`}>
            {item.content && <div className={`bubble ${item.role}`}>{item.content}</div>}
            {item.role === 'assistant' && item.status === 'streaming' && (
              <div className="agent-status">
                <Spin size="small" /> {toolStatus || t('guide.thinking')}
              </div>
            )}
            {item.error && <div className="agent-error">{item.error}</div>}
            {item.plan && (
              <PlanCard
                plan={item.plan}
                language={language}
                confirming={confirmingRunId === item.runId}
                confirmedRouteId={item.runId ? confirmedRoutes[item.runId] : undefined}
                onConfirm={() => void confirm(item)}
              />
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div className="chat-input-bar">
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onPressEnter={() => void send()}
          placeholder={t('guide.placeholder')}
        />
        {sending ? (
          <Button danger icon={<StopOutlined />} onClick={() => void stop()}>
            {t('guide.stop')}
          </Button>
        ) : (
          <Button type="primary" icon={<SendOutlined />} onClick={() => void send()} />
        )}
      </div>
    </div>
  );
}
