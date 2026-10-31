import { useEffect, useRef, useState } from 'react';
import { Alert, Spin } from 'antd';
import AMapLoader from '@amap/amap-jsapi-loader';
import type { Coordinate, Waypoint } from '../api/client';

const AMAP_KEY = import.meta.env.VITE_AMAP_JS_KEY || '';
const AMAP_SECURITY_CODE = import.meta.env.VITE_AMAP_JS_SECURITY_CODE || '';

declare global {
  interface Window {
    _AMapSecurityConfig?: { securityJsCode?: string };
  }
}

let amapPromise: Promise<any> | null = null;

/** Loads the AMap JS API once and reuses the promise across mounts. */
function loadAMap(): Promise<any> {
  if (!amapPromise) {
    if (AMAP_SECURITY_CODE) {
      window._AMapSecurityConfig = { securityJsCode: AMAP_SECURITY_CODE };
    }
    amapPromise = AMapLoader.load({
      key: AMAP_KEY,
      version: '2.0',
      plugins: ['AMap.Scale'],
    });
  }
  return amapPromise;
}

interface Props {
  polyline: Coordinate[];
  waypoints: Waypoint[];
  language: string;
}

/**
 * AMap (Gaode) route map. All coordinates are GCJ-02, which is exactly what
 * the seeded places and the AMap driving API return, so no conversion is needed.
 */
export default function AmapRouteMap({ polyline, waypoints, language }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const amapRef = useRef<any>(null);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!AMAP_KEY) {
      setError('missing-key');
      return;
    }
    let disposed = false;
    let observer: ResizeObserver | null = null;

    loadAMap()
      .then((AMap) => {
        if (disposed || !containerRef.current) return;
        amapRef.current = AMap;
        const map = new AMap.Map(containerRef.current, {
          zoom: 12,
          center: [113.2644, 23.1291],
          viewMode: '2D',
          resizeEnable: true,
        });
        map.addControl?.(new AMap.Scale());
        mapRef.current = map;
        setReady(true);

        observer = new ResizeObserver(() => map.resize());
        observer.observe(containerRef.current);
      })
      .catch((e: Error) => {
        if (!disposed) setError(e?.message || 'load-failed');
      });

    return () => {
      disposed = true;
      observer?.disconnect();
      mapRef.current?.destroy?.();
      mapRef.current = null;
      amapRef.current = null;
      setReady(false);
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const AMap = amapRef.current;
    if (!map || !AMap || !ready) return;

    map.clearMap();
    const overlays: any[] = [];

    const path = polyline.map((p) => [p.lng, p.lat]);
    if (path.length >= 2) {
      overlays.push(
        new AMap.Polyline({
          path,
          strokeColor: '#0052d9',
          strokeWeight: 5,
          strokeOpacity: 0.9,
          lineJoin: 'round',
          lineCap: 'round',
          showDir: true,
          zIndex: 50,
        }),
      );
    }

    waypoints.forEach((wp, i) => {
      overlays.push(
        new AMap.Marker({
          position: [wp.lng, wp.lat],
          content: `<div class="amap-marker">${i + 1}</div>`,
          offset: new AMap.Pixel(-13, -13),
          title: language === 'zh' ? wp.name : wp.nameEn || wp.name,
          zIndex: 100 + i,
        }),
      );
    });

    if (overlays.length > 0) {
      map.add(overlays);
      map.setFitView(overlays, false, [40, 40, 40, 40], 15);
    }
  }, [polyline, waypoints, language, ready]);

  if (error === 'missing-key') {
    return (
      <div className="route-map route-map-placeholder">
        <Alert
          type="warning"
          showIcon
          message="地图未配置 / Map key missing"
          description="请在 web/.env.local 设置 VITE_AMAP_JS_KEY 与 VITE_AMAP_JS_SECURITY_CODE 后重新构建。"
        />
      </div>
    );
  }

  if (error) {
    return (
      <div className="route-map route-map-placeholder">
        <Alert
          type="error"
          showIcon
          message="地图加载失败 / Map failed to load"
          description={error}
        />
      </div>
    );
  }

  return (
    <div className="route-map-wrap">
      <div ref={containerRef} className="route-map" />
      {!ready && (
        <div className="route-map-loading">
          <Spin />
        </div>
      )}
    </div>
  );
}
