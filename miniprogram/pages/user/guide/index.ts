import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    inputValue: '',
    messages: [] as any[],
    toView: '',
    t: {} as any,
    loading: false,
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
          this.fetchWelcomeMessage(center);
        }
      },
      fail: () => {
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
          this.fetchWelcomeMessage(center);
        }
      }
    });
  },
  async fetchWelcomeMessage(location: { latitude: number, longitude: number }) {
    this.setData({
      loading: true,
      messages: [{ text: 'Thinking...', isUser: false, isThinking: true }]
    });

    try {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'ai_guide_chat',
          payload: {
            type: 'welcome',
            context: {
              location
            }
          }
        }
      });

      if (res.result && (res.result as any).success) {
        const aiMsg = (res.result as any).data.content;
        this.setData({
          messages: [{ text: aiMsg, isUser: false }],
          loading: false
        });
      } else {
        throw new Error('AI Service Failed');
      }
    } catch (err) {
      console.error(err);
      this.setData({
        messages: [{ text: this.data.t.guide.initialMsg, isUser: false }],
        loading: false
      });
    }
  },
  onInputChange(e: any) {
    this.setData({ inputValue: e.detail.value });
  },
  async onSend() {
    const text = this.data.inputValue;
    if (!text.trim() || this.data.loading) return;
    
    const newMessages = [...this.data.messages, { text, isUser: true }];
    this.setData({
      messages: newMessages,
      inputValue: '',
      toView: `msg-${newMessages.length - 1}`,
      loading: true
    });
    
    // Add a placeholder for AI thinking
    this.setData({
      messages: [...newMessages, { text: 'Thinking...', isUser: false, isThinking: true }]
    });

    try {
      // Convert history to format expected by AI
      const historyMessages = this.data.messages.filter(m => !m.isThinking);
      // Exclude the last message (current user question) as it's sent separately
      const history = historyMessages.slice(0, -1).map(m => ({
        role: m.isUser ? 'user' : 'assistant',
        content: m.text
      }));

      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'ai_guide_chat',
          payload: {
            question: text,
            history: history.slice(-6), // Keep last 6 turns context
            context: {
              location: {
                latitude: this.data.mapCenter.latitude,
                longitude: this.data.mapCenter.longitude
              }
            }
          }
        }
      });

      const result = res.result as any;
      
      // Remove thinking message and add actual response
      const updatedMessages = [...this.data.messages]; // get latest
      // The thinking message is the last one we added, but user might have typed more? 
      // Safe to just append the response.
      
      // Actually, let's filter out the thinking bubble first
      const cleanMessages = updatedMessages.filter(m => !m.isThinking);

      if (result.success && result.data) {
        this.setData({
          messages: [...cleanMessages, { 
            text: result.data.content, 
            isUser: false 
          }],
          toView: `msg-${cleanMessages.length}`
        });
      } else {
        this.setData({
          messages: [...cleanMessages, { 
            text: 'Sorry, I encountered an error. Please try again.', 
            isUser: false,
            isError: true
          }]
        });
      }

    } catch (error) {
      console.error('AI Chat Error:', error);
      const updatedMessages = this.data.messages.filter(m => !m.isThinking);
      this.setData({
        messages: [...updatedMessages, { 
          text: 'Network error. Please check your connection.', 
          isUser: false,
          isError: true
        }]
      });
    } finally {
      this.setData({ loading: false });
    }
  }
})
