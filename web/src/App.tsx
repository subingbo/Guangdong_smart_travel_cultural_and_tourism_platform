import { HashRouter, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import AppHeader from './components/AppHeader';
import BottomNav from './components/BottomNav';
import ErrorBoundary from './components/ErrorBoundary';
import HomePage from './pages/HomePage';
import RoutePage from './pages/RoutePage';
import BlogPage from './pages/BlogPage';
import GuidePage from './pages/GuidePage';
import ProfilePage from './pages/ProfilePage';

export default function App() {
  return (
    <AuthProvider>
      <HashRouter>
        <AppHeader />
        <div className="page-container">
          <ErrorBoundary>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/route" element={<RoutePage />} />
              <Route path="/blog" element={<BlogPage />} />
              <Route path="/guide" element={<GuidePage />} />
              <Route path="/profile" element={<ProfilePage />} />
            </Routes>
          </ErrorBoundary>
        </div>
        <BottomNav />
      </HashRouter>
    </AuthProvider>
  );
}
