import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Alert, Button, Card, Spin, Tag } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, type Route } from '../api/client';
import AmapRouteMap from '../components/AmapRouteMap';

export default function RouteDetailPage() {
  const { id } = useParams();
  const { t, i18n } = useTranslation();
  const [route, setRoute] = useState<Route | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const language = i18n.language === 'zh' ? 'zh' : 'en';

  useEffect(() => {
    let disposed = false;
    setLoading(true);
    api
      .get<Route>(`/api/routes/${id}`)
      .then((value) => {
        if (!disposed) setRoute(value);
      })
      .catch((e: Error) => {
        if (!disposed) setError(e.message);
      })
      .finally(() => {
        if (!disposed) setLoading(false);
      });
    return () => {
      disposed = true;
    };
  }, [id]);

  if (loading) {
    return (
      <div className="stack">
        <Spin />
      </div>
    );
  }
  if (error || !route) {
    return (
      <div className="stack">
        <Alert type="error" message={error || t('route.loadFailed')} />
        <Link to="/guide">
          <Button icon={<ArrowLeftOutlined />}>{t('route.backToGuide')}</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="stack route-detail">
      <div className="route-detail-head">
        <Link to="/guide">
          <Button size="small" icon={<ArrowLeftOutlined />}>
            {t('route.backToGuide')}
          </Button>
        </Link>
        <div className="section-title" style={{ margin: 0 }}>
          {route.routeName}
        </div>
      </div>
      <Card size="small">
        <div className="route-detail-meta">
          <Tag color="blue">
            {t('route.distance')}: {route.totalDistance || '—'}
          </Tag>
          <Tag>{t('route.mode')}: {route.mode}</Tag>
          <Tag>
            {t('route.days')}: {route.durationDays}
          </Tag>
        </div>
        {route.polyline?.length > 1 && (
          <div className="route-detail-map">
            <AmapRouteMap
              polyline={route.polyline}
              waypoints={route.waypoints || []}
              language={language}
            />
          </div>
        )}
        <div className="agent-timeline">
          {(route.waypoints || []).map((wp, index) => (
            <div key={`${wp.name}-${index}`} className="agent-stop">
              <span className="agent-stop-time">{index + 1}</span>
              <span className="agent-stop-name">{language === 'zh' ? wp.name : wp.nameEn || wp.name}</span>
              {wp.stayTime && <Tag>{wp.stayTime}</Tag>}
              {wp.description && <span className="route-detail-desc">{wp.description}</span>}
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
