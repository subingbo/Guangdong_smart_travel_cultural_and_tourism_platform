import { HashRouter, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import BottomNav from './components/BottomNav';
import HomePage from './pages/HomePage';
import RoutePage from './pages/RoutePage';
import BlogPage from './pages/BlogPage';
import GuidePage from './pages/GuidePage';
import ProfilePage from './pages/ProfilePage';

export default function App() {
  return (
    <AuthProvider>
      <HashRouter>
        <div className="page-container">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/route" element={<RoutePage />} />
            <Route path="/blog" element={<BlogPage />} />
            <Route path="/guide" element={<GuidePage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Routes>
        </div>
        <BottomNav />
      </HashRouter>
    </AuthProvider>
  );
}
