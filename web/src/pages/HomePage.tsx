import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Carousel, Spin } from 'antd';
import {
  NodeIndexOutlined,
  FileTextOutlined,
  CustomerServiceOutlined,
  ReadOutlined,
} from '@ant-design/icons';
import { api, ScenicImage } from '../api/client';

export default function HomePage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [scenics, setScenics] = useState<ScenicImage[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<ScenicImage[]>('/api/scenics')
      .then(setScenics)
      .catch(() => setScenics([]))
      .finally(() => setLoading(false));
  }, []);

  const lang = i18n.language === 'zh' ? 'Cn' : 'En';
  const titleOf = (s: ScenicImage) => (lang === 'Cn' ? s.titleCn : s.titleEn);
  const descOf = (s: ScenicImage) => (lang === 'Cn' ? s.descriptionCn : s.descriptionEn);

  const quickEntries = [
    { icon: <NodeIndexOutlined />, label: t('home.routes'), path: '/route' },
    { icon: <ReadOutlined />, label: t('home.travelBlog'), path: '/blog' },
    { icon: <CustomerServiceOutlined />, label: t('home.aiGuide'), path: '/guide' },
    { icon: <FileTextOutlined />, label: t('blog.title'), path: '/blog' },
  ];

  return (
    <div>
      <div className="home-hero">
        <h1>{t('home.title')}</h1>
        <p>{t('home.subtitle')}</p>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin />
        </div>
      ) : (
        <Carousel autoplay dots>
          {scenics.map((s) => (
            <div key={s.id} className="scenic-card">
              <img src={s.imageUrl} alt={titleOf(s)} />
              <div style={{ padding: '10px 4px' }}>
                <b>{titleOf(s)}</b>
                <div style={{ fontSize: 12, color: '#777' }}>{descOf(s)}</div>
              </div>
            </div>
          ))}
        </Carousel>
      )}

      <div className="quick-grid">
        {quickEntries.map((e, i) => (
          <button key={i} className="quick-item" onClick={() => navigate(e.path)}>
            {e.icon}
            <span>{e.label}</span>
          </button>
        ))}
      </div>

      <div className="section-title">{t('home.todaysPick')}</div>
      {scenics.length > 0 &&
        scenics
          .slice()
          .sort(() => Math.random() - 0.5)
          .slice(0, 2)
          .map((s) => (
            <div key={s.id} className="scenic-card" style={{ marginBottom: 14 }}>
              <img src={s.imageUrl} alt={titleOf(s)} />
              <div style={{ padding: '8px 4px' }}>
                <b>{titleOf(s)}</b>
                <div style={{ fontSize: 12, color: '#777' }}>{descOf(s)}</div>
              </div>
            </div>
          ))}
    </div>
  );
}
