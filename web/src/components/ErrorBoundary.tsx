import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Button, Result } from 'antd';
import i18n from '../i18n';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  message: string;
}

// Catches render errors so a data edge case never turns into a white screen.
export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: '' };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error?.message || String(error) };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Render error:', error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <Result
          status="error"
          title={i18n.t('common.error')}
          subTitle={this.state.message}
          extra={
            <Button type="primary" onClick={() => window.location.reload()}>
              {i18n.t('common.reload')}
            </Button>
          }
        />
      );
    }
    return this.props.children;
  }
}
