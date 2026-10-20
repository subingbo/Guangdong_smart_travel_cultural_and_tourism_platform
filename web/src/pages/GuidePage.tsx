import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Input, Spin, message } from 'antd';
import { SendOutlined } from '@ant-design/icons';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface ChatItem {
  role: 'user' | 'assistant';
  content: string;
}

export default function GuidePage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [items, setItems] = useState<ChatItem[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [sessionId, setSessionId] = useState<number | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [items, sending]);

  const send = async () => {
    const question = input.trim();
    if (!question || sending) return;
    if (!user) {
      message.warning(t('profile.loginToView'));
      return;
    }
    setInput('');
    setItems((prev) => [...prev, { role: 'user', content: question }]);
    setSending(true);
    try {
      const resp = await api.post<{ sessionId: number; reply: string }>('/api/guide/chat', {
        sessionId: sessionId ?? undefined,
        question,
      });
      setSessionId(resp.sessionId);
      setItems((prev) => [...prev, { role: 'assistant', content: resp.reply }]);
    } catch (e) {
      message.error((e as Error).message);
      setItems((prev) => prev.slice(0, -1));
      setInput(question);
    } finally {
      setSending(false);
    }
  };

  return (
    <div>
      <div className="section-title">{t('guide.title')}</div>
      <div className="chat-box">
        <div className="bubble assistant">{t('guide.welcome')}</div>
        {items.map((m, i) => (
          <div key={i} className={`bubble ${m.role}`}>
            {m.content}
          </div>
        ))}
        {sending && (
          <div className="bubble assistant" style={{ padding: '14px 16px' }}>
            <Spin size="small" />
          </div>
        )}
        <div ref={bottomRef} />
      </div>
      <div style={{ display: 'flex', gap: 8, position: 'sticky', bottom: 70, background: '#f5f6f8', padding: '8px 0' }}>
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onPressEnter={send}
          placeholder={t('guide.placeholder')}
        />
        <Button type="primary" icon={<SendOutlined />} onClick={send} loading={sending} />
      </div>
    </div>
  );
}
