import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    t: {} as any,
    list: [] as any[]
  },

  onShow() {
    const lang = getLanguage();
    const t = getTranslations(lang);
    this.setData({ t });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Knowledge Base' : '知识库'
    });
    this.initData(lang);
  },

  initData(lang: 'en' | 'zh') {
    const list = lang === 'en' ? [
      {
        title: 'Five Rams Statue',
        tag: 'Outdoor',
        category: 'Attraction',
        time: '1-2 Hours',
        tags: ['History', 'Landmark'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/c/c2/Five_Rams_Statue.jpg/800px-Five_Rams_Statue.jpg'
      },
      {
        title: 'Shrimp Dumpling',
        tag: 'Food',
        category: 'Dim Sum',
        time: '< 1 Hour',
        tags: ['Classic', 'Seafood'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Har_gow.jpg/1200px-Har_gow.jpg'
      },
      {
        title: 'Cantonese Opera',
        tag: 'Culture',
        category: 'Art',
        time: '2-3 Hours',
        tags: ['Music', 'Performance'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/e/ec/Map_of_Guangzhou.jpg/1200px-Map_of_Guangzhou.jpg'
      }
    ] : [
      {
        title: '五羊石像',
        tag: '户外',
        category: '景点',
        time: '1-2 小时',
        tags: ['历史', '地标'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/c/c2/Five_Rams_Statue.jpg/800px-Five_Rams_Statue.jpg'
      },
      {
        title: '虾饺',
        tag: '美食',
        category: '点心',
        time: '< 1 小时',
        tags: ['经典', '海鲜'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Har_gow.jpg/1200px-Har_gow.jpg'
      },
      {
        title: '粤剧',
        tag: '文化',
        category: '艺术',
        time: '2-3 小时',
        tags: ['音乐', '表演'],
        image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/e/ec/Map_of_Guangzhou.jpg/1200px-Map_of_Guangzhou.jpg'
      }
    ];
    this.setData({ list });
  }
})