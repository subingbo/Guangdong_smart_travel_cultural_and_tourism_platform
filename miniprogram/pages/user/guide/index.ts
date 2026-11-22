import { getLanguage, getTranslations } from '../../../utils/i18n';
import { guideChat } from '../../../services/api';
import { getToken } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';

Page({
  data: {
    inputValue: '',
    messages: [] as any[],
    toView: '',
    t: {} as any,
    loading: false,
    sessionId: 0,
    mapCenter: {
      latitude: 23.1291,
      longitude: 113.2644
    },
    mapScale: 14,
    mapMarkers: [] as any[]
  },
  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/user/guide/index'
      })
    }
    const lang = getLanguage();
    const t = getTranslations(lang);
    this.setData({ t });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'AI Guide' : 'AI 导游'
    });

    this.initMap();
  },
  initMap() {
    const lang = getLanguage();
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        const center = { latitude: res.latitude, longitude: res.longitude };
        this.setData({
          mapCenter: center,
          mapMarkers: [{
            id: 0,
            latitude: res.latitude,
            longitude: res.longitude,
            width: 40,
            height: 40
          }]
        });

        if (this.data.messages.length === 0) {
          this.showWelcome(lang);
        }
      },
      fail: () => {
        // Devtools often denies location; keep the Guangzhou default center.
        const center = this.data.mapCenter;
        this.setData({
          mapMarkers: [{
            id: 0,
            latitude: center.latitude,
            longitude: center.longitude,
            width: 40,
            height: 40
          }]
        });

        if (this.data.messages.length === 0) {
          this.showWelcome(lang);
        }
      }
    });
  },
  showWelcome(lang: string) {
    const welcome = lang === 'zh'
      ? '你好！我是你的 AI 导游，欢迎来到广东。想了解早茶怎么吃、去哪里看夜景，还是规划一条一日游路线？直接问我吧。'
      : 'Hi! I am your AI guide for Guangdong. Ask me about dim sum, night views, or a one-day route plan.';
    this.setData({
      messages: [{ text: welcome, isUser: false }]
    });
  },
  requireLogin(): boolean {
    if (getToken()) return true;
    wx.showModal({
      title: '需要登录',
      content: 'AI 导游对话需要先登录，是否前往登录页？',
      confirmText: '去登录',
      success: (res) => {
        if (res.confirm) gotoLogin();
      }
    });
    return false;
  },
  onInputChange(e: any) {
    this.setData({ inputValue: e.detail.value });
  },
  async onSend() {
    const text = this.data.inputValue;
    if (!text.trim() || this.data.loading) return;
    if (!this.requireLogin()) return;

    const newMessages = [...this.data.messages, { text, isUser: true }];
    this.setData({
      messages: [...newMessages, { text: 'Thinking...', isUser: false, isThinking: true }],
      inputValue: '',
      toView: `msg-${newMessages.length}`,
      loading: true
    });

    try {
      const resp = await guideChat(this.data.sessionId, text.trim());
      const cleanMessages = this.data.messages.filter(m => !m.isThinking);
      this.setData({
        messages: [...cleanMessages, { text: resp.reply, isUser: false }],
        sessionId: resp.sessionId || this.data.sessionId,
        toView: `msg-${cleanMessages.length}`
      });
    } catch (error) {
      console.error('AI Chat Error:', error);
      const cleanMessages = this.data.messages.filter(m => !m.isThinking);
      this.setData({
        messages: [...cleanMessages, {
          text: 'Sorry, I encountered an error. Please try again.',
          isUser: false,
          isError: true
        }],
        toView: `msg-${cleanMessages.length}`
      });
    } finally {
      this.setData({ loading: false });
    }
  }
})
