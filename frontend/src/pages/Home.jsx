import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

// Helper: format Rupiah
const formatRp = (val) => {
  const num = Number(val) || 0;
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(num);
};

const Home = () => {
  const [assets, setAssets] = useState([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAssets = async () => {
      setLoading(true);
      try {
        const res = await api.post('/assets/search', { query: search, category: '' });
        const list = res.data?.assets || res.data || [];
        setAssets(Array.isArray(list) ? list : []);
      } catch (error) {
        console.error('Error fetching assets:', error);
        setAssets([]);
      } finally {
        setLoading(false);
      }
    };
    
    const timeoutId = setTimeout(() => {
      fetchAssets();
    }, 400);
    return () => clearTimeout(timeoutId);
  }, [search]);

  return (
    <div className="main-content animate-fade-in">
      {/* Hero Section */}
      <div className="glass-panel" style={{ textAlign: 'center', marginBottom: '3rem', padding: '4rem 2rem' }}>
        <h1 style={{ fontSize: '3rem', marginBottom: '1rem' }} className="text-gradient">
          Temukan Aset Digital Premium
        </h1>
        <p style={{ color: 'var(--text-muted)', fontSize: '1.15rem', maxWidth: '600px', margin: '0 auto 2rem' }}>
          Tingkatkan kualitas proyek Anda dengan UI Kit, Paket Audio, Model 3D, dan banyak lagi karya kelas dunia.
        </p>
        
        <div style={{ maxWidth: '500px', margin: '0 auto', display: 'flex', gap: '1rem' }}>
          <input 
            type="text" 
            className="input-field" 
            placeholder="Cari aset..." 
            style={{ flex: 1 }}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button className="btn btn-primary" disabled={loading} style={{ whiteSpace: 'nowrap' }}>
            {loading ? 'Mencari...' : 'Cari'}
          </button>
        </div>
      </div>

      {/* Asset Grid Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h2>Aset Unggulan</h2>
        <span style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>{assets.length} produk ditemukan</span>
      </div>

      {/* Asset Grid */}
      <div className="grid-cards">
        {assets.map((asset, index) => {
          const title = asset.title || asset.name || 'Tanpa Judul';
          const urls = (asset.preview_url || '').split('|').filter(u => u.trim() !== '');
          const previewImg = urls.length > 0 ? urls[0] : (asset.image_url || 'https://images.unsplash.com/photo-1618761714954-0b8cd0026356?auto=format&fit=crop&w=500&q=80');
          
          return (
            <Link to={`/assets/${asset.id}`} key={asset.id} className={`glass animate-fade-in stagger-${(index % 3) + 1}`} style={{ display: 'block', overflow: 'hidden', transition: 'transform 0.3s ease, box-shadow 0.3s ease' }}>
              <div style={{ height: '200px', width: '100%', overflow: 'hidden' }}>
                <img src={previewImg} alt={title} style={{ width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.3s' }} />
              </div>
              <div style={{ padding: '1.5rem' }}>
                <h3 style={{ marginBottom: '0.5rem', fontSize: '1.2rem' }}>{title}</h3>
                <p style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '1rem' }}>
                  {asset.category || 'Aset Digital'}
                </p>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '1.2rem', fontWeight: '700', color: 'var(--secondary)' }}>
                    {formatRp(asset.price)}
                  </span>
                  <span className="badge badge-primary">LIHAT</span>
                </div>
              </div>
            </Link>
          );
        })}
        {assets.length === 0 && !loading && (
          <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
            Tidak ada aset yang ditemukan. Coba kata kunci pencarian yang lain.
          </div>
        )}
        {loading && (
          <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
            Memuat aset...
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;
