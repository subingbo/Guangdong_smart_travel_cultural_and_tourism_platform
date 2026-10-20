import { getLanguage, getTranslations } from '../../utils/i18n';

Page({
  data: {
    t: {} as any
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/merchant/index'
      })
    }
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Merchant Center' : '商家中心'
    });
  }
})
