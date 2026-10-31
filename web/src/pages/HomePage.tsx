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
      .then((data) => setScenics(data || []))
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

  // Rotate the picks by day instead of shuffling randomly (stable, reproducible UI).
  const offset = scenics.length > 0 ? new Date().getDate() % scenics.length : 0;
  const picks = [...scenics.slice(offset), ...scenics.slice(0, offset)].slice(0, 6);

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
        <Carousel
          autoplay
          dots
          slidesToShow={3}
          responsive={[{ breakpoint: 900, settings: { slidesToShow: 1 } }]}
        >
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
          <button key={i} type="button" className="quick-item" onClick={() => navigate(e.path)}>
            {e.icon}
            <span>{e.label}</span>
          </button>
        ))}
      </div>

      <div className="section-title">{t('home.todaysPick')}</div>
      <div className="card-grid">
        {picks.map((s) => (
          <article key={s.id} className="pick-card">
            <img src={s.imageUrl} alt={titleOf(s)} />
            <div className="pick-body">
              <b>{titleOf(s)}</b>
              <div style={{ fontSize: 12, color: '#777', marginTop: 4 }}>{descOf(s)}</div>
            </div>
          </article>
        ))}
      </div>
    </div>
  );
}
