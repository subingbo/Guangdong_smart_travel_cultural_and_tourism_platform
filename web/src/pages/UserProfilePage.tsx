import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Alert, Avatar, Button, Card, Empty, Modal, Pagination, Skeleton, message } from 'antd';
import { UserAddOutlined, UserDeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, Blog, Page, PublicUser, UserProfile } from '../api/client';
import { useAuth } from '../context/AuthContext';
import BlogCard from '../components/BlogCard';

const PAGE_SIZE = 9;

export default function UserProfilePage() {
  const { id } = useParams<{ id: string }>();
  const targetId = Number(id);
  const { t, i18n } = useTranslation();
  const { user, openAuth } = useAuth();

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [blogs, setBlogs] = useState<Page<Blog> | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [listKind, setListKind] = useState<'followers' | 'following' | null>(null);
  const [list, setList] = useState<Page<PublicUser> | null>(null);
  const [listLoading, setListLoading] = useState(false);

  const loadProfile = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setProfile(await api.get<UserProfile>(`/api/users/${targetId}`));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [targetId]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  useEffect(() => {
    api
      .get<Page<Blog>>(`/api/users/${targetId}/blogs?page=${page}&pageSize=${PAGE_SIZE}`)
      .then(setBlogs)
      .catch(() => setBlogs(null));
  }, [targetId, page]);

  const toggleFollow = async () => {
    if (!profile) return;
    if (!user) {
      openAuth();
      return;
    }
    try {
      if (profile.isFollowing) {
        await api.del(`/api/users/${profile.user.id}/follow`);
      } else {
        await api.post(`/api/users/${profile.user.id}/follow`);
      }
      setProfile({
        ...profile,
        isFollowing: !profile.isFollowing,
        stats: {
          ...profile.stats,
          followers: Math.max(0, profile.stats.followers + (profile.isFollowing ? -1 : 1)),
        },
      });
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const openList = async (kind: 'followers' | 'following') => {
    setListKind(kind);
    setListLoading(true);
    try {
      setList(await api.get<Page<PublicUser>>(`/api/users/${targetId}/${kind}?page=1&pageSize=50`));
    } catch {
      setList(null);
    } finally {
      setListLoading(false);
    }
  };

  const toggleLike = async (blog: Blog) => {
    if (!user) {
      openAuth();
      return;
    }
    try {
      const res = blog.likedByMe
        ? await api.del<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`)
        : await api.post<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`);
      setBlogs((prev) =>
        prev
          ? {
              ...prev,
              items: prev.items.map((b) =>
                b.id === blog.id ? { ...b, likes: res.likes, likedByMe: res.likedByMe } : b,
              ),
            }
          : prev,
      );
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  if (loading) return <Skeleton active />;
  if (error || !profile) {
    return <Alert type="error" showIcon message={error || t('user.loadFailed')} />;
  }

  const joined = new Date(profile.user.createdAt).toLocaleDateString(
    i18n.language === 'zh' ? 'zh-CN' : 'en-US',
  );

  return (
    <div>
      <Card className="profile-hero">
        <div className="profile-hero-main">
          <Avatar size={80} src={assetUrl(profile.user.avatarUrl)}>
            {profile.user.nickname?.[0]}
          </Avatar>
          <div className="profile-hero-text">
            <h2>{profile.user.nickname}</h2>
            <div className="author-sub">@{profile.user.username}</div>
            <div className="author-sub">
              {t('user.joined')}: {joined}
            </div>
          </div>
          {user?.id !== profile.user.id && (
            <Button
              type={profile.isFollowing ? 'default' : 'primary'}
              icon={profile.isFollowing ? <UserDeleteOutlined /> : <UserAddOutlined />}
              onClick={toggleFollow}
            >
              {profile.isFollowing ? t('user.unfollow') : t('user.follow')}
            </Button>
          )}
        </div>

        <div className="profile-stats">
          <div>
            <b>{profile.stats.posts}</b>
            <span>{t('user.posts')}</span>
          </div>
          <button type="button" onClick={() => void openList('followers')}>
            <b>{profile.stats.followers}</b>
            <span>{t('user.followers')}</span>
          </button>
          <button type="button" onClick={() => void openList('following')}>
            <b>{profile.stats.following}</b>
            <span>{t('user.following')}</span>
          </button>
          <div>
            <b>{profile.stats.likesReceived}</b>
            <span>{t('user.likesReceived')}</span>
          </div>
        </div>
      </Card>

      <div className="section-title">{t('profile.myPosts')}</div>
      {!blogs || blogs.items.length === 0 ? (
        <Empty description={t('user.noPosts')} />
      ) : (
        <>
          <div className="blog-grid">
            {blogs.items.map((b) => (
              <BlogCard key={b.id} blog={b} currentUserId={user?.id} onLike={toggleLike} />
            ))}
          </div>
          <div className="pagination-row">
            <Pagination
              current={blogs.page}
              pageSize={blogs.pageSize}
              total={blogs.total}
              showSizeChanger={false}
              onChange={(p) => setPage(p)}
            />
          </div>
        </>
      )}

      <Modal
        open={listKind !== null}
        title={listKind === 'followers' ? t('user.followersTitle') : t('user.followingTitle')}
        footer={null}
        onCancel={() => setListKind(null)}
      >
        {listLoading ? (
          <Skeleton active />
        ) : !list || list.items.length === 0 ? (
          <Empty />
        ) : (
          <div className="user-list">
            {list.items.map((u) => (
              <Link key={u.id} to={`/user/${u.id}`} className="user-list-item" onClick={() => setListKind(null)}>
                <Avatar src={assetUrl(u.avatarUrl)}>{u.nickname?.[0]}</Avatar>
                <span>{u.nickname}</span>
                <small>@{u.username}</small>
              </Link>
            ))}
          </div>
        )}
      </Modal>
    </div>
  );
}
