import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';

const formatRp = (val) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(Number(val) || 0);
};

const AssetDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [asset, setAsset] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [reviewText, setReviewText] = useState('');
  const [reviewRating, setReviewRating] = useState(5);
  const [loading, setLoading] = useState(true);
  const [checkingOut, setCheckingOut] = useState(false);
  const [activeImage, setActiveImage] = useState(0);
  const [showDescription, setShowDescription] = useState(false);

  const [toast, setToast] = useState({ show: false, message: '', type: 'success' });

  const showToast = (message, type = 'success') => {
    setToast({ show: true, message, type });
    setTimeout(() => setToast({ show: false, message: '', type: 'success' }), 3000);
  };

  useEffect(() => {
    const fetchAssetData = async () => {
      setLoading(true);
      try {
        const [assetRes, reviewsRes] = await Promise.all([
          api.get(`/assets/${id}`),
          api.get(`/assets/${id}/reviews`).catch(() => ({ data: { reviews: [] } }))
        ]);
        
        const fetched = assetRes.data;
        setAsset(fetched);
        let list = reviewsRes.data?.reviews || reviewsRes.data;
        setReviews(Array.isArray(list) ? list : []);
      } catch (err) {
        console.error('Failed to fetch asset', err);
        setAsset(null);
      } finally {
        setLoading(false);
      }
    };
    
    fetchAssetData();
  }, [id]);

  const handleCheckout = async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      showToast('Silakan masuk terlebih dahulu untuk membeli aset ini.', 'error');
      navigate('/login');
      return;
    }

    setCheckingOut(true);
    try {
      const res = await api.post('/orders', { asset_id: id, quantity: 1 });
      const snapToken = res.data?.snap_token;
      const orderId = res.data?.id || res.data?.order_id;
      
      if (snapToken) {
        window.snap.pay(snapToken, {
          onSuccess: async function(result) {
            // Karena ini di environment lokal dan Midtrans tidak bisa mengirim Webhook ke localhost,
            // kita panggil secara manual mock-pay untuk mengubah status di database.
            try {
              if (orderId) await api.put(`/orders/${orderId}/mock-pay`);
            } catch (err) {}

            showToast('Pembayaran berhasil! Aset sudah ditambahkan ke Pustaka Anda.', 'success');
            navigate('/dashboard/buyer');
          },
          onPending: function(result) {
            showToast('Menunggu pembayaran Anda diselesaikan.', 'warning');
            navigate('/dashboard/buyer');
          },
          onError: function(result) {
            showToast('Pembayaran gagal atau dibatalkan.', 'error');
          },
          onClose: function() {
            showToast('Pembayaran ditutup. Anda bisa melanjutkannya nanti di halaman Pustaka Saya.', 'warning');
            navigate('/dashboard/buyer');
          }
        });
      } else {
        showToast('Sistem pembayaran gagal memuat token. Silakan cek Pustaka Saya.', 'error');
        navigate('/dashboard/buyer');
      }
    } catch (e) {
      console.error('Checkout error', e);
      showToast(e.response?.data?.error || 'Gagal membuat pesanan. Silakan coba lagi.', 'error');
    } finally {
      setCheckingOut(false);
    }
  };

  const submitReview = async (e) => {
    e.preventDefault();
    if (!reviewText.trim()) return;
    
    try {
      const res = await api.post(`/assets/${id}/reviews`, { comment: reviewText, rating: reviewRating });
      setReviews([res.data, ...reviews]);
      setReviewText('');
      showToast('Ulasan berhasil dikirim!', 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal mengirim ulasan. Pastikan Anda sudah membeli aset ini.', 'error');
    }
    setReviewText('');
  };

  if (loading) {
    return (
      <div className="main-content" style={{ textAlign: 'center', paddingTop: '10rem' }}>
        <div className="text-gradient" style={{ fontSize: '1.5rem' }}>Memuat Detail Aset...</div>
      </div>
    );
  }
  
  if (!asset) {
    return (
      <div className="main-content" style={{ textAlign: 'center', paddingTop: '10rem' }}>
        <h2>Aset tidak ditemukan.</h2>
        <p style={{ color: 'var(--text-muted)', marginTop: '1rem' }}>
          Aset mungkin telah dihapus atau ID-nya tidak valid.
        </p>
      </div>
    );
  }

  const title = asset.title || asset.name || 'Tanpa Judul';
  const urls = (asset.preview_url || '').split('|').filter(u => u.trim() !== '');
  const previewImages = urls.length > 0 ? urls : [asset.image_url || 'https://images.unsplash.com/photo-1618761714954-0b8cd0026356?auto=format&fit=crop&w=1200&q=80'];

  return (
    <div className="main-content animate-fade-in">
      {/* Toast Notification */}
      {toast.show && (
        <div style={{
          position: 'fixed', top: '6rem', right: '2rem', zIndex: 9999,
          background: toast.type === 'error' ? 'var(--danger)' : toast.type === 'warning' ? 'var(--warning)' : 'var(--success)',
          color: toast.type === 'warning' ? 'var(--text)' : 'var(--background)', padding: '1rem 2rem',
          border: 'var(--border-width) solid var(--border)',
          boxShadow: '4px 4px 0px var(--border)', borderRadius: '8px',
          fontWeight: '700', fontSize: '1.1rem', animation: 'fadeInUp 0.3s ease'
        }}>
          {toast.message}
        </div>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(min(100%, 400px), 1fr))', gap: '4rem', marginBottom: '4rem', alignItems: 'start' }}>
        
        {/* Left Column - Images Slider */}
        <div style={{ position: 'sticky', top: '2rem' }}>
          <div 
            className="glass-panel" 
            style={{ padding: '0', overflow: 'hidden', position: 'relative', height: '400px', borderRadius: '24px', cursor: previewImages.length > 1 ? 'pointer' : 'default' }}
            onClick={() => {
              if (previewImages.length > 1) {
                setActiveImage(prev => prev === previewImages.length - 1 ? 0 : prev + 1);
              }
            }}
          >
            <img 
              src={previewImages[activeImage]} 
              alt={title} 
              style={{ width: '100%', height: '100%', objectFit: 'cover', transition: 'opacity 0.3s' }} 
            />
            
            {/* Image Counter Badge */}
            {previewImages.length > 1 && (
              <div style={{ position: 'absolute', top: '1rem', right: '1rem', background: 'var(--surface)', border: 'var(--border-width) solid var(--border)', padding: '0.25rem 0.75rem', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 'bold', boxShadow: '2px 2px 0px var(--border)' }}>
                {activeImage + 1} / {previewImages.length}
              </div>
            )}
          </div>
        </div>
        
        {/* Info */}
        <div className="glass-panel" style={{ display: 'flex', flexDirection: 'column' }}>
          <div style={{ marginBottom: 'auto' }}>
            
            {asset.ai_tool_used && (
              <div style={{ marginBottom: '1rem' }}>
                <span style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>Dibuat dengan: </span>
                <span className="badge badge-warning">{asset.ai_tool_used}</span>
              </div>
            )}

            {asset.resolution && (
              <div style={{ marginBottom: '1rem' }}>
                <span style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>Resolusi: </span>
                <span>{asset.resolution}</span>
              </div>
            )}

            <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem', flexWrap: 'wrap' }}>
              <span className="badge badge-success">✓ Lisensi Komersial</span>
              <span className="badge badge-success">✓ Pembaruan Gratis</span>
            </div>
          </div>

          <div style={{ padding: '1.5rem', background: 'var(--surface-hover)', borderRadius: '8px', border: 'var(--border-width) solid var(--border)', marginTop: '2rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
              <span style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>Terjual:</span>
              <span style={{ fontWeight: '600' }}>{asset.sales_count || 0} kali</span>
            </div>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginTop: '0.5rem' }}>
              <span style={{ fontSize: '2.5rem', fontWeight: '800', color: 'var(--text)', wordBreak: 'break-word' }}>
                {formatRp(asset.price)}
              </span>
              <button 
                onClick={handleCheckout} 
                className="btn btn-primary" 
                style={{ padding: '1rem', fontSize: '1.1rem', width: '100%', fontWeight: 'bold', letterSpacing: '1px' }} 
                disabled={checkingOut}
              >
                {checkingOut ? 'Memproses...' : 'Beli Sekarang'}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Title Section */}
      <div className="glass-panel" style={{ marginBottom: '2rem' }}>
        <span className="badge badge-primary" style={{ marginBottom: '1rem', wordBreak: 'break-word', display: 'inline-block' }}>{asset.category || 'Aset Digital'}</span>
        <h1 style={{ fontSize: '2.5rem', margin: 0, wordBreak: 'break-word' }}>{title}</h1>
      </div>

      {/* Description Section */}
      <div className="glass-panel" style={{ marginBottom: '4rem' }}>
        <div 
          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
          onClick={() => setShowDescription(!showDescription)}
        >
          <h2 style={{ margin: 0 }}>Deskripsi Aset</h2>
          <button className="btn btn-outline" style={{ padding: '0.5rem 1rem' }}>
            {showDescription ? 'Sembunyikan' : 'Tampilkan'}
          </button>
        </div>
        
        {showDescription && (
          <div style={{ marginTop: '2rem', paddingTop: '2rem', borderTop: 'var(--border-width) solid var(--border)', animation: 'fadeInUp 0.3s ease' }}>
            <p style={{ color: 'var(--text-muted)', fontSize: '1.05rem', lineHeight: '1.8', wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}>
              {asset.description || 'Tidak ada deskripsi untuk aset ini.'}
            </p>
          </div>
        )}
      </div>

      {/* Reviews Section */}
      <div className="glass-panel">
        <h2 style={{ marginBottom: '2rem' }}>Ulasan Pelanggan</h2>
        
        {localStorage.getItem('token') && (
          <form onSubmit={submitReview} style={{ marginBottom: '3rem' }}>
            <div className="input-group">
              <label>Rating</label>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                {[1,2,3,4,5].map(n => (
                  <button 
                    key={n} 
                    type="button" 
                    onClick={() => setReviewRating(n)} 
                    style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.5rem', color: n <= reviewRating ? 'var(--warning)' : 'var(--text-muted)' }}
                  >
                    ★
                  </button>
                ))}
              </div>
            </div>
            <div className="input-group">
              <textarea 
                className="input-field" 
                placeholder="Tulis ulasan Anda..." 
                rows="3"
                value={reviewText}
                onChange={(e) => setReviewText(e.target.value)}
              ></textarea>
            </div>
            <button type="submit" className="btn btn-outline">Kirim Ulasan</button>
          </form>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {reviews.length === 0 ? (
            <p style={{color: 'var(--text-muted)'}}>Belum ada ulasan untuk aset ini.</p>
          ) : reviews.map((review, idx) => (
            <div key={review.id || idx} style={{ padding: '1.5rem', borderBottom: '1px solid var(--border)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                <strong style={{ fontSize: '1rem' }}>
                  {review.user_email || review.user || 'Pengguna'}
                </strong>
                <span style={{ color: 'var(--warning)' }}>{'★'.repeat(review.rating || 5)}</span>
              </div>
              <p style={{ color: 'var(--text-muted)' }}>{review.comment || review.text}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default AssetDetail;
