import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    inputValue: '',
    messages: [] as any[],
    toView: '',
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
    
    const newMessages = [...this.data.messages, { text, isUser: true }];
    this.setData({
      messages: newMessages,
      inputValue: '',
      toView: `msg-${newMessages.length - 1}`
    });
    
    try {
      const apiMessages = newMessages.map(m => ({
        role: m.isUser ? 'user' : 'assistant',
        content: m.text
      }));

      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'chat',
          payload: {
            messages: apiMessages
          }
        }
      });

      const result = res.result as any;
      if (result.success && result.data && result.data.content) {
        const reply = { text: result.data.content, isUser: false };
        this.setData({
          messages: [...this.data.messages, reply],
          toView: `msg-${this.data.messages.length}`
        });
      } else {
        console.error('AI Error:', result);
        const errorReply = { text: this.data.t.chatbot.error, isUser: false };
        this.setData({
          messages: [...this.data.messages, errorReply],
          toView: `msg-${this.data.messages.length}`
        });
      }
    } catch (err) {
      console.error('Cloud Function Error:', err);
      const errorReply = { text: this.data.t.chatbot.networkError, isUser: false };
      this.setData({
        messages: [...this.data.messages, errorReply],
        toView: `msg-${this.data.messages.length}`
      });
    }
  }
})
