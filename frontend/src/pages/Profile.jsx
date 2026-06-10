import { useState, useEffect } from 'react';
import api from '../services/api';
import { useNavigate } from 'react-router-dom';

const Profile = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  
  // Password state
  const [passwordForm, setPasswordForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: ''
  });
  const [passLoading, setPassLoading] = useState(false);
  const [passMessage, setPassMessage] = useState({ type: '', text: '' });
  
  const navigate = useNavigate();

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const res = await api.get('/auth/me');
        setUser(res.data);
      } catch (e) {
        console.error("Failed to fetch profile", e);
        if (e.response?.status === 401) {
          navigate('/login');
        }
      } finally {
        setLoading(false);
      }
    };
    fetchUser();
  }, [navigate]);

  if (loading) {
    return (
      <div className="main-content" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <p style={{ fontSize: '1.5rem', color: 'var(--text-muted)' }}>Memuat profil...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="main-content" style={{ textAlign: 'center', paddingTop: '5rem' }}>
        <h2>Gagal memuat profil</h2>
        <p style={{ color: 'var(--text-muted)' }}>Silakan coba login kembali.</p>
      </div>
    );
  }

  const getRoleLabel = (role) => {
    const map = { 'ADMIN': 'Administrator', 'CREATOR': 'Kreator', 'BUYER': 'Pembeli' };
    return map[role] || role;
  };

  const handlePasswordSubmit = async (e) => {
    e.preventDefault();
    setPassMessage({ type: '', text: '' });
    
    if (passwordForm.new_password !== passwordForm.confirm_password) {
      setPassMessage({ type: 'error', text: 'Konfirmasi sandi baru tidak cocok.' });
      return;
    }
    
    setPassLoading(true);
    try {
      await api.put('/auth/password', {
        current_password: passwordForm.current_password,
        new_password: passwordForm.new_password
      });
      setPassMessage({ type: 'success', text: 'Kata sandi berhasil diperbarui.' });
      setPasswordForm({ current_password: '', new_password: '', confirm_password: '' });
    } catch (err) {
      setPassMessage({ type: 'error', text: err.response?.data?.error || 'Gagal mengubah kata sandi.' });
    } finally {
      setPassLoading(false);
    }
  };

  return (
    <div className="main-content animate-fade-in" style={{ maxWidth: '600px', margin: '0 auto' }}>
      <div style={{ marginBottom: '2rem', textAlign: 'center' }}>
        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>Profil Saya</h1>
        <p style={{ color: 'var(--text-muted)' }}>Informasi akun Azetqu Anda</p>
      </div>

      <div className="glass-panel" style={{ padding: '2.5rem', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '1rem' }}>
          <div style={{ 
            width: '100px', height: '100px', borderRadius: '50%', 
            background: 'linear-gradient(135deg, var(--primary), var(--secondary))',
            display: 'flex', justifyContent: 'center', alignItems: 'center',
            fontSize: '3rem', fontWeight: 'bold', color: 'white'
          }}>
            {user.email.charAt(0).toUpperCase()}
          </div>
        </div>

        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
          <label style={{ fontSize: '0.85rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>Nama Lengkap</label>
          <div style={{ fontSize: '1.1rem', fontWeight: '500' }}>{user.name || 'Tidak ada nama'}</div>
        </div>

        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
          <label style={{ fontSize: '0.85rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>Alamat Email</label>
          <div style={{ fontSize: '1.1rem', fontWeight: '500' }}>{user.email}</div>
        </div>

        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
          <label style={{ fontSize: '0.85rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>Tipe Akun (Peran)</label>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.25rem' }}>
            <span className={`badge ${user.role === 'ADMIN' ? 'badge-primary' : user.role === 'CREATOR' ? 'badge-warning' : 'badge-success'}`} style={{ fontSize: '0.9rem', padding: '0.4rem 1rem' }}>
              {getRoleLabel(user.role)}
            </span>
          </div>
        </div>

        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
          <label style={{ fontSize: '0.85rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>Status Akun</label>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.25rem' }}>
            <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: user.status === 'ACTIVE' ? 'var(--success)' : 'var(--danger)' }}></div>
            <span style={{ fontSize: '1.1rem', fontWeight: '500' }}>{user.status === 'ACTIVE' ? 'Aktif' : 'Diblokir'}</span>
          </div>
        </div>

        <div style={{ borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
          <label style={{ fontSize: '0.85rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.25rem' }}>Bergabung Pada</label>
          <div style={{ fontSize: '1.1rem', fontWeight: '500' }}>
            {new Date(user.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })}
          </div>
        </div>

        <div style={{ marginTop: '2rem' }}>
          <h3 style={{ marginBottom: '1rem', fontSize: '1.2rem', borderBottom: '1px solid var(--border)', paddingBottom: '0.5rem' }}>Ubah Kata Sandi</h3>
          
          {passMessage.text && (
            <div style={{ 
              background: 'var(--surface)', 
              border: passMessage.type === 'success' ? 'var(--border-width) solid var(--success)' : 'var(--border-width) solid var(--danger)',
              boxShadow: passMessage.type === 'success' ? '4px 4px 0px var(--success)' : '4px 4px 0px var(--danger)',
              color: 'var(--text)', 
              padding: '1rem', 
              borderRadius: '8px', 
              marginBottom: '1.5rem',
              fontWeight: 'bold'
            }}>
              {passMessage.text}
            </div>
          )}

          <form onSubmit={handlePasswordSubmit}>
            <div className="input-group">
              <label>Kata Sandi Saat Ini</label>
              <input 
                type="password" 
                className="input-field" 
                value={passwordForm.current_password}
                onChange={(e) => setPasswordForm({...passwordForm, current_password: e.target.value})}
                required
              />
            </div>
            <div className="input-group">
              <label>Kata Sandi Baru</label>
              <input 
                type="password" 
                className="input-field" 
                value={passwordForm.new_password}
                onChange={(e) => setPasswordForm({...passwordForm, new_password: e.target.value})}
                minLength={6}
                required
              />
            </div>
            <div className="input-group">
              <label>Konfirmasi Sandi Baru</label>
              <input 
                type="password" 
                className="input-field" 
                value={passwordForm.confirm_password}
                onChange={(e) => setPasswordForm({...passwordForm, confirm_password: e.target.value})}
                minLength={6}
                required
              />
            </div>
            <button type="submit" className="btn btn-primary" disabled={passLoading}>
              {passLoading ? 'Menyimpan...' : 'Perbarui Sandi'}
            </button>
          </form>
        </div>

      </div>
    </div>
  );
};

export default Profile;

