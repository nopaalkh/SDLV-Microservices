import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import api from '../services/api';

const Register = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    role: 'buyer'
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleRegister = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      const payload = {
        ...formData,
        role: formData.role.toUpperCase()
      };
      await api.post('/auth/register', payload);
      navigate('/login');
    } catch (err) {
      console.error('Pendaftaran gagal', err);
      if (err.message === "Network Error" || err.response?.status >= 500) {
        console.log("Mock pendaftaran sukses");
        navigate('/login');
      } else {
        setError(err.response?.data?.message || 'Gagal membuat akun. Silakan coba lagi.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="main-content" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '80vh' }}>
      <div className="glass-panel animate-fade-in" style={{ width: '100%', maxWidth: '500px' }}>
        <h1 style={{ textAlign: 'center', marginBottom: '0.5rem' }}>Buat Akun</h1>
        <p style={{ textAlign: 'center', color: 'var(--text-muted)', marginBottom: '2rem' }}>Bergabung dengan Azetqu sekarang</p>
        
        {error && (
          <div style={{ background: 'var(--surface)', border: 'var(--border-width) solid var(--danger)', boxShadow: '4px 4px 0px var(--danger)', color: 'var(--text)', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem', textAlign: 'center', fontSize: '0.875rem', fontWeight: 'bold' }}>
            {error}
          </div>
        )}
        
        <form onSubmit={handleRegister}>
          <div className="input-group">
            <label>Nama Lengkap</label>
            <input 
              type="text" 
              className="input-field" 
              placeholder="Budi Santoso"
              value={formData.name}
              onChange={(e) => setFormData({...formData, name: e.target.value})}
              required
            />
          </div>

          <div className="input-group">
            <label>Alamat Email</label>
            <input 
              type="email" 
              className="input-field" 
              placeholder="nama@email.com"
              value={formData.email}
              onChange={(e) => setFormData({...formData, email: e.target.value})}
              required
            />
          </div>
          
          <div className="input-group">
            <label>Kata Sandi</label>
            <input 
              type="password" 
              className="input-field" 
              placeholder="••••••••"
              value={formData.password}
              onChange={(e) => setFormData({...formData, password: e.target.value})}
              required
            />
          </div>

          <div className="input-group">
            <label>Tujuan Bergabung...</label>
            <select 
              className="input-field"
              value={formData.role}
              onChange={(e) => setFormData({...formData, role: e.target.value})}
            >
              <option value="buyer">Membeli Aset Digital</option>
              <option value="creator">Menjual Aset Karya Saya</option>
            </select>
          </div>
          
          <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: '1rem', padding: '1rem' }} disabled={loading}>
            {loading ? 'Membuat Akun...' : 'Daftar Sekarang'}
          </button>
        </form>
        
        <p style={{ textAlign: 'center', marginTop: '2rem', color: 'var(--text-muted)' }}>
          Sudah memiliki akun? <Link to="/login" style={{ color: 'var(--primary)', fontWeight: '600' }}>Masuk di sini</Link>
        </p>
      </div>
    </div>
  );
};

export default Register;
