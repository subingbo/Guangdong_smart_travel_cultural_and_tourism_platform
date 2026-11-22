import { getLanguage, getTranslations } from '../../../utils/i18n';
import { aiCopy, createPoster, getPosterTask } from '../../../services/api';
import { getToken } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';

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

  requireLogin(): boolean {
    if (getToken()) return true;
    wx.showModal({
      title: '需要登录',
      content: 'AI 文案与海报生成需要先登录，是否前往登录页？',
      confirmText: '去登录',
      success: (res) => {
        if (res.confirm) gotoLogin();
      }
    });
    return false;
  },

  async onGenerate() {
    if (!this.data.inputText) {
      wx.showToast({ title: this.data.t.generate.pleaseEnter, icon: 'none' })
      return
    }
    if (!this.requireLogin()) return

    this.setData({ loading: true, results: null })

    try {
      // Backend tone vocabulary: friendly | literary | promotional.
      const toneMap: Record<string, string> = {
        instagram: 'promotional',
        twitter: 'friendly',
        facebook: 'literary'
      };
      const results: Record<string, string> = {};
      for (const lang of this.data.languages) {
        const resp = await aiCopy(this.data.inputText, 'general', toneMap[this.data.tone] || 'friendly', lang);
        results[lang] = resp.content;
      }
      this.setData({ results });
    } catch (error) {
      console.error(error)
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
    if (!this.requireLogin()) return;

    this.setData({ posterLoading: true, posterUrl: '' });

    // Use custom prompt if available, otherwise construct one
    let posterPrompt = this.data.customPosterPrompt;
    if (!posterPrompt) {
        posterPrompt = `${this.data.inputText}, promotional poster, high quality`;
        // Update the custom prompt field so user sees what was used and can edit it
        this.setData({ customPosterPrompt: posterPrompt });
    }

    try {
      const created = await createPoster(posterPrompt, this.data.posterStyle, getLanguage());
      await this.pollPosterTask(created.taskId);
    } catch (error) {
      console.error(error);
    } finally {
      this.setData({ posterLoading: false });
    }
  },
  async pollPosterTask(taskId: string) {
    const maxAttempts = 12;
    const interval = 3000;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const task = await getPosterTask(taskId);
      if (task.status === 'succeeded' && task.imageUrl) {
        this.setData({ posterUrl: task.imageUrl });
        // Show the poster in the card; the user opens the preview explicitly.
        return;
      }
      if (task.status === 'failed') {
        wx.showToast({ title: task.message || 'Poster generation failed', icon: 'none' });
        return;
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
