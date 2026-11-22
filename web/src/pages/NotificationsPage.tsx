import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Alert, Avatar, Button, Empty, Pagination, Segmented, Skeleton, message } from 'antd';
import { CheckOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, NotificationItem, NotificationPage } from '../api/client';

const PAGE_SIZE = 20;

export default function NotificationsPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [filter, setFilter] = useState<'all' | 'unread'>('all');
  const [page, setPage] = useState(1);
  const [data, setData] = useState<NotificationPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(
        await api.get<NotificationPage>(
          `/api/notifications?page=${page}&pageSize=${PAGE_SIZE}&filter=${filter}`,
        ),
      );
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [page, filter]);

  useEffect(() => {
    void load();
  }, [load]);

  const openItem = async (n: NotificationItem) => {
    if (!n.isRead) {
      await api.post('/api/notifications/read', { ids: [n.id] }).catch(() => undefined);
    }
    if (n.type === 'follow' && n.actor) {
      navigate(`/user/${n.actor.id}`);
    } else if (n.blogId) {
      navigate(`/blog/${n.blogId}`);
    }
  };

  const markAll = async () => {
    try {
      await api.post('/api/notifications/read-all');
      message.success(t('notifications.markReadSuccess'));
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const describe = (n: NotificationItem) => {
    const name = n.actor?.nickname || '—';
    if (n.type === 'like') return `${name} ${t('notifications.likedYourPost')}`;
    if (n.type === 'comment') return `${name} ${t('notifications.commented')}`;
    return `${name} ${t('notifications.followedYou')}`;
  };

  const detail = (n: NotificationItem) => {
    if (n.type === 'comment') return n.commentContent || n.blogTitle;
    return n.blogTitle;
  };

  const time = (value: string) =>
    new Date(value).toLocaleString(i18n.language === 'zh' ? 'zh-CN' : 'en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });

  return (
    <div className="notifications-page">
      <div className="notifications-toolbar">
        <div className="section-title" style={{ margin: 0 }}>
          {t('notifications.title')}
        </div>
        <div className="notifications-actions">
          <Segmented
            value={filter}
            onChange={(v) => {
              setFilter(v as 'all' | 'unread');
              setPage(1);
            }}
            options={[
              { label: t('notifications.all'), value: 'all' },
              { label: t('notifications.unread'), value: 'unread' },
            ]}
          />
          <Button icon={<CheckOutlined />} onClick={markAll}>
            {t('notifications.markAllRead')}
          </Button>
        </div>
      </div>

      {loading ? (
        <Skeleton active />
      ) : error ? (
        <Alert
          type="error"
          showIcon
          message={error}
          action={
            <Button size="small" onClick={() => void load()}>
              {t('common.retry')}
            </Button>
          }
        />
      ) : !data || data.items.length === 0 ? (
        <Empty description={t('notifications.empty')} />
      ) : (
        <>
          <div className="notification-list">
            {data.items.map((n) => (
              <button
                key={n.id}
                type="button"
                className={`notification-item${n.isRead ? '' : ' unread'}`}
                onClick={() => void openItem(n)}
              >
                <Avatar size={40} src={assetUrl(n.actor?.avatarUrl)}>
                  {n.actor?.nickname?.[0]}
                </Avatar>
                <div className="notification-body">
                  <div className="notification-text">{describe(n)}</div>
                  {detail(n) && <div className="notification-detail">{detail(n)}</div>}
                  <div className="notification-time">{time(n.createdAt)}</div>
                </div>
                {!n.isRead && <span className="notification-dot" />}
              </button>
            ))}
          </div>
          <div className="pagination-row">
            <Pagination
              current={data.page}
              pageSize={data.pageSize}
              total={data.total}
              showSizeChanger={false}
              onChange={(p) => setPage(p)}
            />
          </div>
        </>
      )}
    </div>
  );
}
