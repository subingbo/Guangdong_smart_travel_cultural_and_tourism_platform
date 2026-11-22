import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Avatar, Badge, Button } from 'antd';
import { BellOutlined, GlobalOutlined } from '@ant-design/icons';
import { api, assetUrl, setLanguage } from '../api/client';
import { useAuth } from '../context/AuthContext';

const NAV = [
  { path: '/', key: 'home' },
  { path: '/route', key: 'route' },
  { path: '/shop', key: 'shop' },
  { path: '/blog', key: 'blog' },
  { path: '/activities', key: 'activities' },
  { path: '/guide', key: 'guide' },
  { path: '/merchant', key: 'merchant' },
  { path: '/profile', key: 'profile' },
];

export default function AppHeader() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { user, signOut, openAuth } = useAuth();
  const [unread, setUnread] = useState(0);

  useEffect(() => {
    if (!user) {
      setUnread(0);
      return;
    }
    api
      .get<{ unread: number }>('/api/notifications/unread-count')
      .then((r) => setUnread(r.unread || 0))
      .catch(() => setUnread(0));
  }, [user?.id, location.pathname]);

  const switchLanguage = async () => {
    const next = i18n.language === 'zh' ? 'en' : 'zh';
    setLanguage(next);
    void i18n.changeLanguage(next);
    if (user) {
      await api.put('/api/me', { language: next }).catch(() => undefined);
    }
  };

  const navItems = user?.role === 'admin' ? [...NAV, { path: '/admin', key: 'admin' }] : NAV;

  return (
    <header className="app-header">
      <div className="app-header-inner">
        <button type="button" className="app-brand" onClick={() => navigate('/')}>
          <span className="app-brand-mark">粤</span>
          <span className="app-brand-text">
            <b>{t('common.brand')}</b>
            <small>South Guangdong Tour</small>
          </span>
        </button>

        <nav className="app-header-nav">
          {navItems.map((item) => (
            <button
              key={item.key}
              type="button"
              className={`app-header-link${location.pathname === item.path ? ' active' : ''}`}
              onClick={() => navigate(item.path)}
            >
              {t(`tab.${item.key}`)}
            </button>
          ))}
        </nav>

        <div className="app-header-actions">
          <Button type="text" icon={<GlobalOutlined />} onClick={switchLanguage}>
            {i18n.language === 'zh' ? 'EN' : '中文'}
          </Button>
          {user && (
            <Badge count={unread} size="small" offset={[-2, 4]}>
              <Button
                type="text"
                icon={<BellOutlined />}
                aria-label={t('notifications.title')}
                onClick={() => navigate('/notifications')}
              />
            </Badge>
          )}
          {user ? (
            <>
              <Button type="text" onClick={() => navigate('/profile')}>
                <Avatar size={26} src={assetUrl(user.avatarUrl)} style={{ marginRight: 6 }}>
                  {user.nickname?.[0]}
                </Avatar>
                {user.nickname}
              </Button>
              <Button type="text" onClick={signOut}>
                {t('common.logout')}
              </Button>
            </>
          ) : (
            <Button type="primary" onClick={openAuth}>
              {t('common.login')}
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
