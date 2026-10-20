import { getLanguage, getTranslations } from '../utils/i18n';

Component({
  data: {
    selected: 'pages/index/index',
    t: {} as any
  },
  methods: {
    onChange(e: any) {
      const url = e.detail.value;
      wx.switchTab({
        url: `/${url}`,
      });
    },
    init() {
      const page = getCurrentPages().pop();
      const route = page ? page.route : 'pages/index/index';
      const lang = getLanguage();
      this.setData({
        selected: route,
        t: getTranslations(lang)
      });
    }
  },
  lifetimes: {
    attached() {
      const lang = getLanguage();
      this.setData({
        t: getTranslations(lang)
      });
    }
  },
  pageLifetimes: {
    show() {
      const lang = getLanguage();
      this.setData({
        t: getTranslations(lang)
      });
    }
  }
});
