import { useState } from 'react';
import { Button, Form, Input, Modal, Tabs, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';

interface LoginValues {
  username: string;
  password: string;
  nickname?: string;
}

// Global login/register modal; any page can open it with useAuth().openAuth().
export default function AuthModal() {
  const { authOpen, closeAuth, signIn, signUp } = useAuth();
  const { t } = useTranslation();
  const [tab, setTab] = useState('login');
  const [submitting, setSubmitting] = useState(false);
  const [loginForm] = Form.useForm<LoginValues>();
  const [registerForm] = Form.useForm<LoginValues>();

  const finish = async (action: () => Promise<void>, reset: () => void) => {
    setSubmitting(true);
    try {
      await action();
      reset();
      closeAuth();
      message.success(t('auth.success'));
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  const login = (values: LoginValues) =>
    finish(
      () => signIn(values.username, values.password),
      () => loginForm.resetFields(),
    );
  const register = (values: LoginValues) =>
    finish(
      () => signUp(values.username, values.password, values.nickname),
      () => registerForm.resetFields(),
    );

  return (
    <Modal
      open={authOpen}
      onCancel={closeAuth}
      footer={null}
      title={t('auth.title')}
      destroyOnHidden
    >
      <Tabs
        activeKey={tab}
        onChange={setTab}
        items={[
          {
            key: 'login',
            label: t('auth.loginTab'),
            children: (
              <Form form={loginForm} layout="vertical" onFinish={login}>
                <Form.Item name="username" label={t('auth.username')} rules={[{ required: true, min: 3 }]}>
                  <Input autoComplete="username" />
                </Form.Item>
                <Form.Item name="password" label={t('auth.passwordLabel')} rules={[{ required: true, min: 6 }]}>
                  <Input.Password autoComplete="current-password" />
                </Form.Item>
                <Button type="primary" block htmlType="submit" loading={submitting}>
                  {t('auth.signIn')}
                </Button>
              </Form>
            ),
          },
          {
            key: 'register',
            label: t('auth.registerTab'),
            children: (
              <Form form={registerForm} layout="vertical" onFinish={register}>
                <Form.Item name="username" label={t('auth.username')} rules={[{ required: true, min: 3 }]}>
                  <Input autoComplete="username" />
                </Form.Item>
                <Form.Item name="password" label={t('auth.passwordLabel')} rules={[{ required: true, min: 6 }]}>
                  <Input.Password autoComplete="new-password" />
                </Form.Item>
                <Form.Item name="nickname" label={t('auth.nickname')}>
                  <Input />
                </Form.Item>
                <Button type="primary" block htmlType="submit" loading={submitting}>
                  {t('auth.signUp')}
                </Button>
              </Form>
            ),
          },
        ]}
      />
    </Modal>
  );
}
