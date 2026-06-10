import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import api, { decodeToken } from '../services/api';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      const res = await api.post('/auth/login', { email, password });
      
      if (res.data && res.data.token) {
        const token = res.data.token;
        localStorage.setItem('token', token);
        
        try {
          const meRes = await api.get('/auth/me', { headers: { Authorization: `Bearer ${token}` } });
          const role = meRes.data?.role || 'BUYER';
          localStorage.setItem('role', role.toLowerCase());
        } catch (e) {
          console.error('Failed to get user profile', e);
          localStorage.setItem('role', 'buyer');
        }
        
        navigate('/');
        window.location.reload(); 
      } else {
        throw new Error("Format respons tidak valid");
      }
    } catch (err) {
      console.error('Login gagal', err);
      setError(err.response?.data?.error || err.response?.data?.message || 'Kredensial tidak valid. Silakan coba lagi.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="main-content" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '80vh' }}>
      <div className="glass-panel animate-fade-in" style={{ width: '100%', maxWidth: '450px' }}>
        <h1 style={{ textAlign: 'center', marginBottom: '0.5rem' }}>Selamat Datang Kembali</h1>
        <p style={{ textAlign: 'center', color: 'var(--text-muted)', marginBottom: '2rem' }}>Masuk untuk melanjutkan ke Azetqu</p>
        
        {error && (
          <div style={{ background: 'var(--surface)', border: 'var(--border-width) solid var(--danger)', boxShadow: '4px 4px 0px var(--danger)', color: 'var(--text)', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem', textAlign: 'center', fontSize: '0.875rem', fontWeight: 'bold' }}>
            {error}
          </div>
        )}
        
        <form onSubmit={handleLogin}>
          <div className="input-group">
            <label>Alamat Email</label>
            <input 
              type="email" 
              className="input-field" 
              placeholder="nama@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          
          <div className="input-group">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <label style={{ margin: 0 }}>Kata Sandi</label>
              <Link to="/forgot-password" style={{ fontSize: '0.875rem', color: 'var(--primary)', fontWeight: '500' }}>Lupa Kata Sandi?</Link>
            </div>
            <input 
              type="password" 
              className="input-field" 
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          
          <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: '1rem', padding: '1rem' }} disabled={loading}>
            {loading ? 'Memproses...' : 'Masuk'}
          </button>
        </form>
        
        <p style={{ textAlign: 'center', marginTop: '2rem', color: 'var(--text-muted)' }}>
          Belum memiliki akun? <Link to="/register" style={{ color: 'var(--primary)', fontWeight: '600' }}>Daftar sekarang</Link>
        </p>
      </div>
    </div>
  );
};

export default Login;