import { getLanguage, getTranslations } from '../../../utils/i18n';

Page({
  mapContext: null as any,
  pendingFitPoints: [] as any[],
  data: {
    t: {} as any,
    routeData: null as any,
    center: {
      latitude: 23.1291,
      longitude: 113.2644
    },
    mapScale: 14,
    markers: [] as any[],
    polyline: [] as any[],
    sheetHeight: 40, // Initial height in vh
    isDragging: false
  },

  // Drag state variables
  dragState: {
    startY: 0,
    startHeight: 0
  },

  onLoad(options: any) {
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Route Plan' : '路线详情'
    });

    if (options.routeId) {
      this.loadRoute(options.routeId);
    }
  },
  onReady() {
    this.mapContext = wx.createMapContext('routeMap', this);
    if (this.pendingFitPoints.length > 0) {
      this.fitMapToPoints(this.pendingFitPoints);
    }
  },

  async loadRoute(routeId: string) {
    wx.showLoading({ title: 'Loading...' });
    try {
      const db = wx.cloud.database();
      const res = await db.collection('routes').doc(routeId).get();
      const route = res.data;

      console.log('Loaded route data:', route);

      if (!route) {
        throw new Error('Route not found');
      }

      this.setData({ routeData: route });
      this.renderMap(route);
    } catch (error) {
      console.error('Failed to load route:', error);
      wx.showToast({
        title: 'Failed to load route',
        icon: 'none'
      });
      
      // Optional: Navigate back after delay
      setTimeout(() => {
        wx.navigateBack();
      }, 2000);
    } finally {
      wx.hideLoading();
    }
  },

  renderMap(route: any) {
    if (!route.waypoints || route.waypoints.length === 0) return;

    const totalPoints = route.waypoints.length;
    const markers = route.waypoints.map((wp: any, index: number) => {
      const isStart = index === 0;
      const isEnd = index === totalPoints - 1;
      const size = isStart || isEnd ? 40 : 32;
      const display = totalPoints > 6 && !isStart && !isEnd ? 'BYCLICK' : 'ALWAYS';
      return {
        id: index,
        latitude: wp.latitude,
        longitude: wp.longitude,
        title: wp.name,
        width: size,
        height: size,
        zIndex: isStart || isEnd ? 100 : 80,
        callout: {
          content: wp.name,
          padding: 8,
          borderRadius: 4,
          display
        }
      };
    });

    let points: any[] = [];
    if (route.polyline && Array.isArray(route.polyline)) {
      points = route.polyline.map((p: string) => {
        const [lat, lng] = p.split(',').map(Number);
        return { latitude: lat, longitude: lng };
      });
    } else {
      // Fallback: connect waypoints
      points = route.waypoints.map((wp: any) => ({
        latitude: wp.latitude,
        longitude: wp.longitude
      }));
    }

    if (points.length > 800) {
      const step = Math.ceil(points.length / 800);
      points = points.filter((_, index) => index % step === 0);
    }

    const polyline = [{
      points,
      color: '#3CAA3C',
      width: 6,
      arrowLine: true
    }];

    const center = {
      latitude: route.waypoints[0].latitude,
      longitude: route.waypoints[0].longitude
    };

    this.setData({
      markers,
      polyline,
      center
    });
    const fitPoints = this.reduceNearbyPoints(
      route.waypoints.map((wp: any) => ({
        latitude: wp.latitude,
        longitude: wp.longitude
      }))
    );
    this.fitMapToPoints(fitPoints);
  },

  reduceNearbyPoints(points: any[]) {
    if (!points || points.length <= 1) return points;
    const reduced = [points[0]];
    for (let i = 1; i < points.length; i += 1) {
      const prev = reduced[reduced.length - 1];
      const curr = points[i];
      const latDiff = Math.abs(prev.latitude - curr.latitude);
      const lngDiff = Math.abs(prev.longitude - curr.longitude);
      if (latDiff > 0.0003 || lngDiff > 0.0003) {
        reduced.push(curr);
      }
    }
    if (reduced[reduced.length - 1] !== points[points.length - 1]) {
      reduced.push(points[points.length - 1]);
    }
    return reduced;
  },

  fitMapToPoints(points: any[]) {
    if (!points || points.length === 0) return;
    if (!this.mapContext) {
      this.pendingFitPoints = points;
      return;
    }
    if (points.length <= 3) {
      this.setData({ mapScale: 16 });
      this.mapContext.moveToLocation({
        latitude: points[0].latitude,
        longitude: points[0].longitude
      });
      return;
    }
    const fitPoints = points.length > 10 ? points.filter((_, index) => index % Math.ceil(points.length / 10) === 0) : points;
    this.setData({ mapScale: 13 });
    this.mapContext.includePoints({
      points: fitPoints,
      padding: [100, 60, 200, 60]
    });
  },

  onBack() {
    wx.navigateBack();
  },

  onDragStart(e: any) {
    this.dragState.startY = e.touches[0].clientY;
    this.dragState.startHeight = this.data.sheetHeight;
    this.setData({ isDragging: true });
  },

  onDragMove(e: any) {
    const deltaY = e.touches[0].clientY - this.dragState.startY;
    const { windowHeight } = wx.getSystemInfoSync();
    // Convert pixel delta to vh
    // Dragging down (positive deltaY) decreases height
    const deltaVh = (deltaY / windowHeight) * 100;
    
    let newHeight = this.dragState.startHeight - deltaVh;
    
    // Clamp height between 15vh and 90vh
    if (newHeight < 15) newHeight = 15;
    if (newHeight > 90) newHeight = 90;

    this.setData({ sheetHeight: newHeight });
  },

  onDragEnd() {
    this.setData({ isDragging: false });
    
    // Snap to nearest position
    const { sheetHeight } = this.data;
    let finalHeight = 40;
    
    if (sheetHeight < 25) {
      finalHeight = 15; // Collapsed
    } else if (sheetHeight > 65) {
      finalHeight = 90; // Expanded
    } else {
      finalHeight = 40; // Half
    }
    
    this.setData({ sheetHeight: finalHeight });
  }
})
