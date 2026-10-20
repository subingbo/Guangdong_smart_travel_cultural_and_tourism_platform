import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    loading: false,
    inputText: '',
    languages: ['en'], // Default English
    tone: 'instagram',
    posterStyle: 'Default',
    results: null as any, // { en: '...', ja: '...' }
    posterLoading: false,
    posterUrl: '',
    customPosterPrompt: '', // User adjustable prompt
    t: {} as any
  },

  onShow() {
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'AI Copywriter' : 'AI 文案助手'
    });
  },

  onInputChange(e: any) {
    this.setData({ inputText: e.detail.value })
  },
  
  onPosterPromptChange(e: any) {
      this.setData({ customPosterPrompt: e.detail.value });
  },

  onLanguageChange(e: any) {
    const lang = e.currentTarget.dataset.value
    const languages = [...this.data.languages]
    const index = languages.indexOf(lang)
    if (index > -1) {
      if (languages.length > 1) {
        languages.splice(index, 1)
      } else {
        wx.showToast({ title: this.data.t.generate.atLeastOne, icon: 'none' })
        return
      }
    } else {
      languages.push(lang)
    }
    this.setData({ languages })
  },

  onToneChange(e: any) {
    this.setData({ tone: e.detail.value })
  },

  onPosterStyleChange(e: any) {
    this.setData({ posterStyle: e.detail.value })
  },

  async onGenerate() {
    if (!this.data.inputText) {
      wx.showToast({ title: this.data.t.generate.pleaseEnter, icon: 'none' })
      return
    }

    this.setData({ loading: true, results: null })

    try {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'generate_copy',
          payload: {
            text: this.data.inputText,
            languages: this.data.languages,
            style: this.data.tone
          }
        }
      })
      
      const result = res.result as any;
      if (result.success && result.data) {
        this.setData({ results: result.data });
      } else {
        wx.showToast({ title: this.data.t.generate.generationFailed, icon: 'none' });
        console.error('Copy Error:', result);
      }
      
    } catch (error) {
      console.error(error)
      wx.showToast({ title: this.data.t.generate.generationFailed, icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async onGeneratePoster() {
    // Allow generation if there is either inputText OR a customPosterPrompt
    if (!this.data.inputText && !this.data.customPosterPrompt) {
       wx.showToast({ title: this.data.t.generate.enterTextFirst, icon: 'none' });
       return;
    }
    
    this.setData({ posterLoading: true, posterUrl: '' });
    
    // Use custom prompt if available, otherwise construct one
    let posterPrompt = this.data.customPosterPrompt;
    if (!posterPrompt) {
        posterPrompt = `${this.data.inputText}, ${this.data.tone} style, high quality, promotional poster`;
        // Update the custom prompt field so user sees what was used and can edit it
        this.setData({ customPosterPrompt: posterPrompt });
    }

    try {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'generate_poster',
          payload: {
            prompt: posterPrompt,
            style: this.data.posterStyle
          }
        }
      });

      const result = res.result as any;
      if (result.success && result.taskId) {
        await this.pollPosterTask(result.taskId, posterPrompt, this.data.posterStyle);
      } else {
        wx.showToast({ title: 'Poster generation failed', icon: 'none' });
        console.error('Poster Error:', result);
      }
    } catch (error) {
      console.error(error);
      wx.showToast({ title: 'Network error', icon: 'none' });
    } finally {
      this.setData({ posterLoading: false });
    }
  },
  async pollPosterTask(taskId: string, prompt: string, style: string) {
    const maxAttempts = 12;
    const interval = 3000;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'get_poster_task',
          payload: {
            taskId,
            prompt,
            style
          }
        }
      });
      const result = res.result as any;
      if (result.success && result.imageUrl) {
        this.setData({ posterUrl: result.imageUrl });
        // Don't auto-preview, just show in card as requested
        // wx.previewImage({ ... }); 
        return;
      }
      if (!result.success || !result.pending) {
        throw new Error(result.message || 'Poster generation failed');
      }
      await this.sleep(interval);
    }
    wx.showToast({ title: 'Poster generation timeout', icon: 'none' });
  },
  sleep(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
  },

  onPreviewPoster() {
    if (this.data.posterUrl) {
      wx.previewImage({
        urls: [this.data.posterUrl],
        current: this.data.posterUrl
      })
    }
  }
})
