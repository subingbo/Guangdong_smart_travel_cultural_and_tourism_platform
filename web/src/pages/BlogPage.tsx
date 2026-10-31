import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Empty,
  Form,
  Input,
  Modal,
  Skeleton,
  message,
} from 'antd';
import { RobotOutlined, PlusOutlined } from '@ant-design/icons';
import { api, Blog } from '../api/client';
import { useAuth } from '../context/AuthContext';

export default function BlogPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [blogs, setBlogs] = useState<Blog[]>([]);
  const [loading, setLoading] = useState(true);
  const [publishOpen, setPublishOpen] = useState(false);
  const [draftOpen, setDraftOpen] = useState(false);
  const [drafting, setDrafting] = useState(false);
  const [draft, setDraft] = useState<{ title: string; content: string } | null>(null);
  const [form] = Form.useForm();

  const load = () =>
    api
      .get<Blog[]>('/api/blogs')
      .then((data) => setBlogs(data || []))
      .catch(() => setBlogs([]))
      .finally(() => setLoading(false));

  useEffect(() => {
    void load();
  }, []);

  const publish = async (values: { title: string; content: string }) => {
    try {
      await api.post('/api/blogs', values);
      message.success(t('blog.publish'));
      setPublishOpen(false);
      form.resetFields();
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const generateDraft = async (topic: string) => {
    setDrafting(true);
    setDraft(null);
    try {
      setDraft(await api.post<{ title: string; content: string }>('/api/blogs/draft', { topic }));
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setDrafting(false);
    }
  };

  const like = async (blog: Blog) => {
    try {
      const { likes } = await api.post<{ likes: number }>(`/api/blogs/${blog.id}/like`);
      setBlogs((prev) => prev.map((b) => (b.id === blog.id ? { ...b, likes, likedByMe: true } : b)));
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  return (
    <div>
      <div
        className="section-title"
        style={{ display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}
      >
        {t('blog.title')}
        <span>
          <Button icon={<RobotOutlined />} onClick={() => setDraftOpen(true)}>
            {t('blog.aiDraft')}
          </Button>{' '}
          {user && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setPublishOpen(true)}>
              {t('blog.newPost')}
            </Button>
          )}
        </span>
      </div>

      {loading ? (
        <Skeleton active />
      ) : blogs.length === 0 ? (
        <Empty description={t('blog.empty')} />
      ) : (
        <div className="blog-grid">
          {blogs.map((b) => (
            <div key={b.id} className="blog-card">
              <h3 className="blog-title">{b.title}</h3>
              <p className="blog-content">{b.content}</p>
              {(b.images || []).map((img, i) => (
                <img key={i} src={img} alt={b.title} />
              ))}
              <div style={{ marginTop: 8 }}>
                <Button
                  size="small"
                  type={b.likedByMe ? 'primary' : 'default'}
                  onClick={() => like(b)}
                >
                  ♥ {b.likes}
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal
        title={t('blog.newPost')}
        open={publishOpen}
        onCancel={() => setPublishOpen(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={publish}>
          <Form.Item name="title" label="Title" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="content" label="Content" rules={[{ required: true }]}>
            <Input.TextArea rows={6} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('blog.aiDraft')}
        open={draftOpen}
        footer={null}
        onCancel={() => {
          setDraftOpen(false);
          setDraft(null);
        }}
      >
        <Input.Search
          placeholder={t('blog.draftTopic')}
          enterButton={t('blog.draftGenerate')}
          onSearch={generateDraft}
          loading={drafting}
        />
        {drafting && <Skeleton active style={{ marginTop: 16 }} />}
        {draft && (
          <Card style={{ marginTop: 16 }}>
            <b>{draft.title}</b>
            <p className="blog-content">{draft.content}</p>
          </Card>
        )}
      </Modal>
    </div>
  );
}
