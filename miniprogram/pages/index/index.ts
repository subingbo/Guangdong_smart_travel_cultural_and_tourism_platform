import { getLanguage, getTranslations } from '../../utils/i18n';

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
    // Removed fetchSceneryImages from onLoad to ensure fresh random images on every show
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
    const translations = getTranslations(lang);
    // Initialize 't' with default translations if not set to avoid null prop errors
    if (!this.data.t || Object.keys(this.data.t).length === 0) {
      this.setData({ t: translations });
    } else {
      this.setData({ t: translations });
    }
    this.fetchSceneryImages();
  },

  async fetchSceneryImages() {
    try {
      if (!wx.cloud) {
        console.error('wx.cloud not available');
        return;
      }
      
      const res = await wx.cloud.callFunction({
        name: 'getSceneryImages',
      });
      
      console.log('Fetched images:', res);
      
      if (res.result && (res.result as any).images) {
        this.setData({
          swiperList: (res.result as any).images
        });
      } else {
        console.error('Cloud function returned error or no images:', res.result);
        // Fallback to default images if cloud function returns error
        throw new Error('Cloud function error');
      }
    } catch (err) {
      console.error('Failed to fetch scenery images', err);
      // Fallback images if cloud function fails (e.g. offline dev)
      this.setData({
        swiperList: [
          { url: 'https://images.unsplash.com/photo-1599571234909-29ed5d1321d6', name: 'Guangzhou Tower' },
          { url: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e', name: 'Beach Sunset' }
        ]
      });
    }
  },

  onSwiperTap(e: any) {
    console.log('Swiper tap', e);
  },

  onRecommendTap() {
    console.log('Recommend tap');
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
