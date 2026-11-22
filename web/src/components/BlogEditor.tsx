import { useEffect, useState } from 'react';
import { Alert, Button, Form, Input, Modal, Upload, message } from 'antd';
import { PlusOutlined, RobotOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, Blog, uploadImage } from '../api/client';

interface Props {
  open: boolean;
  initial: Blog | null;
  onClose: () => void;
  onSaved: () => void;
}

interface FormValues {
  title: string;
  content: string;
}

// Create/edit modal with AI draft and image upload.
export default function BlogEditor({ open, initial, onClose, onSaved }: Props) {
  const { t } = useTranslation();
  const [form] = Form.useForm<FormValues>();
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [saving, setSaving] = useState(false);
  const [topic, setTopic] = useState('');
  const [drafting, setDrafting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    form.setFieldsValue({
      title: initial?.title || '',
      content: initial?.content || '',
    });
    setFileList(
      (initial?.images || []).map((url, i): UploadFile => ({
        uid: `existing-${i}-${url}`,
        name: `image-${i + 1}`,
        status: 'done',
        // Preview through the same-origin asset URL, but keep the portable
        // /uploads/... path in response.url so edits never persist a host.
        url: assetUrl(url),
        response: { url },
      })),
    );
    setTopic('');
    setError(null);
  }, [open, initial, form]);

  const customUpload = async (options: any) => {
    const { file, onSuccess, onError } = options;
    try {
      const url = await uploadImage(file as File, 'blog');
      onSuccess?.({ url });
    } catch (e) {
      onError?.(e as Error);
      message.error((e as Error).message);
    }
  };

  const beforeUpload = (file: File) => {
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

  const generateDraft = async () => {
    if (!topic.trim()) return;
    setDrafting(true);
    setError(null);
    try {
      const draft = await api.post<{ title: string; content: string }>('/api/blogs/draft', {
        topic: topic.trim(),
      });
      form.setFieldsValue({ title: draft.title, content: draft.content });
      message.success(t('blog.useDraft'));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setDrafting(false);
    }
  };

  const submit = async () => {
    const values = await form.validateFields();
    const images = fileList
      // response.url is the portable path returned by the upload API; fall
      // back to url only for entries that already contain it.
      .map((f) => (f.response as { url?: string } | undefined)?.url || f.url)
      .filter((v): v is string => !!v);
    setSaving(true);
    setError(null);
    try {
      const payload = { title: values.title, content: values.content, images };
      if (initial) {
        await api.put(`/api/blogs/${initial.id}`, payload);
        message.success(t('blog.updateSuccess'));
      } else {
        await api.post('/api/blogs', payload);
        message.success(t('blog.publishSuccess'));
      }
      onSaved();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      title={initial ? t('blog.editPost') : t('blog.newPost')}
      open={open}
      onCancel={onClose}
      onOk={submit}
      okText={t('common.save')}
      cancelText={t('common.cancel')}
      confirmLoading={saving}
      width={720}
      destroyOnHidden
    >
      {error && <Alert type="error" showIcon message={error} style={{ marginBottom: 12 }} />}

      <div className="draft-bar">
        <Input
          value={topic}
          onChange={(e) => setTopic(e.target.value)}
          placeholder={t('blog.draftTopic')}
          onPressEnter={generateDraft}
        />
        <Button icon={<RobotOutlined />} onClick={generateDraft} loading={drafting}>
          {t('blog.draftGenerate')}
        </Button>
      </div>

      <Form form={form} layout="vertical">
        <Form.Item
          name="title"
          label={t('blog.titleLabel')}
          rules={[{ required: true, max: 200 }]}
        >
          <Input maxLength={200} showCount />
        </Form.Item>
        <Form.Item
          name="content"
          label={t('blog.contentLabel')}
          rules={[{ required: true, max: 20000 }]}
        >
          <Input.TextArea rows={10} maxLength={20000} showCount />
        </Form.Item>
        <Form.Item label={t('blog.imagesLabel')} extra={t('blog.uploadHint')}>
          <Upload
            listType="picture-card"
            accept="image/*"
            multiple
            maxCount={9}
            fileList={fileList}
            customRequest={customUpload}
            beforeUpload={beforeUpload}
            onChange={({ fileList: next }) => setFileList(next)}
          >
            {fileList.length >= 9 ? null : (
              <div>
                <PlusOutlined />
                <div style={{ marginTop: 6 }}>{t('blog.imagesLabel')}</div>
              </div>
            )}
          </Upload>
        </Form.Item>
      </Form>
    </Modal>
  );
}
