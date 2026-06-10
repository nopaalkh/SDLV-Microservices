import { useState, useEffect } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import api from '../services/api';

const ResetPassword = () => {
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [status, setStatus] = useState({ type: '', message: '' });
  const [loading, setLoading] = useState(false);
  
  const navigate = useNavigate();
  const location = useLocation();
  
  // Extract token from URL
  const queryParams = new URLSearchParams(location.search);
  const token = queryParams.get('token');

  useEffect(() => {
    if (!token) {
      setStatus({ 
        type: 'error', 
        message: 'Token reset kata sandi tidak valid atau tidak ditemukan.' 
      });
    }
  }, [token]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (password !== confirmPassword) {
      setStatus({ type: 'error', message: 'Kata sandi tidak cocok.' });
      return;
    }
    
    if (password.length < 6) {
      setStatus({ type: 'error', message: 'Kata sandi harus minimal 6 karakter.' });
      return;
    }
    
    setStatus({ type: '', message: '' });
    setLoading(true);
    
    try {
      const res = await api.post('/auth/reset-password', { 
        token, 
        new_password: password 
      });
      
      setStatus({ 
        type: 'success', 
        message: res.data?.message || 'Kata sandi berhasil direset. Silakan masuk dengan kata sandi baru Anda.' 
      });
      
      setTimeout(() => {
        navigate('/login');
      }, 3000);
    } catch (err) {
      console.error('Reset password error', err);
      setStatus({ 
        type: 'error', 
        message: err.response?.data?.error || err.response?.data?.message || 'Gagal mereset kata sandi. Token mungkin sudah kedaluwarsa.' 
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="main-content" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '80vh' }}>
      <div className="glass-panel animate-fade-in" style={{ width: '100%', maxWidth: '450px' }}>
        <h1 style={{ textAlign: 'center', marginBottom: '0.5rem' }}>Reset Kata Sandi</h1>
        <p style={{ textAlign: 'center', color: 'var(--text-muted)', marginBottom: '2rem' }}>
          Masukkan kata sandi baru untuk akun Anda.
        </p>
        
        {status.message && (
          <div style={{ 
            background: 'var(--surface)', 
            border: status.type === 'success' ? 'var(--border-width) solid var(--success)' : 'var(--border-width) solid var(--danger)',
            boxShadow: status.type === 'success' ? '4px 4px 0px var(--success)' : '4px 4px 0px var(--danger)',
            color: 'var(--text)', 
            padding: '1rem', 
            borderRadius: '8px', 
            marginBottom: '1.5rem', 
            textAlign: 'center', 
            fontSize: '0.875rem',
            fontWeight: 'bold'
          }}>
            {status.message}
          </div>
        )}
        
        {!token ? (
          <div style={{ textAlign: 'center', marginTop: '2rem' }}>
            <Link to="/forgot-password" className="btn btn-primary" style={{ display: 'inline-block', padding: '0.75rem 1.5rem' }}>
              Minta Tautan Baru
            </Link>
          </div>
        ) : status.type === 'success' ? (
          <div style={{ textAlign: 'center', marginTop: '2rem' }}>
            <Link to="/login" className="btn btn-primary" style={{ display: 'inline-block', padding: '0.75rem 1.5rem' }}>
              Ke Halaman Masuk
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="input-group">
              <label>Kata Sandi Baru</label>
              <input 
                type="password" 
                className="input-field" 
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            
            <div className="input-group">
              <label>Konfirmasi Kata Sandi Baru</label>
              <input 
                type="password" 
                className="input-field" 
                placeholder="••••••••"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </div>
            
            <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: '1rem', padding: '1rem' }} disabled={loading}>
              {loading ? 'Menyimpan...' : 'Simpan Kata Sandi Baru'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default ResetPassword;
