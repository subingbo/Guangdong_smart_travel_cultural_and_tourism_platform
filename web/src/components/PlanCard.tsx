import { Alert, Button, Card, Space, Tag, Typography } from 'antd';
import { CheckCircleOutlined, EnvironmentOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { AgentPlan, AgentPlanStop } from '../api/agent';
import type { Coordinate, Waypoint } from '../api/client';
import AmapRouteMap from './AmapRouteMap';

interface Props {
  plan: AgentPlan;
  language: 'en' | 'zh';
  confirming?: boolean;
  confirmedRouteId?: number;
  onConfirm: () => void;
}

export default function PlanCard({ plan, language, confirming, confirmedRouteId, onConfirm }: Props) {
  const { t } = useTranslation();
  const locale = language === 'zh' ? 'zh-CN' : 'en-US';
  const placeStops = plan.stops.filter((s) => s.placeId > 0);
  const waypoints: Waypoint[] = placeStops.map((s) => ({
    name: s.name,
    nameEn: s.nameEn || s.name,
    lat: s.lat,
    lng: s.lng,
    stayTime: `${s.stayMinutes} min`,
    description: '',
  }));
  const polyline: Coordinate[] =
    plan.segments.flatMap((seg) => seg.polyline || []).length > 1
      ? plan.segments.flatMap((seg) => seg.polyline || [])
      : placeStops.map((s) => ({ lat: s.lat, lng: s.lng }));
  const estimated = plan.routeConfidence === 'ESTIMATED';

  return (
    <Card
      size="small"
      className="agent-plan-card"
      title={
        <Space size={6} wrap>
          <EnvironmentOutlined />
          <span>{plan.title}</span>
          <Tag color="blue">{Math.round(plan.score * 100)}%</Tag>
          <Tag color={estimated ? 'orange' : 'green'}>
            {estimated ? t('guide.estimated') : t('guide.liveTraffic')}
          </Tag>
        </Space>
      }
    >
      {plan.summary && <Typography.Paragraph className="agent-plan-summary">{plan.summary}</Typography.Paragraph>}
      <div className="agent-plan-meta">
        <span>
          <ClockCircleOutlined /> {formatTime(plan.windowStart, locale)} – {formatTime(plan.windowEnd, locale)}
        </span>
        <span>
          {t('guide.totalTravel')}: {plan.totalTravelMinutes} min · {formatDistance(plan.totalDistanceMeters)}
        </span>
      </div>

      <div className="agent-timeline">
        {plan.stops.map((stop, index) => {
          const leg = plan.segments.find((seg) => seg.toIndex === index);
          return (
            <div key={`${stop.placeId}-${index}`} className="agent-timeline-row">
              {leg && (
                <div className="agent-leg">
                  <span className="agent-leg-line" />
                  <span>
                    {leg.durationMinutes} min · {formatDistance(leg.distanceMeters)}
                    {leg.estimated && <Tag color="orange">{t('guide.estimated')}</Tag>}
                  </span>
                </div>
              )}
              <StopRow stop={stop} locale={locale} />
            </div>
          );
        })}
      </div>

      {polyline.length > 1 && (
        <div className="agent-plan-map">
          <AmapRouteMap polyline={polyline} waypoints={waypoints} language={language} />
        </div>
      )}

      {confirmedRouteId ? (
        <Alert
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
          message={t('guide.planConfirmed')}
          description={
            <a href={`#/routes/${confirmedRouteId}`} className="agent-plan-link">
              {t('guide.viewRoute')}
            </a>
          }
        />
      ) : (
        <Button type="primary" block loading={confirming} onClick={onConfirm}>
          {t('guide.confirmPlan')}
        </Button>
      )}
    </Card>
  );
}

function StopRow({ stop, locale }: { stop: AgentPlanStop; locale: string }) {
  const { t } = useTranslation();
  if (stop.placeId === 0) {
    return (
      <div className="agent-stop agent-origin">
        <span className="agent-stop-time">
          {formatTime(stop.arrivalTime, locale)} – {formatTime(stop.leaveTime, locale)}
        </span>
        <span className="agent-stop-name">
          {stop.name || t('guide.origin')} · {t('guide.originLabel')}
        </span>
      </div>
    );
  }
  return (
    <div className="agent-stop">
      <span className="agent-stop-time">
        {formatTime(stop.arrivalTime, locale)} – {formatTime(stop.leaveTime, locale)}
      </span>
      <span className="agent-stop-name">{stop.name}</span>
      <Tag>{stop.stayMinutes} min</Tag>
    </div>
  );
}

function formatTime(value: string, locale: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit', hour12: false });
}

function formatDistance(meters: number): string {
  if (meters >= 1000) return `${(meters / 1000).toFixed(1)} km`;
  return `${meters} m`;
}
