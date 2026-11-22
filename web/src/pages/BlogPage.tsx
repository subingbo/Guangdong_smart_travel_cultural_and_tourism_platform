import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Empty, Input, Pagination, Skeleton, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, Blog, Page } from '../api/client';
import { useAuth } from '../context/AuthContext';
import BlogCard from '../components/BlogCard';
import BlogEditor from '../components/BlogEditor';

const PAGE_SIZE = 9;

export default function BlogPage() {
  const { t } = useTranslation();
  const { user, openAuth } = useAuth();
  const [data, setData] = useState<Page<Blog> | null>(null);
  const [page, setPage] = useState(1);
  const [q, setQ] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<Blog | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.get<Page<Blog>>(
        `/api/blogs?page=${page}&pageSize=${PAGE_SIZE}&q=${encodeURIComponent(q)}`,
      );
      setData(res);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [page, q]);

  useEffect(() => {
    void load();
  }, [load]);

  const toggleLike = async (blog: Blog) => {
    if (!user) {
      openAuth();
      return;
    }
    try {
      const res = blog.likedByMe
        ? await api.del<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`)
        : await api.post<{ likes: number; likedByMe: boolean }>(`/api/blogs/${blog.id}/like`);
      setData((prev) =>
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

  const remove = async (blog: Blog) => {
    try {
      await api.del(`/api/blogs/${blog.id}`);
      message.success(t('blog.deleteSuccess'));
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const openEditor = (blog: Blog | null) => {
    if (!user) {
      openAuth();
      return;
    }
    setEditing(blog);
    setEditorOpen(true);
  };

  return (
    <div>
      <div className="blog-toolbar">
        <div className="section-title" style={{ margin: 0 }}>
          {t('blog.title')}
        </div>
        <div className="blog-toolbar-actions">
          <Input.Search
            allowClear
            placeholder={t('blog.searchPlaceholder')}
            onSearch={(value) => {
              setQ(value.trim());
              setPage(1);
            }}
            style={{ width: 240 }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor(null)}>
            {t('blog.newPost')}
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
        <Empty description={t('blog.empty')} />
      ) : (
        <>
          <div className="blog-grid">
            {data.items.map((b) => (
              <BlogCard
                key={b.id}
                blog={b}
                currentUserId={user?.id}
                onLike={toggleLike}
                onEdit={(item) => openEditor(item)}
                onDelete={remove}
              />
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

      <BlogEditor
        open={editorOpen}
        initial={editing}
        onClose={() => setEditorOpen(false)}
        onSaved={() => {
          setEditorOpen(false);
          setPage(1);
          void load();
        }}
      />
    </div>
  );
}
