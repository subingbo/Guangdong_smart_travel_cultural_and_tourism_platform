import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { api, setLanguage } from '../api/client';
import { useAuth } from '../context/AuthContext';

const NAV = [
  { path: '/', key: 'home' },
  { path: '/route', key: 'route' },
  { path: '/blog', key: 'blog' },
  { path: '/guide', key: 'guide' },
  { path: '/profile', key: 'profile' },
];

export default function AppHeader() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { user, signOut } = useAuth();

  const switchLanguage = async () => {
    const next = i18n.language === 'zh' ? 'en' : 'zh';
    setLanguage(next);
    void i18n.changeLanguage(next);
    if (user) {
      await api.put('/api/me', { language: next }).catch(() => undefined);
    }
  };

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
          {NAV.map((item) => (
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
          {user ? (
            <>
              <Button type="text" onClick={() => navigate('/profile')}>
                {user.nickname}
              </Button>
              <Button type="text" onClick={signOut}>
                {t('common.logout')}
              </Button>
            </>
          ) : (
            <Button type="primary" onClick={() => navigate('/profile')}>
              {t('common.login')}
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
