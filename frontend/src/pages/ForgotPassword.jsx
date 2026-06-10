import { useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const ForgotPassword = () => {
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState({ type: '', message: '' });
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus({ type: '', message: '' });
    setLoading(true);
    
    try {
      const res = await api.post('/auth/forgot-password', { email });
      setStatus({ 
        type: 'success', 
        message: res.data?.message || 'Tautan untuk mereset kata sandi telah dikirim ke email Anda.' 
      });
      setEmail('');
    } catch (err) {
      console.error('Forgot password error', err);
      setStatus({ 
        type: 'error', 
        message: err.response?.data?.error || err.response?.data?.message || 'Gagal mengirim email reset kata sandi. Silakan coba lagi.' 
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="main-content" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '80vh' }}>
      <div className="glass-panel animate-fade-in" style={{ width: '100%', maxWidth: '450px' }}>
        <h1 style={{ textAlign: 'center', marginBottom: '0.5rem' }}>Lupa Kata Sandi</h1>
        <p style={{ textAlign: 'center', color: 'var(--text-muted)', marginBottom: '2rem' }}>
          Masukkan alamat email Anda yang terdaftar dan kami akan mengirimkan tautan untuk mengatur ulang kata sandi Anda.
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
        
        <form onSubmit={handleSubmit}>
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
          
          <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: '1rem', padding: '1rem' }} disabled={loading}>
            {loading ? 'Mengirim...' : 'Kirim Tautan Reset'}
          </button>
        </form>
        
        <p style={{ textAlign: 'center', marginTop: '2rem', color: 'var(--text-muted)' }}>
          Ingat kata sandi Anda? <Link to="/login" style={{ color: 'var(--primary)', fontWeight: '600' }}>Kembali ke Masuk</Link>
        </p>
      </div>
    </div>
  );
};

export default ForgotPassword;

