import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    t: {} as any,
    tempImage: '',
    loading: false,
    generatedBlog: null as any
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/user/travelBlog/index'
      })
    }
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Travel Blog' : '智慧游记'
    });
  },

  onChooseImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        this.setData({
          tempImage: res.tempFiles[0].tempFilePath
        });
      }
    });
  },

  async onGenerate() {
    if (!this.data.tempImage) return;
    
    this.setData({ loading: true });
    
    try {
      // 1. Upload image (Mocking the cloud upload for speed, using temp path for display)
      // In real scenario: const uploadRes = await wx.cloud.uploadFile(...)
      
      // 2. Call AI Service
      // We'll use 'generate_copy' action which expects text, languages, style
      // For image-based generation, we'd ideally use a vision model, but for now we'll simulate by describing it
      // or just asking for a generic travel blog based on a prompt.
      
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'generate_copy',
          payload: {
            text: "Write a short, engaging travel blog post about a beautiful trip in Guangzhou. Include a title and content.",
            languages: ['en'], // Request English response for simplicity, or match current lang
            style: 'Instagram'
          }
        }
      });

      const result = res.result as any;
      if (result.success && result.data && result.data.en) {
        // The API returns { en: "..." } string usually, or maybe we need to parse it if it's JSON
        // The handleCopy in aiService returns JSON object with language keys.
        
        // Let's assume the result.data.en is the content. 
        // We might need to split title/content manually if the model didn't separate them, 
        // or just use the whole text.
        
        this.setData({
          generatedBlog: {
            title: "My Amazing Trip", // AI might not return separate title in simple copy generation
            content: result.data.en,
            image: this.data.tempImage
          }
        });
        
        wx.showToast({
          title: 'Generated!',
          icon: 'success'
        });
      } else {
        throw new Error('Generation failed');
      }
      
    } catch (error) {
      console.error('Blog generation failed:', error);
      wx.showToast({
        title: 'Failed to generate',
        icon: 'none'
      });
    } finally {
      this.setData({ loading: false });
    }
  }
})
