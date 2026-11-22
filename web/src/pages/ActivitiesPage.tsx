import { useCallback, useEffect, useState } from 'react';
import { Button, Empty, Modal, Spin, Tag, message } from 'antd';
import { CalendarOutlined, EnvironmentOutlined, TeamOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { Activity, api, Page } from '../api/client';
import SafeImage from '../components/SafeImage';

const COPY = {
  zh: {
    eyebrow: '岭南日历',
    title: '近期文旅活动',
    subtitle: '美食节、非遗手作、珠江夜游、摄影漫步与粤剧体验，看看这周广州有什么值得去。',
    upcoming: '即将开始',
    all: '全部活动',
    empty: '暂时没有活动，去看看路线和游记吧。',
    detail: '活动详情',
    organizer: '主办方',
    location: '地点',
    time: '时间',
    close: '关闭',
    loadFailed: '活动加载失败',
    ongoing: '进行中',
    upcomingStatus: '即将开始',
    ended: '已结束',
  },
  en: {
    eyebrow: 'Lingnan Calendar',
    title: 'Upcoming Cultural Events',
    subtitle: 'Food festivals, heritage crafts, river cruises, photo walks and Cantonese opera — what is worth doing in Guangzhou this week.',
    upcoming: 'Upcoming',
    all: 'All events',
    empty: 'No events right now. Explore routes and stories instead.',
    detail: 'Event details',
    organizer: 'Organizer',
    location: 'Location',
    time: 'Time',
    close: 'Close',
    loadFailed: 'Could not load events',
    ongoing: 'Ongoing',
    upcomingStatus: 'Upcoming',
    ended: 'Ended',
  },
} as const;

function formatRange(start: string, end: string, zh: boolean) {
  const s = new Date(start);
  const e = new Date(end);
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
  if (zh) {
    return `${s.getMonth() + 1}月${s.getDate()}日 - ${e.getMonth() + 1}月${e.getDate()}日`;
  }
  return `${s.toLocaleDateString('en-US', opts)} – ${e.toLocaleDateString('en-US', opts)}`;
}

export default function ActivitiesPage() {
  const { i18n } = useTranslation();
  const zh = i18n.language === 'zh';
  const L = zh ? COPY.zh : COPY.en;
  const [status, setStatus] = useState<'upcoming' | 'all'>('upcoming');
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<Activity | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.get<Page<Activity>>(
        `/api/activities?status=${status}&page=1&pageSize=30`,
      );
      setActivities(res.items || []);
    } catch (e) {
      message.error((e as Error).message || L.loadFailed);
      setActivities([]);
    } finally {
      setLoading(false);
    }
  }, [status, L.loadFailed]);

  useEffect(() => {
    void load();
  }, [load]);

  const title = (a: Activity) => (zh ? a.titleCn : a.titleEn);
  const description = (a: Activity) => (zh ? a.descriptionCn : a.descriptionEn);
  const statusLabel = (value: string) =>
    value === 'ongoing' ? L.ongoing : value === 'ended' ? L.ended : L.upcomingStatus;

  return (
    <div className="activities-page">
      <div className="activities-hero">
        <div>
          <div className="activities-eyebrow">{L.eyebrow}</div>
          <h1>{L.title}</h1>
          <p>{L.subtitle}</p>
        </div>
        <div className="activities-filters">
          <Button
            type={status === 'upcoming' ? 'primary' : 'default'}
            onClick={() => setStatus('upcoming')}
          >
            {L.upcoming}
          </Button>
          <Button
            type={status === 'all' ? 'primary' : 'default'}
            onClick={() => setStatus('all')}
          >
            {L.all}
          </Button>
        </div>
      </div>

      {loading ? (
        <div className="activities-loading">
          <Spin />
        </div>
      ) : activities.length === 0 ? (
        <Empty description={L.empty} />
      ) : (
        <div className="activities-grid">
          {activities.map((activity) => (
            <article
              key={activity.id}
              className="activity-card"
              onClick={() => setDetail(activity)}
            >
              <div className="activity-image">
                <SafeImage src={activity.imageUrl} alt={title(activity)} />
                <Tag className="activity-tag">{activity.tag}</Tag>
              </div>
              <div className="activity-body">
                <div className="activity-date">
                  <CalendarOutlined /> {formatRange(activity.startAt, activity.endAt, zh)}
                </div>
                <h3>{title(activity)}</h3>
                <p>{description(activity)}</p>
                <div className="activity-meta">
                  <span>
                    <EnvironmentOutlined /> {activity.location}
                  </span>
                  <span>
                    <TeamOutlined /> {activity.organizer}
                  </span>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}

      <Modal
        open={!!detail}
        title={L.detail}
        onCancel={() => setDetail(null)}
        footer={
          <Button type="primary" onClick={() => setDetail(null)}>
            {L.close}
          </Button>
        }
        width={760}
      >
        {detail && (
          <div className="activity-detail">
            <SafeImage className="activity-detail-image" src={detail.imageUrl} alt={title(detail)} />
            <div className="activity-detail-body">
              <Tag color="blue">{detail.tag}</Tag>
              <Tag color="gold">{statusLabel(detail.status)}</Tag>
              <h2>{title(detail)}</h2>
              <p>{description(detail)}</p>
              <div className="activity-detail-row">
                <CalendarOutlined />
                <span>{formatRange(detail.startAt, detail.endAt, zh)}</span>
              </div>
              <div className="activity-detail-row">
                <EnvironmentOutlined />
                <span>{detail.location}</span>
              </div>
              <div className="activity-detail-row">
                <TeamOutlined />
                <span>
                  {L.organizer}: {detail.organizer}
                </span>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
