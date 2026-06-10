import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const formatRp = (val) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(Number(val) || 0);
};

const CreatorDashboard = () => {
  const [revenue, setRevenue] = useState({ balance: 0, total_earned: 0 });
  const [myAssets, setMyAssets] = useState([]);
  const [withdrawals, setWithdrawals] = useState([]);

  // Modals state
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);

  // Form states
  const [withdrawAmount, setWithdrawAmount] = useState('');
  const [withdrawBank, setWithdrawBank] = useState('');
  const [withdrawAccount, setWithdrawAccount] = useState('');
  const [assetForm, setAssetForm] = useState({ id: null, title: '', description: '', price: '' });
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState({ show: false, message: '', type: 'success' });

  const showToast = (message, type = 'success') => {
    setToast({ show: true, message, type });
    setTimeout(() => setToast({ show: false, message: '', type: 'success' }), 3000);
  };

  useEffect(() => {
    const fetchData = async () => {
      // Loading state handled by initial state
      try {
        const revRes = await api.get('/orders/revenue').catch(() => ({ data: { revenue: 0 } }));
        const totalEarned = revRes.data?.revenue || 0;

        const wdRes = await api.get('/withdrawals/my-withdrawals').catch(() => ({ data: { withdrawals: [] } }));
        const wdList = wdRes.data?.withdrawals || wdRes.data || [];
        const validWd = Array.isArray(wdList) ? wdList : [];
        setWithdrawals(validWd);

        const totalWithdrawn = validWd
          .filter(w => {
            const st = (w.status || '').toLowerCase();
            return st === 'approved' || st === 'pending';
          })
          .reduce((sum, w) => sum + (w.amount || 0), 0);

        setRevenue({
          total_earned: totalEarned,
          balance: totalEarned - totalWithdrawn
        });
      } catch (e) {
        console.error(e);
        setRevenue({ balance: 0, total_earned: 0 });
      }

      try {
        const assetsRes = await api.get('/assets/my-assets').catch(() => ({ data: { assets: [] } }));
        const list = assetsRes.data?.assets || assetsRes.data || [];
        setMyAssets(Array.isArray(list) ? list : []);
      } catch (e) { console.error(e); }

      setLoading(false);
    };

    fetchData();
    const interval = setInterval(fetchData, 15000); // Auto-refresh setiap 15 detik
    return () => clearInterval(interval);
  }, []);

  const handleWithdraw = async (e) => {
    e.preventDefault();
    const amount = parseFloat(withdrawAmount);
    if (!amount || amount <= 0) return;
    if (!withdrawBank.trim() || !withdrawAccount.trim()) {
      showToast('Mohon lengkapi data Bank dan Nomor Rekening!', 'error');
      return;
    }

    if (amount > revenue.balance) {
      showToast('Saldo tidak mencukupi untuk penarikan ini!', 'error');
      return;
    }

    try {
      const res = await api.post('/withdrawals', {
        amount,
        bank_code: withdrawBank,
        account_number: withdrawAccount
      });
      setWithdrawals([res.data, ...withdrawals]);
      setRevenue(prev => ({ ...prev, balance: (prev.balance || 0) - amount }));
      showToast('Permintaan penarikan dana berhasil dikirim!', 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal mengirim permintaan penarikan.', 'error');
    }
    setWithdrawAmount('');
    setWithdrawBank('');
    setWithdrawAccount('');
  };

  const handleUpload = async () => {
    const validUrls = (assetForm.preview_urls || []).filter(u => u.trim() !== '');
    if (!assetForm.title.trim() || !assetForm.price || !assetForm.description.trim() || !assetForm.storage_path?.trim() || validUrls.length === 0) {
      showToast('Semua kolom wajib diisi (Judul, Deskripsi, Harga, Tautan Akses, dan minimal 1 Foto Preview).', 'error');
      return;
    }

    if (assetForm.title.trim().split(/\s+/).filter(w => w).length > 15) {
      showToast('Judul maksimal 15 kata!', 'error');
      return;
    }

    if (assetForm.description.trim().split(/\s+/).filter(w => w).length > 200) {
      showToast('Deskripsi maksimal 200 kata!', 'error');
      return;
    }

    try {
      const payload = {
        title: assetForm.title,
        description: assetForm.description,
        price: parseFloat(assetForm.price),
        preview_url: validUrls.join('|'),
        storage_path: assetForm.storage_path.trim()
      };

      const res = await api.post('/assets', payload);
      setMyAssets([res.data, ...myAssets]);
      showToast('Aset berhasil diunggah!', 'success');
      setShowUploadModal(false);
      setAssetForm({ id: null, title: '', description: '', price: '', storage_path: '', preview_urls: [''] });
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal mengunggah aset.', 'error');
    }
  };

  const handleEdit = async () => {
    const validUrls = (assetForm.preview_urls || []).filter(u => u.trim() !== '');
    if (!assetForm.title.trim() || !assetForm.price || !assetForm.description.trim() || !assetForm.storage_path?.trim() || validUrls.length === 0) {
      showToast('Semua kolom wajib diisi (Judul, Deskripsi, Harga, Tautan Akses, dan minimal 1 Foto Preview).', 'error');
      return;
    }

    if (assetForm.title.trim().split(/\s+/).filter(w => w).length > 15) {
      showToast('Judul maksimal 15 kata!', 'error');
      return;
    }

    if (assetForm.description.trim().split(/\s+/).filter(w => w).length > 200) {
      showToast('Deskripsi maksimal 200 kata!', 'error');
      return;
    }

    try {
      const payload = {
        title: assetForm.title,
        description: assetForm.description,
        price: parseFloat(assetForm.price),
        preview_url: validUrls.join('|'),
        storage_path: assetForm.storage_path.trim()
      };

      const res = await api.put(`/assets/${assetForm.id}`, payload);
      setMyAssets(myAssets.map(a => a.id === assetForm.id ? { ...a, ...res.data } : a));
      showToast('Aset berhasil diperbarui!', 'success');
      setShowEditModal(false);
      setAssetForm({ id: null, title: '', description: '', price: '', preview_urls: [''] });
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal memperbarui aset.', 'error');
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Yakin ingin menghapus karya ini? Anda tidak bisa mengembalikannya.')) return;
    try {
      await api.delete(`/assets/${id}`);
      setMyAssets(myAssets.filter(a => a.id !== id));
      showToast('Aset berhasil dihapus!', 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal menghapus aset.', 'error');
    }
  };

  const openEditModal = (asset) => {
    const urls = (asset.preview_url || '').split('|').filter(u => u.trim() !== '');
    setAssetForm({
      id: asset.id,
      title: asset.title || asset.name || '',
      description: asset.description || '',
      price: asset.price || '',
      preview_urls: urls.length > 0 ? urls : ['']
    });
    setShowEditModal(true);
  };

  const getStatusLabel = (status) => {
    const st = (status || '').toLowerCase();
    const map = { 'pending': 'Menunggu', 'approved': 'Disetujui', 'rejected': 'Ditolak', 'completed': 'Selesai' };
    return map[st] || status;
  };

  return (
    <div className="main-content animate-fade-in">
      {/* Toast Notification */}
      {toast.show && (
        <div style={{
          position: 'fixed', top: '6rem', right: '2rem', zIndex: 9999,
          background: toast.type === 'error' ? 'var(--danger)' : 'var(--success)',
          color: 'var(--background)', padding: '1rem 2rem',
          border: 'var(--border-width) solid var(--border)',
          boxShadow: '4px 4px 0px var(--border)', borderRadius: '8px',
          fontWeight: '700', fontSize: '1.1rem', animation: 'fadeInUp 0.3s ease'
        }}>
          {toast.message}
        </div>
      )}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '3rem' }}>
        <div>
          <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>Studio Kreator</h1>
          <p style={{ color: 'var(--text-muted)' }}>Kelola aset dan pendapatan Anda</p>
        </div>
        <button onClick={() => { setAssetForm({ id: null, title: '', description: '', price: '', preview_urls: [''] }); setShowUploadModal(true); }} className="btn btn-primary">
          + Unggah Aset Baru
        </button>
      </div>

      {/* Revenue Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '2rem', marginBottom: '3rem' }}>
        <div className="glass-panel" style={{ padding: '2rem' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Saldo Tersedia</p>
          <h2 style={{ fontSize: '2.5rem', color: 'var(--success)', marginBottom: '1rem' }}>{formatRp(revenue.balance)}</h2>
          <form onSubmit={handleWithdraw} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <select
                className="input-field"
                value={withdrawBank}
                onChange={(e) => setWithdrawBank(e.target.value)}
                style={{ flex: 1, padding: '0.5rem', cursor: 'pointer' }}
              >
                <option value="" disabled>Pilih Bank / E-Wallet</option>
                <option value="BCA">BCA (Bank Central Asia)</option>
                <option value="MANDIRI">Bank Mandiri</option>
                <option value="BNI">BNI (Bank Negara Indonesia)</option>
                <option value="BRI">BRI (Bank Rakyat Indonesia)</option>
                <option value="BSI">BSI (Bank Syariah Indonesia)</option>
                <option value="CIMB">CIMB Niaga</option>
                <option value="PERMATA">Bank Permata</option>
                <option value="DANA">DANA (E-Wallet)</option>
                <option value="GOPAY">GoPay (E-Wallet)</option>
                <option value="OVO">OVO (E-Wallet)</option>
              </select>
              <input
                type="text"
                className="input-field"
                placeholder="No Rekening"
                value={withdrawAccount}
                onChange={(e) => setWithdrawAccount(e.target.value)}
                style={{ flex: 1, padding: '0.5rem' }}
              />
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <input
                type="number"
                className="input-field"
                placeholder="Jumlah (Rp)"
                value={withdrawAmount}
                onChange={(e) => setWithdrawAmount(e.target.value)}
                style={{ flex: 1, padding: '0.5rem' }}
              />
              <button type="submit" className="btn btn-outline" style={{ padding: '0.5rem 1rem', width: '30%' }}>Tarik Dana</button>
            </div>
          </form>
        </div>

        <div className="glass-panel" style={{ padding: '2rem' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Total Pendapatan</p>
          <h2 style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>{formatRp(revenue.total_earned)}</h2>
          <p style={{ color: 'var(--text-muted)' }}>Sepanjang waktu</p>
        </div>
      </div>

      <div style={{ background: 'var(--cyan)', border: 'var(--border-width) solid var(--border)', boxShadow: '4px 4px 0px var(--border)', borderRadius: '12px', padding: '1rem', marginBottom: '2rem' }}>
        <p style={{ fontSize: '0.85rem', color: 'var(--text)' }}>
          <strong>Informasi Penarikan:</strong> Pencairan dana diproses secara manual oleh Admin ke rekening bank terdaftar Anda. Status akan berubah menjadi "Disetujui" setelah dana berhasil ditransfer.
        </p>
      </div>

      {/* Assets & Withdrawals Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '2rem' }}>
        <div className="glass-panel">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h3 style={{ margin: 0 }}>Aset Saya ({myAssets.length})</h3>
            <Link to="/creator/assets" className="btn btn-primary" style={{ padding: '0.5rem 1rem', textDecoration: 'none' }}>
              Kelola Asset Detail
            </Link>
          </div>
          {myAssets.length === 0 ? (
            <p style={{ color: 'var(--text-muted)', padding: '2rem', textAlign: 'center' }}>
              {loading ? 'Memuat...' : 'Belum ada aset. Klik "Unggah Aset Baru" untuk mulai berjualan.'}
            </p>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                    <th style={{ padding: '1rem' }}>Nama Produk</th>
                    <th style={{ padding: '1rem' }}>Harga</th>
                    <th style={{ padding: '1rem' }}>Terjual</th>
                    <th style={{ padding: '1rem' }}>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {myAssets.map((asset) => (
                    <tr key={asset.id} style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                      <td style={{ padding: '1rem', fontWeight: '500' }}>{asset.title || asset.name || '-'}</td>
                      <td style={{ padding: '1rem' }}>{formatRp(asset.price)}</td>
                      <td style={{ padding: '1rem' }}>{asset.sales_count || 0}</td>
                      <td style={{ padding: '1rem', display: 'flex', gap: '0.5rem' }}>
                        <button
                          onClick={() => openEditModal(asset)}
                          className="btn btn-outline"
                          style={{ padding: '0.25rem 0.75rem', fontSize: '0.85rem' }}
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(asset.id)}
                          className="btn btn-danger"
                          style={{ padding: '0.25rem 0.75rem', fontSize: '0.85rem' }}
                        >
                          Hapus
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <div className="glass-panel">
          <h3 style={{ marginBottom: '1.5rem' }}>Riwayat Penarikan</h3>
          {withdrawals.length === 0 ? (
            <p style={{ color: 'var(--text-muted)' }}>{loading ? 'Memuat...' : 'Belum ada riwayat penarikan.'}</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {withdrawals.map(wd => (
                <div key={wd.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingBottom: '1rem', borderBottom: '1px solid var(--border)' }}>
                  <div>
                    <p style={{ fontWeight: '500' }}>{formatRp(wd.amount)}</p>
                    <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                      {wd.created_at ? new Date(wd.created_at).toLocaleDateString('id-ID') : wd.date}
                    </p>
                  </div>
                  <span className={`badge ${wd.status?.toLowerCase() === 'approved' ? 'badge-success' : wd.status?.toLowerCase() === 'rejected' ? 'badge-danger' : 'badge-warning'}`}>
                    {getStatusLabel(wd.status)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Upload/Edit Modal */}
      {(showUploadModal || showEditModal) && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.6)', zIndex: 99999, overflowY: 'auto', paddingTop: '8rem', paddingBottom: '4rem', paddingLeft: '1rem', paddingRight: '1rem' }}>
          <div className="glass-panel animate-fade-in" style={{ width: '100%', maxWidth: '600px', margin: '0 auto', background: 'var(--surface)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2rem', paddingBottom: '1rem', borderBottom: 'var(--border-width) solid var(--border)' }}>
              <h2 style={{ margin: 0 }}>{showEditModal ? 'Edit Aset' : 'Unggah Aset Baru'}</h2>
              <button onClick={() => { setShowUploadModal(false); setShowEditModal(false); }} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.5rem', lineHeight: 1 }}>×</button>
            </div>

            <div className="input-group">
              <label>Judul Aset (Maks 15 Kata)</label>
              <input type="text" className="input-field" placeholder="Contoh: Minimalist UI Kit"
                value={assetForm.title} onChange={e => setAssetForm({ ...assetForm, title: e.target.value })} />
              <div style={{ textAlign: 'right', fontSize: '0.8rem', color: (assetForm.title || '').trim().split(/\s+/).filter(w=>w).length > 15 ? 'var(--danger)' : 'var(--text-muted)' }}>
                {(assetForm.title || '').trim().split(/\s+/).filter(w=>w).length}/15 Kata
              </div>
            </div>
            <div className="input-group">
              <label>Deskripsi (Maks 200 Kata)</label>
              <textarea className="input-field" rows="3" placeholder="Deskripsikan aset Anda..."
                value={assetForm.description} onChange={e => setAssetForm({ ...assetForm, description: e.target.value })}></textarea>
              <div style={{ textAlign: 'right', fontSize: '0.8rem', color: (assetForm.description || '').trim().split(/\s+/).filter(w=>w).length > 200 ? 'var(--danger)' : 'var(--text-muted)' }}>
                {(assetForm.description || '').trim().split(/\s+/).filter(w=>w).length}/200 Kata
              </div>
            </div>
            <div className="input-group">
              <label>Harga (Rp)</label>
              <input type="number" className="input-field" placeholder="150000"
                value={assetForm.price} onChange={e => setAssetForm({ ...assetForm, price: e.target.value })} />
            </div>

            <div className="input-group" style={{ background: 'var(--primary)', padding: '1rem', borderRadius: '6px', border: 'var(--border-width) solid var(--border)', boxShadow: '4px 4px 0px var(--border)' }}>
              <label style={{ color: 'var(--primary-light)' }}>Tautan Akses File (GDrive, Dropbox, Notion, dll)</label>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.5rem' }}>
                Masukkan link akses yang berisi file mentahan produk Anda. Link ini akan diberikan secara otomatis kepada pembeli setelah mereka membayar.
              </p>
              <input type="text" className="input-field" placeholder="https://..."
                value={assetForm.storage_path || ''} onChange={e => setAssetForm({ ...assetForm, storage_path: e.target.value })} />
            </div>

            <div className="input-group">
              <label>Gambar Preview (Maks 5 Foto, Maks 2 MB/foto)</label>
              <input
                type="file"
                multiple
                accept="image/*"
                className="input-field"
                onChange={async (e) => {
                  const files = Array.from(e.target.files);
                  if (files.length > 5) {
                    showToast('Maksimal 5 foto!', 'error');
                    return;
                  }

                  const base64Urls = [];
                  for (let file of files) {
                    if (file.size > 2 * 1024 * 1024) {
                      showToast(`Ukuran foto ${file.name} melebihi 2 MB!`, 'error');
                      return;
                    }
                    const reader = new FileReader();
                    const promise = new Promise((resolve, reject) => {
                      reader.onload = () => resolve(reader.result);
                      reader.onerror = reject;
                      reader.readAsDataURL(file);
                    });
                    base64Urls.push(await promise);
                  }

                  setAssetForm({ ...assetForm, preview_urls: base64Urls });
                }}
              />
              {assetForm.preview_urls && assetForm.preview_urls.length > 0 && assetForm.preview_urls[0] !== '' && (
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem', flexWrap: 'wrap' }}>
                  {assetForm.preview_urls.map((url, i) => (
                    <div key={i} style={{ width: '60px', height: '60px', borderRadius: '6px', overflow: 'hidden', border: 'var(--border-width) solid var(--border)', position: 'relative' }}>
                      <img src={url} alt="preview" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                      <button 
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          const newUrls = [...assetForm.preview_urls];
                          newUrls.splice(i, 1);
                          setAssetForm({ ...assetForm, preview_urls: newUrls });
                        }}
                        title="Hapus gambar"
                        style={{ position: 'absolute', top: 0, right: 0, background: 'var(--danger)', color: 'white', border: 'none', borderLeft: 'var(--border-width) solid var(--border)', borderBottom: 'var(--border-width) solid var(--border)', borderBottomLeftRadius: '4px', width: '20px', height: '20px', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', fontSize: '0.9rem', fontWeight: 'bold', paddingBottom: '2px' }}
                      >
                        ×
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div style={{ paddingTop: '1rem', marginTop: '1rem', borderTop: 'var(--border-width) solid var(--border)' }}>
              <button
                className="btn btn-primary"
                style={{ width: '100%', padding: '1rem', fontSize: '1.1rem', fontWeight: 'bold' }}
                onClick={showEditModal ? handleEdit : handleUpload}
              >
                {showEditModal ? 'Simpan Perubahan' : 'Unggah Sekarang'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CreatorDashboard;
