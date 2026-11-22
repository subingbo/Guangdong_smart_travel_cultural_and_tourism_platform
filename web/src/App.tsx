import { lazy, Suspense } from 'react';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { Spin } from 'antd';
import { AuthProvider } from './context/AuthContext';
import AppHeader from './components/AppHeader';
import AuthModal from './components/AuthModal';
import BottomNav from './components/BottomNav';
import ErrorBoundary from './components/ErrorBoundary';

// Route-level code splitting keeps the first paint small; shop/admin/merchant
// bundles only load when the visitor opens those pages.
const HomePage = lazy(() => import('./pages/HomePage'));
const RoutePage = lazy(() => import('./pages/RoutePage'));
const RouteDetailPage = lazy(() => import('./pages/RouteDetailPage'));
const BlogPage = lazy(() => import('./pages/BlogPage'));
const BlogDetailPage = lazy(() => import('./pages/BlogDetailPage'));
const ActivitiesPage = lazy(() => import('./pages/ActivitiesPage'));
const UserProfilePage = lazy(() => import('./pages/UserProfilePage'));
const NotificationsPage = lazy(() => import('./pages/NotificationsPage'));
const GuidePage = lazy(() => import('./pages/GuidePage'));
const MerchantPage = lazy(() => import('./pages/MerchantPage'));
const ShopPage = lazy(() => import('./pages/ShopPage'));
const OrdersPage = lazy(() => import('./pages/OrdersPage'));
const AdminPage = lazy(() => import('./pages/AdminPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage'));

function RouteFallback() {
  return (
    <div className="route-loading">
      <Spin />
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <HashRouter>
        <AppHeader />
        <div className="page-container">
          <ErrorBoundary>
            <Suspense fallback={<RouteFallback />}>
              <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/route" element={<RoutePage />} />
                <Route path="/routes/:id" element={<RouteDetailPage />} />
                <Route path="/blog" element={<BlogPage />} />
                <Route path="/blog/:id" element={<BlogDetailPage />} />
                <Route path="/activities" element={<ActivitiesPage />} />
                <Route path="/user/:id" element={<UserProfilePage />} />
                <Route path="/notifications" element={<NotificationsPage />} />
                <Route path="/guide" element={<GuidePage />} />
                <Route path="/merchant" element={<MerchantPage />} />
                <Route path="/shop" element={<ShopPage />} />
                <Route path="/orders" element={<OrdersPage />} />
                <Route path="/admin" element={<AdminPage />} />
                <Route path="/profile" element={<ProfilePage />} />
              </Routes>
            </Suspense>
          </ErrorBoundary>
        </div>
        <BottomNav />
      </HashRouter>
      <AuthModal />
    </AuthProvider>
  );
}
