import { useCallback, useEffect, useState } from 'react';
import {
  Avatar,
  Button,
  Card,
  Empty,
  Form,
  Input,
  Modal,
  Skeleton,
  Tabs,
  Upload,
  message,
} from 'antd';
import { EditOutlined, GlobalOutlined, UploadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, Blog, MyStats, Page, Place, uploadImage } from '../api/client';
import { useAuth } from '../context/AuthContext';
import BlogCard from '../components/BlogCard';
import BlogEditor from '../components/BlogEditor';

const PAGE_SIZE = 6;

export default function ProfilePage() {
  const { t, i18n } = useTranslation();
  const { user, refresh, signOut, openAuth } = useAuth();

  const [stats, setStats] = useState<MyStats | null>(null);
  const [posts, setPosts] = useState<Page<Blog> | null>(null);
  const [favorites, setFavorites] = useState<Place[]>([]);
  const [postPage, setPostPage] = useState(1);
  const [tab, setTab] = useState('posts');
  const [loading, setLoading] = useState(false);

  const [editOpen, setEditOpen] = useState(false);
  const [nickname, setNickname] = useState('');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [saving, setSaving] = useState(false);

  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<Blog | null>(null);

  const loadStats = useCallback(async () => {
    if (!user) {
      setStats(null);
      return;
    }
    try {
      setStats(await api.get<MyStats>('/api/me/stats'));
    } catch {
      setStats(null);
    }
  }, [user?.id]);

  const loadPosts = useCallback(async (page: number) => {
    setLoading(true);
    try {
      setPosts(await api.get<Page<Blog>>(`/api/me/blogs?page=${page}&pageSize=${PAGE_SIZE}`));
    } catch {
      setPosts(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadFavorites = useCallback(async () => {
    try {
      setFavorites(await api.get<Place[]>('/api/favorites'));
    } catch {
      setFavorites([]);
    }
  }, []);

  useEffect(() => {
    void loadStats();
    void loadFavorites();
  }, [loadStats, loadFavorites]);

  useEffect(() => {
    if (user) void loadPosts(postPage);
  }, [user?.id, postPage, loadPosts]);

  const openEdit = () => {
    if (!user) return;
    setNickname(user.nickname);
    setAvatarUrl(user.avatarUrl || '');
    setEditOpen(true);
  };

  const uploadAvatar = async (options: any) => {
    try {
      const url = await uploadImage(options.file as File, 'avatar');
      setAvatarUrl(url);
      options.onSuccess?.({ url });
    } catch (e) {
      options.onError?.(e as Error);
      message.error((e as Error).message);
    }
  };

  const beforeAvatar = (file: File) => {
    if (!file.type.startsWith('image/')) {
      message.error(t('common.onlyImages'));
      return Upload.LIST_IGNORE;
    }
    if (file.size > 5 * 1024 * 1024) {
      message.error(t('common.imageTooLarge'));
      return Upload.LIST_IGNORE;
    }
    return true;
  };

  const saveProfile = async () => {
    const name = nickname.trim();
    if (!name) {
      message.error(t('auth.nickname'));
      return;
    }
    setSaving(true);
    try {
      await api.put('/api/me', { nickname: name, avatarUrl });
      await refresh();
      message.success(t('profile.saved'));
      setEditOpen(false);
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const toggleLike = async (blog: Blog) => {
    if (!user) return;
    try {
      const res = blog.likedByMe
        ? await api.del<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`)
        : await api.post<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`);
      setPosts((prev) =>
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

  const removePost = async (blog: Blog) => {
    try {
      await api.del(`/api/blogs/${blog.id}`);
      message.success(t('blog.deleteSuccess'));
      void loadPosts(postPage);
      void loadStats();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const removeFavorite = async (place: Place) => {
    try {
      await api.del(`/api/favorites/${place.id}`);
      message.success(t('profile.saved'));
      void loadFavorites();
      void loadStats();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const switchLanguage = async () => {
    const next = i18n.language === 'zh' ? 'en' : 'zh';
    void i18n.changeLanguage(next);
    if (user) {
      await api.put('/api/me', { language: next }).catch(() => undefined);
    }
  };

  if (!user) {
    return (
      <div>
        <div className="section-title">{t('profile.title')}</div>
        <Card>
          <div className="profile-hero-main">
            <Avatar size={64}>?</Avatar>
            <div className="profile-hero-text">
              <b>{t('profile.guest')}</b>
              <div className="author-sub">{t('profile.loginToView')}</div>
            </div>
            <Button type="primary" onClick={openAuth}>
              {t('common.login')}
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <div className="section-title">{t('profile.title')}</div>

      <div className="profile-layout">
        <div className="stack">
          <Card>
            <div className="profile-hero-main">
              <Avatar size={64} src={assetUrl(user.avatarUrl)}>
                {user.nickname?.[0]}
              </Avatar>
              <div className="profile-hero-text">
                <b>{user.nickname}</b>
                <div className="author-sub">@{user.username}</div>
              </div>
              <div className="profile-head-actions">
                <Button icon={<EditOutlined />} onClick={openEdit}>
                  {t('profile.editProfile')}
                </Button>
                <Button onClick={signOut}>{t('common.logout')}</Button>
              </div>
            </div>
          </Card>

          <Card>
            {stats ? (
              <div className="profile-stats">
                <div>
                  <b>{stats.favorites}</b>
                  <span>{t('profile.favorites')}</span>
                </div>
                <div>
                  <b>{stats.routes}</b>
                  <span>{t('profile.history')}</span>
                </div>
                <div>
                  <b>{stats.posts}</b>
                  <span>{t('profile.myPosts')}</span>
                </div>
                <div>
                  <b>{stats.followers}</b>
                  <span>{t('profile.followers')}</span>
                </div>
                <div>
                  <b>{stats.following}</b>
                  <span>{t('profile.following')}</span>
                </div>
              </div>
            ) : (
              <Skeleton active paragraph={{ rows: 2 }} />
            )}
          </Card>

          <Button block icon={<GlobalOutlined />} onClick={switchLanguage}>
            {i18n.language === 'zh' ? t('profile.switchToEn') : t('profile.switchToZh')}
          </Button>
        </div>

        <Card>
          <Tabs
            activeKey={tab}
            onChange={setTab}
            items={[
              {
                key: 'posts',
                label: `${t('profile.myPosts')} (${stats?.posts ?? 0})`,
                children: loading ? (
                  <Skeleton active />
                ) : !posts || posts.items.length === 0 ? (
                  <Empty description={t('user.noPosts')} />
                ) : (
                  <>
                    <div className="blog-grid">
                      {posts.items.map((b) => (
                        <BlogCard
                          key={b.id}
                          blog={b}
                          currentUserId={user.id}
                          onLike={toggleLike}
                          onEdit={(item) => {
                            setEditing(item);
                            setEditorOpen(true);
                          }}
                          onDelete={removePost}
                        />
                      ))}
                    </div>
                    <div className="pagination-row">
                      <Button
                        disabled={postPage <= 1}
                        onClick={() => setPostPage((p) => Math.max(1, p - 1))}
                      >
                        {t('common.back')}
                      </Button>
                      <span>
                        {posts.page} / {Math.max(1, Math.ceil(posts.total / posts.pageSize))}
                      </span>
                      <Button
                        disabled={posts.page * posts.pageSize >= posts.total}
                        onClick={() => setPostPage((p) => p + 1)}
                      >
                        {t('blog.readMore')}
                      </Button>
                    </div>
                  </>
                ),
              },
              {
                key: 'favorites',
                label: `${t('profile.myFavorites')} (${favorites.length})`,
                children:
                  favorites.length === 0 ? (
                    <Empty description={t('profile.loginToView')} />
                  ) : (
                    <div className="favorite-list">
                      {favorites.map((p) => (
                        <div key={p.id} className="favorite-item">
                          <div>
                            <b>{i18n.language === 'zh' ? p.nameCn : p.nameEn}</b>
                            <div className="author-sub">
                              {i18n.language === 'zh' ? p.descriptionCn : p.descriptionEn}
                            </div>
                          </div>
                          <Button size="small" onClick={() => void removeFavorite(p)}>
                            {t('profile.removeFavorite')}
                          </Button>
                        </div>
                      ))}
                    </div>
                  ),
              },
            ]}
          />
        </Card>
      </div>

      <Modal
        title={t('profile.editProfile')}
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={saveProfile}
        confirmLoading={saving}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form layout="vertical">
          <Form.Item label={t('profile.avatar')}>
            <div className="avatar-edit">
              <Avatar size={72} src={assetUrl(avatarUrl)}>
                {nickname?.[0]}
              </Avatar>
              <Upload showUploadList={false} accept="image/*" customRequest={uploadAvatar} beforeUpload={beforeAvatar}>
                <Button icon={<UploadOutlined />}>{t('profile.avatar')}</Button>
              </Upload>
            </div>
          </Form.Item>
          <Form.Item label={t('profile.nickname')}>
            <Input
              value={nickname}
              maxLength={50}
              showCount
              onChange={(e) => setNickname(e.target.value)}
            />
          </Form.Item>
        </Form>
      </Modal>

      <BlogEditor
        open={editorOpen}
        initial={editing}
        onClose={() => setEditorOpen(false)}
        onSaved={() => {
          setEditorOpen(false);
          setPostPage(1);
          void loadPosts(1);
          void loadStats();
        }}
      />
    </div>
  );
}
