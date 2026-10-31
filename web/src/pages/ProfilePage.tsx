import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Avatar, Button, Card, Empty, Modal, Statistic, Tabs, message } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { api, Place, User, setLanguage } from '../api/client';
import { useAuth } from '../context/AuthContext';

export default function ProfilePage() {
  const { t, i18n } = useTranslation();
  const { user, signIn, signUp, signOut, refresh } = useAuth();
  const [authOpen, setAuthOpen] = useState(false);
  const [favorites, setFavorites] = useState<Place[]>([]);
  const [stats, setStats] = useState({ favorites: 0, routes: 0, posts: 0 });

  const loadUserData = async (u: User) => {
    try {
      const [favs, s] = await Promise.all([
        api.get<Place[]>('/api/favorites'),
        api.get<{ favorites: number; routes: number; posts: number }>('/api/me/stats'),
      ]);
      setFavorites(favs || []);
      setStats(s || { favorites: 0, routes: 0, posts: 0 });
      // Align the UI language with the profile saved server-side.
      if (i18n.language !== u.language) void i18n.changeLanguage(u.language);
    } catch {
      /* logged out mid-flight; ignore */
    }
  };

  useEffect(() => {
    if (user) void loadUserData(user);
  }, [user?.id]);

  const finishAuth = async (values: { username: string; password: string; nickname?: string }) => {
    try {
      await signIn(values.username, values.password);
      setAuthOpen(false);
      void refresh();
      message.success('OK');
    } catch {
      try {
        await signUp(values.username, values.password, values.nickname);
        setAuthOpen(false);
        message.success('OK');
      } catch (e2) {
        message.error((e2 as Error).message);
      }
    }
  };

  const switchLanguage = async () => {
    const next = i18n.language === 'zh' ? 'en' : 'zh';
    setLanguage(next);
    void i18n.changeLanguage(next);
    if (user) {
      await api.put('/api/me', { language: next }).catch(() => undefined);
    }
  };

  return (
    <div>
      <div className="section-title">{t('profile.title')}</div>

      <div className="profile-layout">
        <div className="stack">
          <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
              <Avatar size={56} src={user?.avatarUrl}>
                {user?.nickname?.[0] || '?'}
              </Avatar>
              <div style={{ flex: 1 }}>
                <b>{user?.nickname || t('profile.guest')}</b>
                <div style={{ fontSize: 12, color: '#999' }}>{user?.username}</div>
              </div>
              {user ? (
                <Button onClick={signOut}>{t('common.logout')}</Button>
              ) : (
                <Button type="primary" onClick={() => setAuthOpen(true)}>
                  {t('common.login')}
                </Button>
              )}
            </div>
          </Card>

          <Card>
            <div style={{ display: 'flex', justifyContent: 'space-around' }}>
              <Statistic title={t('profile.favorites')} value={user ? stats.favorites : 0} />
              <Statistic title={t('profile.history')} value={user ? stats.routes : 0} />
              <Statistic title={t('profile.posts')} value={user ? stats.posts : 0} />
            </div>
          </Card>

          <Button block icon={<GlobalOutlined />} onClick={switchLanguage}>
            {i18n.language === 'zh' ? t('profile.switchToEn') : t('profile.switchToZh')}
          </Button>
        </div>

        <Card title={t('profile.favorites')}>
          {user && favorites.length > 0 ? (
            favorites.map((p) => (
              <div key={p.id} style={{ padding: '6px 0', borderBottom: '1px solid #f0f0f0' }}>
                <b>{i18n.language === 'zh' ? p.nameCn : p.nameEn}</b>
                <div style={{ fontSize: 12, color: '#777' }}>
                  {i18n.language === 'zh' ? p.descriptionCn : p.descriptionEn}
                </div>
              </div>
            ))
          ) : (
            <Empty description={t('profile.loginToView')} />
          )}
        </Card>
      </div>

      <Modal
        title={t('common.login')}
        open={authOpen}
        footer={null}
        onCancel={() => setAuthOpen(false)}
      >
        <Tabs
          items={[
            {
              key: 'login',
              label: t('profile.signIn'),
              children: (
                <AuthForm
                  onOk={finishAuth}
                  submitLabel={t('profile.signIn')}
                  fields={['username', 'password']}
                />
              ),
            },
            {
              key: 'register',
              label: t('profile.signUp'),
              children: (
                <AuthForm
                  onOk={finishAuth}
                  submitLabel={t('profile.signUp')}
                  fields={['username', 'password', 'nickname']}
                />
              ),
            },
          ]}
        />
      </Modal>
    </div>
  );
}

import { Form, Input as AntInput } from 'antd';

function AuthForm({
  onOk,
  submitLabel,
  fields,
}: {
  onOk: (values: { username: string; password: string; nickname?: string }) => Promise<void>;
  submitLabel: string;
  fields: string[];
}) {
  const [form] = Form.useForm();
  return (
    <Form form={form} layout="vertical" onFinish={onOk}>
      <Form.Item name="username" label="Username" rules={[{ required: true, min: 3 }]}>
        <AntInput />
      </Form.Item>
      <Form.Item name="password" label="Password" rules={[{ required: true, min: 6 }]}>
        <AntInput.Password />
      </Form.Item>
      {fields.includes('nickname') && (
        <Form.Item name="nickname" label="Nickname">
          <AntInput />
        </Form.Item>
      )}
      <Button type="primary" block htmlType="submit">
        {submitLabel}
      </Button>
    </Form>
  );
}
