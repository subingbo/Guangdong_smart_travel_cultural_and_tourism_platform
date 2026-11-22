import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  Alert,
  Avatar,
  Button,
  Card,
  Empty,
  Image,
  Input,
  Modal,
  Pagination,
  Skeleton,
  message,
} from 'antd';
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  EditOutlined,
  HeartFilled,
  HeartOutlined,
  SendOutlined,
  UserAddOutlined,
  UserDeleteOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, Blog, Comment, FALLBACK_IMAGE, Page, UserProfile } from '../api/client';
import { useAuth } from '../context/AuthContext';
import BlogEditor from '../components/BlogEditor';

function formatDate(value: string, locale: string) {
  return new Date(value).toLocaleString(locale === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function BlogDetailPage() {
  const { id } = useParams<{ id: string }>();
  const blogId = Number(id);
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const { user, openAuth } = useAuth();

  const [blog, setBlog] = useState<Blog | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [author, setAuthor] = useState<UserProfile | null>(null);
  const [comments, setComments] = useState<Page<Comment> | null>(null);
  const [commentPage, setCommentPage] = useState(1);
  const [commentText, setCommentText] = useState('');
  const [sending, setSending] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);

  const loadBlog = useCallback(async () => {
    if (!blogId) return;
    setLoading(true);
    setError(null);
    try {
      setBlog(await api.get<Blog>(`/api/blogs/${blogId}`));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [blogId]);

  const loadComments = useCallback(async () => {
    if (!blogId) return;
    try {
      setComments(
        await api.get<Page<Comment>>(`/api/blogs/${blogId}/comments?page=${commentPage}&pageSize=20`),
      );
    } catch {
      setComments(null);
    }
  }, [blogId, commentPage]);

  useEffect(() => {
    void loadBlog();
  }, [loadBlog]);

  useEffect(() => {
    void loadComments();
  }, [loadComments]);

  useEffect(() => {
    if (!blog) return;
    api
      .get<UserProfile>(`/api/users/${blog.userId}`)
      .then(setAuthor)
      .catch(() => setAuthor(null));
  }, [blog?.userId]);

  const toggleLike = async () => {
    if (!blog) return;
    if (!user) {
      openAuth();
      return;
    }
    try {
      const res = blog.likedByMe
        ? await api.del<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`)
        : await api.post<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`);
      setBlog({ ...blog, likes: res.likes, likedByMe: res.likedByMe });
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const toggleFollow = async () => {
    if (!author) return;
    if (!user) {
      openAuth();
      return;
    }
    try {
      if (author.isFollowing) {
        await api.del(`/api/users/${author.user.id}/follow`);
      } else {
        await api.post(`/api/users/${author.user.id}/follow`);
      }
      setAuthor({
        ...author,
        isFollowing: !author.isFollowing,
        stats: {
          ...author.stats,
          followers: Math.max(0, author.stats.followers + (author.isFollowing ? -1 : 1)),
        },
      });
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const removeBlog = () => {
    if (!blog) return;
    Modal.confirm({
      title: t('blog.deletePost'),
      content: t('blog.deleteConfirm'),
      okText: t('common.delete'),
      cancelText: t('common.cancel'),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await api.del(`/api/blogs/${blog.id}`);
          message.success(t('blog.deleteSuccess'));
          navigate('/blog');
        } catch (e) {
          message.error((e as Error).message);
        }
      },
    });
  };

  const submitComment = async () => {
    if (!user) {
      openAuth();
      return;
    }
    const content = commentText.trim();
    if (!content) return;
    setSending(true);
    try {
      await api.post(`/api/blogs/${blogId}/comments`, { content });
      setCommentText('');
      message.success(t('blog.commentSuccess'));
      setCommentPage(1);
      await loadComments();
      await loadBlog();
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setSending(false);
    }
  };

  const removeComment = (c: Comment) => {
    Modal.confirm({
      title: t('common.delete'),
      content: t('common.deleteConfirm'),
      okText: t('common.delete'),
      cancelText: t('common.cancel'),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await api.del(`/api/comments/${c.id}`);
          await loadComments();
          await loadBlog();
        } catch (e) {
          message.error((e as Error).message);
        }
      },
    });
  };

  const canDeleteComment = (c: Comment) =>
    !!user && !!blog && (user.id === c.userId || user.id === blog.userId);

  if (loading) return <Skeleton active />;
  if (error || !blog) {
    return (
      <Alert
        type="error"
        showIcon
        message={error || t('common.error')}
        action={
          <Button size="small" onClick={() => void loadBlog()}>
            {t('common.retry')}
          </Button>
        }
      />
    );
  }

  const locale = i18n.language;

  return (
    <div className="blog-detail">
      <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/blog')}>
        {t('common.back')}
      </Button>

      <Card className="blog-detail-card">
        <h1 className="blog-detail-title">{blog.title}</h1>

        <div className="blog-detail-meta">
          <Link to={`/user/${blog.userId}`} className="blog-author">
            <Avatar size={32} src={assetUrl(blog.author?.avatarUrl)}>
              {blog.author?.nickname?.[0]}
            </Avatar>
            <span>{blog.author?.nickname}</span>
          </Link>
          <span>{formatDate(blog.createdAt, locale)}</span>
        </div>

        {blog.images.length > 0 && (
          <Image.PreviewGroup>
            <div className="blog-detail-images">
              {blog.images.map((src, i) => (
                <Image
                  key={i}
                  src={assetUrl(src)}
                  fallback={FALLBACK_IMAGE}
                  alt={`${blog.title}-${i + 1}`}
                />
              ))}
            </div>
          </Image.PreviewGroup>
        )}

        <div className="blog-detail-content">{blog.content}</div>

        <div className="blog-detail-actions">
          <Button
            type={blog.likedByMe ? 'primary' : 'default'}
            icon={blog.likedByMe ? <HeartFilled /> : <HeartOutlined />}
            onClick={toggleLike}
          >
            {blog.likes}
          </Button>
          <span className="blog-comment-count">
            {t('blog.comments')}: {blog.commentCount}
          </span>
          {user?.id === blog.userId && (
            <span className="blog-owner-actions">
              <Button icon={<EditOutlined />} onClick={() => setEditorOpen(true)}>
                {t('common.edit')}
              </Button>
              <Button danger icon={<DeleteOutlined />} onClick={removeBlog}>
                {t('common.delete')}
              </Button>
            </span>
          )}
        </div>
      </Card>

      <Card className="author-card">
        <div className="author-row">
          <Link to={`/user/${blog.userId}`}>
            <Avatar size={48} src={assetUrl(author?.user.avatarUrl || blog.author?.avatarUrl)}>
              {author?.user.nickname?.[0] || blog.author?.nickname?.[0]}
            </Avatar>
          </Link>
          <div className="author-info">
            <Link to={`/user/${blog.userId}`}>
              <b>{author?.user.nickname || blog.author?.nickname}</b>
            </Link>
            <div className="author-sub">
              {author ? `${author.stats.posts} ${t('user.posts')} · ${author.stats.followers} ${t('user.followers')}` : ''}
            </div>
          </div>
          {user?.id !== blog.userId && (
            <Button
              type={author?.isFollowing ? 'default' : 'primary'}
              icon={author?.isFollowing ? <UserDeleteOutlined /> : <UserAddOutlined />}
              onClick={toggleFollow}
            >
              {author?.isFollowing ? t('user.unfollow') : t('user.follow')}
            </Button>
          )}
        </div>
      </Card>

      <Card className="comments-card">
        <div className="section-title" style={{ marginTop: 0 }}>
          {t('blog.comments')} ({blog.commentCount})
        </div>

        {!comments || comments.items.length === 0 ? (
          <Empty description={t('blog.noComments')} />
        ) : (
          <div className="comment-list">
            {comments.items.map((c) => (
              <div key={c.id} className="comment-item">
                <Link to={`/user/${c.userId}`}>
                  <Avatar size={32} src={assetUrl(c.author?.avatarUrl)}>
                    {c.author?.nickname?.[0]}
                  </Avatar>
                </Link>
                <div className="comment-main">
                  <div className="comment-head">
                    <Link to={`/user/${c.userId}`}>{c.author?.nickname}</Link>
                    <span>{formatDate(c.createdAt, locale)}</span>
                  </div>
                  <div className="comment-content">{c.content}</div>
                </div>
                {canDeleteComment(c) && (
                  <Button
                    type="text"
                    danger
                    size="small"
                    icon={<DeleteOutlined />}
                    onClick={() => removeComment(c)}
                  />
                )}
              </div>
            ))}
          </div>
        )}

        {comments && comments.total > comments.pageSize && (
          <div className="pagination-row">
            <Pagination
              current={comments.page}
              pageSize={comments.pageSize}
              total={comments.total}
              showSizeChanger={false}
              onChange={(p) => setCommentPage(p)}
            />
          </div>
        )}

        <div className="comment-form">
          {user ? (
            <>
              <Input.TextArea
                rows={3}
                value={commentText}
                maxLength={1000}
                showCount
                placeholder={t('blog.commentPlaceholder')}
                onChange={(e) => setCommentText(e.target.value)}
              />
              <Button
                type="primary"
                icon={<SendOutlined />}
                loading={sending}
                disabled={!commentText.trim()}
                onClick={submitComment}
              >
                {t('blog.submitComment')}
              </Button>
            </>
          ) : (
            <Button onClick={openAuth}>{t('blog.loginToComment')}</Button>
          )}
        </div>
      </Card>

      <BlogEditor
        open={editorOpen}
        initial={blog}
        onClose={() => setEditorOpen(false)}
        onSaved={() => {
          setEditorOpen(false);
          void loadBlog();
        }}
      />
    </div>
  );
}
