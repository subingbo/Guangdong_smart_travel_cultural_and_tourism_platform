import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Carousel, Spin, Tag } from 'antd';
import {
  NodeIndexOutlined,
  CustomerServiceOutlined,
  ReadOutlined,
  ShoppingOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import { Activity, api, Page, ScenicImage } from '../api/client';
import SafeImage from '../components/SafeImage';

export default function HomePage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [scenics, setScenics] = useState<ScenicImage[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<ScenicImage[]>('/api/scenics')
      .then((data) => setScenics(data || []))
      .catch(() => setScenics([]))
      .finally(() => setLoading(false));
    api
      .get<Page<Activity>>('/api/activities?status=upcoming&page=1&pageSize=3')
      .then((data) => setActivities(data.items || []))
      .catch(() => setActivities([]));
  }, []);

  const lang = i18n.language === 'zh' ? 'Cn' : 'En';
  const titleOf = (s: ScenicImage) => (lang === 'Cn' ? s.titleCn : s.titleEn);
  const descOf = (s: ScenicImage) => (lang === 'Cn' ? s.descriptionCn : s.descriptionEn);

  const quickEntries = [
    { icon: <NodeIndexOutlined />, label: t('home.routes'), path: '/route' },
    { icon: <ReadOutlined />, label: t('home.travelBlog'), path: '/blog' },
    { icon: <CustomerServiceOutlined />, label: t('home.aiGuide'), path: '/guide' },
    { icon: <ShoppingOutlined />, label: t('tab.shop'), path: '/shop' },
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
              <SafeImage src={s.imageUrl} alt={titleOf(s)} />
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
            <SafeImage src={s.imageUrl} alt={titleOf(s)} />
            <div className="pick-body">
              <b>{titleOf(s)}</b>
              <div style={{ fontSize: 12, color: '#777', marginTop: 4 }}>{descOf(s)}</div>
            </div>
          </article>
        ))}
      </div>

      {activities.length > 0 && (
        <>
          <div className="section-title home-section-title">
            <span>{t('home.activities')}</span>
            <button type="button" className="home-view-all" onClick={() => navigate('/activities')}>
              {t('home.viewAll')} →
            </button>
          </div>
          <div className="home-activity-grid">
            {activities.map((a) => (
              <article key={a.id} className="home-activity-card" onClick={() => navigate('/activities')}>
                <SafeImage
                  src={a.imageUrl}
                  alt={lang === 'Cn' ? a.titleCn : a.titleEn}
                />
                <div className="home-activity-body">
                  <Tag color="blue">{a.tag}</Tag>
                  <b>{lang === 'Cn' ? a.titleCn : a.titleEn}</b>
                  <span>
                    <CalendarOutlined /> {new Date(a.startAt).toLocaleDateString(lang === 'Cn' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric' })}
                    {' · '}
                    {a.location}
                  </span>
                </div>
              </article>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
