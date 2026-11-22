import { Avatar, Button, Dropdown, Modal } from 'antd';
import {
  CommentOutlined,
  DeleteOutlined,
  EditOutlined,
  HeartFilled,
  HeartOutlined,
  MoreOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { assetUrl } from '../api/client';
import type { Blog } from '../api/client';
import SafeImage from './SafeImage';

interface Props {
  blog: Blog;
  currentUserId?: number;
  onLike: (blog: Blog) => void;
  onEdit?: (blog: Blog) => void;
  onDelete?: (blog: Blog) => void;
}

function excerpt(text: string, max = 150) {
  const clean = text.replace(/\s+/g, ' ').trim();
  return clean.length > max ? `${clean.slice(0, max)}…` : clean;
}

export default function BlogCard({ blog, currentUserId, onLike, onEdit, onDelete }: Props) {
  const { t, i18n } = useTranslation();
  const date = new Date(blog.createdAt).toLocaleDateString(
    i18n.language === 'zh' ? 'zh-CN' : 'en-US',
  );
  const cover = blog.images?.[0];
  const isOwner = !!currentUserId && currentUserId === blog.userId;

  const confirmDelete = () => {
    Modal.confirm({
      title: t('blog.deletePost'),
      content: t('blog.deleteConfirm'),
      okText: t('common.delete'),
      cancelText: t('common.cancel'),
      okButtonProps: { danger: true },
      onOk: () => onDelete?.(blog),
    });
  };

  return (
    <article className="blog-card-v2">
      <Link to={`/blog/${blog.id}`} className="blog-cover">
        <SafeImage
          src={cover}
          alt={blog.title}
          loading="lazy"
          fallback={<div className="blog-cover-fallback">粤</div>}
        />
      </Link>

      <div className="blog-body">
        <Link to={`/blog/${blog.id}`} className="blog-title-link">
          <h3>{blog.title}</h3>
        </Link>
        <p className="blog-excerpt">{excerpt(blog.content)}</p>

        <div className="blog-meta">
          <Link to={`/user/${blog.userId}`} className="blog-author">
            <Avatar size={24} src={assetUrl(blog.author?.avatarUrl)}>
              {blog.author?.nickname?.[0]}
            </Avatar>
            <span>{blog.author?.nickname || '—'}</span>
          </Link>
          <span className="blog-date">{date}</span>
        </div>

        <div className="blog-actions">
          <Button
            size="small"
            type={blog.likedByMe ? 'primary' : 'text'}
            icon={blog.likedByMe ? <HeartFilled /> : <HeartOutlined />}
            onClick={() => onLike(blog)}
          >
            {blog.likes}
          </Button>
          <Link to={`/blog/${blog.id}`} className="blog-comment-link">
            <CommentOutlined /> {blog.commentCount}
          </Link>
          {isOwner && (
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  {
                    key: 'edit',
                    icon: <EditOutlined />,
                    label: t('blog.editPost'),
                  },
                  {
                    key: 'delete',
                    icon: <DeleteOutlined />,
                    label: t('blog.deletePost'),
                    danger: true,
                  },
                ],
                onClick: ({ key }) => {
                  if (key === 'edit') onEdit?.(blog);
                  if (key === 'delete') confirmDelete();
                },
              }}
            >
              <Button type="text" size="small" icon={<MoreOutlined />} />
            </Dropdown>
          )}
        </div>
      </div>
    </article>
  );
}
