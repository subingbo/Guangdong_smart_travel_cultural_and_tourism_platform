import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  NodeIndexOutlined,
  ReadOutlined,
  HomeOutlined,
  CustomerServiceOutlined,
  UserOutlined,
} from '@ant-design/icons';

const TABS = [
  { path: '/route', key: 'route', icon: <NodeIndexOutlined /> },
  { path: '/blog', key: 'blog', icon: <ReadOutlined /> },
  { path: '/', key: 'home', icon: <HomeOutlined /> },
  { path: '/guide', key: 'guide', icon: <CustomerServiceOutlined /> },
  { path: '/profile', key: 'profile', icon: <UserOutlined /> },
];

export default function BottomNav() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t } = useTranslation();

  return (
    <nav className="bottom-nav">
      {TABS.map((tab) => {
        const active = location.pathname === tab.path;
        return (
          <button
            key={tab.key}
            className={`bottom-nav-item${active ? ' active' : ''}`}
            onClick={() => navigate(tab.path)}
          >
            {tab.icon}
            <span>{t(`tab.${tab.key}`)}</span>
          </button>
        );
      })}
    </nav>
  );
}
