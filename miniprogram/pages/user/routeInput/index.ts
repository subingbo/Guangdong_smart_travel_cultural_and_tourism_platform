import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  data: {
    t: {} as any,
    loading: false,
    previewCenter: {
      latitude: 23.1291,
      longitude: 113.2644
    },
    previewScale: 13,
    previewMarkers: [] as any[],
    previewPolyline: [] as any[],
    durations: [
      { label: '1 Day', value: 1, selected: true },
      { label: '2 Days', value: 2, selected: false },
      { label: '3 Days', value: 3, selected: false },
      { label: '5+ Days', value: 5, selected: false }
    ],
    interestTags: [
      { label: 'Foodie', value: 'foodie', icon: 'fork', selected: false },
      { label: 'History', value: 'history', icon: 'temple', selected: false },
      { label: 'Nature', value: 'nature', icon: 'tree', selected: false },
      { label: 'Art', value: 'art', icon: 'palette', selected: false },
      { label: 'Shopping', value: 'shopping', icon: 'shop', selected: false }
    ],
    paces: [
      { label: 'Relaxed', value: 'relaxed', selected: false },
      { label: 'Balanced', value: 'balanced', selected: true },
      { label: 'Packed', value: 'packed', selected: false }
    ],
    travelModes: [
      { label: 'Driving', value: 'driving', icon: 'car', selected: true },
      { label: 'Walking', value: 'walking', icon: 'user', selected: false },
      { label: 'Bicycling', value: 'bicycling', icon: 'bicycle', selected: false },
      { label: 'Transit', value: 'transit', icon: 'bus', selected: false }
    ]
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/user/routeInput/index'
      })
    }
    const lang = getLanguage();
    const t = getTranslations(lang);
    this.setData({ t });
    
    // Update labels based on language
    if (lang === 'zh') {
      this.setData({
        durations: [
          { label: '1 天', value: 1, selected: this.data.durations[0].selected },
          { label: '2 天', value: 2, selected: this.data.durations[1].selected },
          { label: '3 天', value: 3, selected: this.data.durations[2].selected },
          { label: '5+ 天', value: 5, selected: this.data.durations[3].selected }
        ],
        interestTags: [
          { label: '美食', value: 'foodie', icon: 'fork', selected: this.data.interestTags[0].selected },
          { label: '历史', value: 'history', icon: 'temple', selected: this.data.interestTags[1].selected },
          { label: '自然', value: 'nature', icon: 'tree', selected: this.data.interestTags[2].selected },
          { label: '艺术', value: 'art', icon: 'palette', selected: this.data.interestTags[3].selected },
          { label: '购物', value: 'shopping', icon: 'shop', selected: this.data.interestTags[4].selected }
        ],
        paces: [
          { label: '休闲', value: 'relaxed', selected: this.data.paces[0].selected },
          { label: '适中', value: 'balanced', selected: this.data.paces[1].selected },
          { label: '紧凑', value: 'packed', selected: this.data.paces[2].selected }
        ],
        travelModes: [
          { label: '驾车', value: 'driving', icon: 'car', selected: this.data.travelModes[0].selected },
          { label: '步行', value: 'walking', icon: 'user', selected: this.data.travelModes[1].selected },
          { label: '骑行', value: 'bicycling', icon: 'bicycle', selected: this.data.travelModes[2].selected },
          { label: '公交', value: 'transit', icon: 'bus', selected: this.data.travelModes[3].selected }
        ]
      });
    } else {
      // Restore English labels if needed
      this.setData({
        durations: [
          { label: '1 Day', value: 1, selected: this.data.durations[0].selected },
          { label: '2 Days', value: 2, selected: this.data.durations[1].selected },
          { label: '3 Days', value: 3, selected: this.data.durations[2].selected },
          { label: '5+ Days', value: 5, selected: this.data.durations[3].selected }
        ],
        interestTags: [
          { label: 'Foodie', value: 'foodie', icon: 'fork', selected: this.data.interestTags[0].selected },
          { label: 'History', value: 'history', icon: 'temple', selected: this.data.interestTags[1].selected },
          { label: 'Nature', value: 'nature', icon: 'tree', selected: this.data.interestTags[2].selected },
          { label: 'Art', value: 'art', icon: 'palette', selected: this.data.interestTags[3].selected },
          { label: 'Shopping', value: 'shopping', icon: 'shop', selected: this.data.interestTags[4].selected }
        ],
        paces: [
          { label: 'Relaxed', value: 'relaxed', selected: this.data.paces[0].selected },
          { label: 'Balanced', value: 'balanced', selected: this.data.paces[1].selected },
          { label: 'Packed', value: 'packed', selected: this.data.paces[2].selected }
        ],
        travelModes: [
          { label: 'Driving', value: 'driving', icon: 'car', selected: this.data.travelModes[0].selected },
          { label: 'Walking', value: 'walking', icon: 'user', selected: this.data.travelModes[1].selected },
          { label: 'Bicycling', value: 'bicycling', icon: 'bicycle', selected: this.data.travelModes[2].selected },
          { label: 'Transit', value: 'transit', icon: 'bus', selected: this.data.travelModes[3].selected }
        ]
      });
    }
    
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Route Planner' : '路线规划'
    });
    this.initPreviewMap();
  },
  initPreviewMap() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        const center = { latitude: res.latitude, longitude: res.longitude };
        this.updatePreviewRoute(center);
      },
      fail: () => {
        this.updatePreviewRoute(this.data.previewCenter);
      }
    });
  },
  updatePreviewRoute(center: { latitude: number; longitude: number }) {
    const points = [
      { latitude: center.latitude, longitude: center.longitude },
      { latitude: center.latitude + 0.015, longitude: center.longitude + 0.01 },
      { latitude: center.latitude + 0.03, longitude: center.longitude + 0.02 }
    ];
    const markers = [
      { id: 0, latitude: points[0].latitude, longitude: points[0].longitude, width: 36, height: 36 },
      { id: 1, latitude: points[1].latitude, longitude: points[1].longitude, width: 32, height: 32 },
      { id: 2, latitude: points[2].latitude, longitude: points[2].longitude, width: 36, height: 36 }
    ];
    const polyline = [{
      points,
      color: '#3CAA3C',
      width: 6,
      arrowLine: true
    }];
    this.setData({
      previewCenter: center,
      previewMarkers: markers,
      previewPolyline: polyline
    });
  },

  onDurationChange(e: any) {
    const index = e.currentTarget.dataset.index;
    if (index === undefined) return;
    
    const durations = this.data.durations.map((item, i) => ({
      ...item,
      selected: i === index
    }));
    this.setData({ durations });
  },

  onInterestChange(e: any) {
    const index = e.currentTarget.dataset.index;
    if (index === undefined) return;

    const interestTags = this.data.interestTags.map((item, i) => {
      if (i === index) {
        return { ...item, selected: !item.selected };
      }
      return item;
    });
    this.setData({ interestTags });
  },

  onPaceChange(e: any) {
    const index = e.currentTarget.dataset.index;
    if (index === undefined) return;

    const paces = this.data.paces.map((item, i) => ({
      ...item,
      selected: i === index
    }));
    this.setData({ paces });
  },

  onTravelModeChange(e: any) {
    const index = e.currentTarget.dataset.index;
    if (index === undefined) return;

    const travelModes = this.data.travelModes.map((item, i) => ({
      ...item,
      selected: i === index
    }));
    this.setData({ travelModes });
  },

  async onGenerate() {
    const duration = this.data.durations.find(d => d.selected)?.value || 1;
    const interests = this.data.interestTags.filter(t => t.selected).map(t => t.value);
    const pace = this.data.paces.find(p => p.selected)?.value || 'balanced';
    const mode = this.data.travelModes.find(m => m.selected)?.value || 'driving';

    if (interests.length === 0) {
      wx.showToast({
        title: this.data.t.route?.select_interest || 'Please select at least one interest',
        icon: 'none'
      });
      return;
    }

    this.setData({ loading: true });

    try {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'generate_smart_route',
          payload: {
            duration,
            interests,
            pace,
            mode,
            location: 'Guangzhou' // Default for now
          }
        }
      });

      const result = res.result as any;
      if (result.success) {
        wx.navigateTo({
          url: `/pages/user/map/index?routeId=${result.routeId}`
        });
      } else {
        throw new Error(result.message || 'Generation failed');
      }
    } catch (error: any) {
      console.error('Route generation failed:', error);
      wx.showToast({
        title: 'Failed to generate route',
        icon: 'none'
      });
    } finally {
      this.setData({ loading: false });
    }
  }
})
