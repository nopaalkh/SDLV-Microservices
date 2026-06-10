import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom';
import { decodeToken, default as api } from './services/api';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import AssetDetail from './pages/AssetDetail';
import BuyerDashboard from './pages/BuyerDashboard';
import CreatorDashboard from './pages/CreatorDashboard';
import AdminDashboard from './pages/AdminDashboard';
import Profile from './pages/Profile';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import CreatorGallery from './pages/CreatorGallery';
import './index.css';

const Navbar = () => {
  const navigate = useNavigate();
  const token = localStorage.getItem('token');
  
  const location = useLocation();

  // Derive role from JWT claims, not from a separate localStorage key
  let role = localStorage.getItem('role') || '';
  if (token && !role) {
    const claims = decodeToken(token);
    role = (claims?.role || 'buyer').toLowerCase();
    localStorage.setItem('role', role);
  }

  // Periodic auth check to handle "instant ban"
  useEffect(() => {
    if (!token) return;
    const checkAuth = async () => {
      try {
        const res = await api.get('/auth/me');
        if (res.data?.status === 'SUSPENDED') {
          handleLogout('Akun Anda telah ditangguhkan oleh Admin.');
        }
      } catch (e) {
        if (e.response && (e.response.status === 401 || e.response.status === 403)) {
          handleLogout('Sesi Anda telah berakhir atau dibatasi.');
        }
      }
    };
    checkAuth();
  }, [token, location.pathname]);

  const handleLogout = (msg) => {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    if (typeof msg === 'string') alert(msg);
    navigate('/login');
    window.location.reload();
  };

  return (
    <nav className="navbar">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', maxWidth: '1200px', margin: '0 auto', width: '100%' }}>
        <Link to="/" className="text-gradient-primary" style={{ fontSize: '1.5rem', fontWeight: '800' }}>
          Azetqu
        </Link>
        <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
          <Link to="/" className="nav-link">Jelajahi</Link>
          
          {!token ? (
            <>
              <Link to="/login" className="nav-link">Masuk</Link>
              <Link to="/register" className="btn btn-primary" style={{ padding: '0.5rem 1.5rem' }}>Daftar</Link>
            </>
          ) : (
            <>
              {role === 'admin' && <Link to="/dashboard/admin" className="nav-link">Panel Admin</Link>}
              {role === 'creator' && <Link to="/dashboard/creator" className="nav-link">Studio Kreator</Link>}
              <Link to="/dashboard/buyer" className="nav-link">Pustaka Saya</Link>
              <Link to="/profile" className="nav-link">Profil</Link>
              <button onClick={handleLogout} className="btn btn-danger" style={{ padding: '0.5rem 1.5rem' }}>Keluar</button>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

function App() {
  return (
    <Router>
      <Navbar />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/assets/:id" element={<AssetDetail />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/dashboard/buyer" element={<BuyerDashboard />} />
        <Route path="/dashboard/creator" element={<CreatorDashboard />} />
        <Route path="/creator/assets" element={<CreatorGallery />} />
        <Route path="/dashboard/admin" element={<AdminDashboard />} />
        <Route path="/profile" element={<Profile />} />
      </Routes>
    </Router>
  );
}

export default App;
