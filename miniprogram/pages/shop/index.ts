import { getLanguage, getTranslations } from '../../utils/i18n';

type Banner = { id: string; url: string };
type Category = { key: string; text: string; icon: string };
type Product = { id: string; title: string; price: number; tag: string; image: string };

Page({
  data: {
    t: {} as any,
    keyword: '',
    banners: [] as Banner[],
    categories: [] as Category[],
    hotProducts: [] as Product[],
    cartCount: 0,
    cartTotal: 0
  },

  onShow() {
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Shop' : '线上商城'
    });

    const banners: Banner[] = [
      {
        id: 'b1',
        url: 'https://www.jpchinapress.com/image/2025-11-14/thumb/1438943797337686016.jpg'
      },
      {
        id: 'b2',
        url: 'https://g-search3.alicdn.com/img/bao/uploaded/i4/i3/2212567806198/O1CN01qip8nk1vejlhm3DzU_!!0-item_pic.jpg_360x360q90.jpg_.webp'
      }
    ];

    const categories: Category[] =
      lang === 'en'
        ? [
            { key: 'tickets', text: 'Tickets', icon: 'map' },
            { key: 'food', text: 'Snacks', icon: 'fork' },
            { key: 'souvenir', text: 'Souvenir', icon: 'image' },
            { key: 'coupon', text: 'Coupons', icon: 'discount' }
          ]
        : [
            { key: 'tickets', text: '门票', icon: 'map' },
            { key: 'food', text: '特产', icon: 'fork' },
            { key: 'souvenir', text: '文创', icon: 'image' },
            { key: 'coupon', text: '优惠券', icon: 'discount' }
          ];

    const hotProducts: Product[] =
      lang === 'en'
        ? [
            {
              id: 'p1',
              title: 'Lingnan Souvenir Set',
              price: 99,
              tag: 'Souvenir',
              image: 'https://g-search3.alicdn.com/img/bao/uploaded/i4/i1/3038723547/O1CN01AqShxl1c4Zs7xbmyt_!!4611686018427387355-0-item_pic.jpg_360x360q90.jpg_.webp'
            },
            {
              id: 'p2',
              title: 'Lingnan Blind Box',
              price: 59,
              tag: 'Blind Box',
              image: 'https://www.jpchinapress.com/image/2025-11-14/thumb/1438943797337686016.jpg'
            }
          ]
        : [
            {
              id: 'p1',
              title: '岭南文创套装',
              price: 99,
              tag: '文创',
              image: 'https://g-search3.alicdn.com/img/bao/uploaded/i4/i1/3038723547/O1CN01AqShxl1c4Zs7xbmyt_!!4611686018427387355-0-item_pic.jpg_360x360q90.jpg_.webp'
            },
            {
              id: 'p2',
              title: '岭南盲盒（城市款）',
              price: 59,
              tag: '盲盒',
              image: 'https://www.jpchinapress.com/image/2025-11-14/thumb/1438943797337686016.jpg'
            }
          ];

    this.setData({
      banners,
      categories,
      hotProducts
    });
  },

  onKeywordChange(e: any) {
    this.setData({ keyword: e.detail.value });
  },

  onTapSearch() {
    const lang = getLanguage();
    wx.showToast({
      title: lang === 'en' ? 'Search coming soon' : '搜索功能开发中',
      icon: 'none'
    });
  },

  onTapCart() {
    const lang = getLanguage();
    wx.showToast({
      title: lang === 'en' ? 'Cart coming soon' : '购物车功能开发中',
      icon: 'none'
    });
  },

  onTapCategory(e: any) {
    const { index } = e.detail || {};
    const item = this.data.categories?.[index];
    if (!item) return;
    wx.showToast({ title: item.text, icon: 'none' });
  },

  onAddToCart(e: any) {
    const { id } = e.currentTarget.dataset || {};
    const product = this.data.hotProducts.find((p) => p.id === id);
    if (!product) return;

    const nextCount = this.data.cartCount + 1;
    const nextTotal = Number((this.data.cartTotal + product.price).toFixed(2));
    this.setData({ cartCount: nextCount, cartTotal: nextTotal });
  },

  onCheckout() {
    const lang = getLanguage();
    wx.showToast({
      title: lang === 'en' ? 'Checkout coming soon' : '结算功能开发中',
      icon: 'none'
    });
  }
});
