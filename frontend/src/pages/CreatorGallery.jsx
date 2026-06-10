import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const formatRp = (val) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(Number(val) || 0);
};

const CreatorGallery = () => {
  const [myAssets, setMyAssets] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchAssets = async () => {
    setLoading(true);
    try {
      const assetsRes = await api.get('/assets/my-assets');
      const list = assetsRes.data?.assets || assetsRes.data || [];
      setMyAssets(Array.isArray(list) ? list : []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAssets();
  }, []);

  return (
    <div className="main-content animate-fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div>
          <Link to="/dashboard/creator" className="btn btn-outline" style={{ display: 'inline-block', marginBottom: '1rem' }}>
            ← Kembali ke Dasbor
          </Link>
          <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>Galeri Aset Saya</h1>
          <p style={{ color: 'var(--text-muted)' }}>Pratinjau visual seluruh aset digital Anda</p>
        </div>
      </div>

      <div className="grid-cards">
        {loading && myAssets.length === 0 && (
          <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
            Memuat aset...
          </div>
        )}
        
        {!loading && myAssets.length === 0 && (
          <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>
            Belum ada aset. Kembali ke Dasbor untuk mulai mengunggah.
          </div>
        )}

        {myAssets.map((asset, index) => {
          const title = asset.title || asset.name || 'Tanpa Judul';
          const urls = (asset.preview_url || '').split('|').filter(u => u.trim() !== '');
          const previewImg = urls.length > 0 ? urls[0] : (asset.image_url || 'https://images.unsplash.com/photo-1618761714954-0b8cd0026356?auto=format&fit=crop&w=500&q=80');
          
          return (
            <div key={asset.id} className={`glass animate-fade-in stagger-${(index % 3) + 1}`} style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
              <a href={`/assets/${asset.id}`} target="_blank" rel="noreferrer" style={{ height: '200px', width: '100%', overflow: 'hidden', borderBottom: 'var(--border-width) solid var(--border)', display: 'block' }}>
                <img src={previewImg} alt={title} style={{ width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.3s' }} />
              </a>
              <div style={{ padding: '1.5rem', flex: 1, display: 'flex', flexDirection: 'column' }}>
                <h3 style={{ marginBottom: '0.5rem', fontSize: '1.2rem' }}>
                  <a href={`/assets/${asset.id}`} target="_blank" rel="noreferrer" style={{ color: 'var(--text)', textDecoration: 'none' }}>
                    {title}
                  </a>
                </h3>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '1rem', display: 'flex', justifyContent: 'space-between' }}>
                  <span>Terjual: <strong>{asset.sales_count || 0}</strong></span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 'auto' }}>
                  <span style={{ fontSize: '1.2rem', fontWeight: '800', color: 'var(--secondary)' }}>
                    {formatRp(asset.price)}
                  </span>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <a href={`/assets/${asset.id}`} target="_blank" rel="noreferrer" className="btn btn-primary" style={{ padding: '0.25rem 0.75rem', fontSize: '0.85rem', textDecoration: 'none', display: 'flex', alignItems: 'center' }}>Lihat</a>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

    </div>
  );
};

export default CreatorGallery;
