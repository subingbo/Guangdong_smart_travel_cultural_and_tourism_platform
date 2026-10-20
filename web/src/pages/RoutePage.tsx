import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Card, Empty, InputNumber, Select, Slider, Spin, Tag, message } from 'antd';
import L from 'leaflet';
import { api, Route as RouteType } from '../api/client';

const INTERESTS = ['foodie', 'history', 'nature', 'art', 'shopping'];
const PACES = ['relaxed', 'balanced', 'packed'];
const MODES = ['driving', 'walking', 'bicycling', 'transit'];

export default function RoutePage() {
  const { t, i18n } = useTranslation();
  const [days, setDays] = useState(1);
  const [interests, setInterests] = useState<string[]>(['foodie']);
  const [pace, setPace] = useState('balanced');
  const [mode, setMode] = useState('driving');
  const [routes, setRoutes] = useState<RouteType[]>([]);
  const [current, setCurrent] = useState<RouteType | null>(null);
  const [loading, setLoading] = useState(false);
  const mapRef = useRef<HTMLDivElement>(null);
  const leafletRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  const drawRoute = (route: RouteType) => {
    if (!mapRef.current) return;
    if (!leafletRef.current) {
      leafletRef.current = L.map(mapRef.current).setView([23.1291, 113.2644], 12);
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors',
      }).addTo(leafletRef.current);
      layerRef.current = L.layerGroup().addTo(leafletRef.current);
    }
    const layer = layerRef.current!;
    layer.clearLayers();
    const pts: L.LatLngExpression[] = route.polyline.map((p) => [p.lat, p.lng]);
    if (pts.length >= 2) {
      layer.addLayer(L.polyline(pts, { color: '#0052d9', weight: 4 }));
    }
    route.waypoints.forEach((wp, i) => {
      layer.addLayer(
        L.marker([wp.lat, wp.lng]).bindTooltip(
          `${i + 1}. ${i18n.language === 'zh' ? wp.name : wp.nameEn || wp.name}`,
        ),
      );
    });
    if (pts.length > 0) {
      leafletRef.current.fitBounds(L.latLngBounds(pts), { padding: [30, 30] });
    }
  };

  useEffect(() => {
    if (current) drawRoute(current);
  }, [current, i18n.language]);

  useEffect(() => {
    api
      .get<RouteType[]>('/api/routes')
      .then(setRoutes)
      .catch(() => setRoutes([]));
  }, []);

  const generate = async () => {
    setLoading(true);
    try {
      const route = await api.post<RouteType>('/api/routes/generate', {
        days,
        interests,
        pace,
        mode,
      });
      setCurrent(route);
      setRoutes((prev) => [route, ...prev]);
      message.success(route.routeName);
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="section-title">{t('route.title')}</div>
      <Card>
        <div style={{ marginBottom: 12 }}>
          <b>{t('route.days')}</b>
          <InputNumber min={1} max={7} value={days} onChange={(v) => setDays(v || 1)} style={{ width: '100%' }} />
        </div>
        <div style={{ marginBottom: 12 }}>
          <b>{t('route.interests')}</b>
          <Select
            mode="multiple"
            style={{ width: '100%' }}
            value={interests}
            onChange={setInterests}
            options={INTERESTS.map((i) => ({ value: i, label: t(`route.interest${i[0].toUpperCase()}${i.slice(1)}`) }))}
          />
        </div>
        <div style={{ marginBottom: 12 }}>
          <b>{t('route.pace')}</b>
          <Slider
            min={0}
            max={2}
            step={1}
            marks={{ 0: t('route.paceRelaxed'), 1: t('route.paceBalanced'), 2: t('route.pacePacked') }}
            value={PACES.indexOf(pace)}
            onChange={(v) => setPace(PACES[v as number])}
          />
        </div>
        <div style={{ marginBottom: 16 }}>
          <b>{t('route.mode')}</b>
          <Select
            style={{ width: '100%' }}
            value={mode}
            onChange={setMode}
            options={MODES.map((m) => ({ value: m, label: t(`route.mode${m[0].toUpperCase()}${m.slice(1)}`) }))}
          />
        </div>
        <Button type="primary" block loading={loading} onClick={generate}>
          {loading ? t('route.generating') : t('route.generate')}
        </Button>
      </Card>

      {loading && (
        <div style={{ textAlign: 'center', padding: 20 }}>
          <Spin />
        </div>
      )}

      {current && (
        <Card style={{ marginTop: 16 }}>
          <b style={{ fontSize: 16 }}>{current.routeName}</b>
          <div style={{ margin: '6px 0' }}>
            <Tag>{t('route.distance')}: {current.totalDistance}</Tag>
            <Tag>{current.durationDays}d</Tag>
            <Tag>{current.mode}</Tag>
          </div>
          <div ref={mapRef} className="route-map" />
          <ol style={{ paddingLeft: 20, marginTop: 12 }}>
            {current.waypoints.map((wp, i) => (
              <li key={i} style={{ marginBottom: 6 }}>
                <b>{i18n.language === 'zh' ? wp.name : wp.nameEn || wp.name}</b>
                {wp.stayTime && <Tag style={{ marginLeft: 6 }}>{wp.stayTime}</Tag>}
                <div style={{ fontSize: 12, color: '#777' }}>{wp.description}</div>
              </li>
            ))}
          </ol>
        </Card>
      )}

      <div className="section-title">{t('route.myRoutes')}</div>
      {routes.length === 0 ? (
        <Empty description={t('profile.loginToView')} />
      ) : (
        routes.map((r) => (
          <Card
            key={r.id}
            size="small"
            style={{ marginBottom: 10, cursor: 'pointer' }}
            onClick={() => setCurrent(r)}
          >
            <b>{r.routeName}</b>
            <div style={{ fontSize: 12, color: '#777' }}>
              {r.durationDays}d · {r.totalDistance} · {r.mode}
            </div>
          </Card>
        ))
      )}
    </div>
  );
}
