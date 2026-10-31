import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';
import { useTranslation } from 'react-i18next';
import App from './App';
import 'antd/dist/reset.css';
import './index.css';
import './i18n';

function Root() {
  const { i18n } = useTranslation();
  return (
    <ConfigProvider
      locale={i18n.language === 'zh' ? zhCN : enUS}
      theme={{ token: { colorPrimary: '#0052d9', borderRadius: 10 } }}
    >
      <App />
    </ConfigProvider>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);
