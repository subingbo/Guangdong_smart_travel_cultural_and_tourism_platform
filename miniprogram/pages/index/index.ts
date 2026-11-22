import { getLanguage, getTranslations } from '../../utils/i18n';
import { listScenics, Scenic } from '../../services/api';

Page({
  data: {
    current: 0,
    autoplay: true,
    duration: 500,
    interval: 5000,
    swiperList: [] as { url: string; name: string }[],
    t: {} as any
  },

  onLoad() {
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/index/index'
      })
    }
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    this.fetchSceneryImages();
  },

  async fetchSceneryImages() {
    try {
      const scenics = await listScenics();
      const swiperList = scenics.map((s: Scenic) => ({
        url: s.imageUrl,
        name: getLanguage() === 'zh' ? s.titleCn : s.titleEn,
      }));
      this.setData({ swiperList });
    } catch (err) {
      console.error('Failed to fetch scenery images', err);
      // Fallback images keep the home page alive when the backend is offline.
      this.setData({
        swiperList: [
          { url: 'https://images.unsplash.com/photo-1599571234909-29ed5d1321d6', name: 'Guangzhou Tower' },
          { url: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e', name: 'Beach Sunset' }
        ]
      });
    }
  },

  onTapExploreItem(e: any) {
    const { index } = e.detail || {};
    const items = [
      { url: '/pages/user/routeInput/index', isTab: true },
      { url: '/pages/merchant/menuScan/index', isTab: false },
      { url: '/pages/user/travelBlog/index', isTab: true },
      { url: '/pages/user/guide/index', isTab: true },
      { url: '/pages/shop/index', isTab: false }
    ];
    const target = items[index];
    if (!target) return;

    if (target.isTab) {
      wx.switchTab({ url: target.url });
      return;
    }
    wx.navigateTo({ url: target.url });
  },

  onTapExploreDirect(e: any) {
    const { url, istab } = e.currentTarget.dataset || {};
    if (!url) return;
    const isTab = String(istab) === 'true';
    if (isTab) {
      wx.switchTab({ url });
    } else {
      wx.navigateTo({ url });
    }
  }
});
