import { getLanguage, getTranslations } from '../../../utils/i18n';
import { guideChat } from '../../../services/api';
import { getToken } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';

Page({
  data: {
    inputValue: '',
    messages: [] as any[],
    toView: '',
    sessionId: 0,
    t: {} as any
  },
  onShow() {
    const lang = getLanguage();
    const t = getTranslations(lang);
    this.setData({ t });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Customer Service AI' : 'AI 客服'
    });
    
    // Only set initial message if empty
    if (this.data.messages.length === 0) {
      this.setData({
        messages: [
          { text: t.chatbot.initialMsg, isUser: false }
        ]
      });
    }
  },
  onInputChange(e: any) {
    this.setData({ inputValue: e.detail.value });
  },
  async onSend() {
    const text = this.data.inputValue;
    if (!text.trim()) return;

    if (!getToken()) {
      wx.showModal({
        title: '需要登录',
        content: 'AI 客服需要先登录，是否前往登录页？',
        confirmText: '去登录',
        success: (res) => {
          if (res.confirm) gotoLogin();
        }
      });
      return;
    }

    const newMessages = [...this.data.messages, { text, isUser: true }];
    this.setData({
      messages: newMessages,
      inputValue: '',
      toView: `msg-${newMessages.length - 1}`
    });

    try {
      // The backend keeps the session; we only track the id.
      const resp = await guideChat(this.data.sessionId, text.trim());
      const reply = { text: resp.reply, isUser: false };
      this.setData({
        messages: [...this.data.messages, reply],
        sessionId: resp.sessionId || this.data.sessionId,
        toView: `msg-${this.data.messages.length}`
      });
    } catch (err) {
      console.error('Guide chat error:', err);
      const errorReply = { text: this.data.t.chatbot.networkError, isUser: false };
      this.setData({
        messages: [...this.data.messages, errorReply],
        toView: `msg-${this.data.messages.length}`
      });
    }
  }
})
