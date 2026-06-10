import { useState, useEffect } from 'react';
import api from '../services/api';

const formatRp = (val) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(Number(val) || 0);
};

const AdminDashboard = () => {
  const [stats, setStats] = useState({});
  const [users, setUsers] = useState([]);
  const [withdrawals, setWithdrawals] = useState([]);
  const [assets, setAssets] = useState([]);
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('users'); // 'users', 'withdrawals', 'assets', 'orders'

  const [toast, setToast] = useState({ show: false, message: '', type: 'success' });

  const showToast = (message, type = 'success') => {
    setToast({ show: true, message, type });
    setTimeout(() => setToast({ show: false, message: '', type: 'success' }), 3000);
  };

  // User filters
  const [userFilterRole, setUserFilterRole] = useState('ALL');
  const [userSearchQuery, setUserSearchQuery] = useState('');
  
  // Asset & Withdrawal filters
  const [assetSearchQuery, setAssetSearchQuery] = useState('');
  const [withdrawalSearchQuery, setWithdrawalSearchQuery] = useState('');

  useEffect(() => {
    const fetchAdminData = async () => {
      // Hanya set loading true saat pertama kali render (already handled by initial state)

      try {
        const statsRes = await api.get('/admin/orders/revenue');
        setStats(statsRes.data);
      } catch (e) { setStats({ platform_revenue: 0 }); }

      try {
        const usersRes = await api.get('/admin/users');
        const list = usersRes.data?.users || usersRes.data || [];
        setUsers(Array.isArray(list) ? list : []);
      } catch (e) { setUsers([]); }

      try {
        const wdRes = await api.get('/admin/withdrawals');
        const wdList = wdRes.data?.withdrawals || wdRes.data || [];
        setWithdrawals(Array.isArray(wdList) ? wdList : []);
      } catch (e) { setWithdrawals([]); }

      try {
        const asRes = await api.get('/admin/assets');
        const asList = asRes.data?.assets || asRes.data || [];
        setAssets(Array.isArray(asList) ? asList : []);
      } catch (e) { setAssets([]); }

      try {
        const orRes = await api.get('/admin/orders');
        const orList = orRes.data?.orders || orRes.data || [];
        setOrders(Array.isArray(orList) ? orList : []);
      } catch (e) { setOrders([]); }

      setLoading(false);
    };
    
    fetchAdminData();
    const interval = setInterval(fetchAdminData, 15000); // Polling setiap 15 detik
    return () => clearInterval(interval);
  }, []);

  const handleSuspend = async (userId, currentStatus) => {
    const endpoint = currentStatus === 'ACTIVE' 
      ? `/admin/users/${userId}/suspend` 
      : `/admin/users/${userId}/unsuspend`;
    
    try {
      await api.put(endpoint);
      setUsers(users.map(u => u.id === userId 
        ? { ...u, status: currentStatus === 'ACTIVE' ? 'SUSPENDED' : 'ACTIVE' } 
        : u
      ));
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal mengubah status pengguna.', 'error');
    }
  };

  const handleDeleteUser = async (userId) => {
    if (!window.confirm('Yakin ingin menghapus pengguna ini beserta seluruh datanya? Tindakan ini tidak bisa dibatalkan.')) return;
    try {
      await api.delete(`/admin/users/${userId}`);
      setUsers(users.filter(u => u.id !== userId));
      showToast('Pengguna berhasil dihapus secara permanen.', 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal menghapus pengguna.', 'error');
    }
  };

  const handleWithdrawalAction = async (id, action) => {
    try {
      await api.put(`/admin/withdrawals/${id}/${action}`);
      setWithdrawals(withdrawals.map(w => w.id === id ? { ...w, status: action === 'approve' ? 'approved' : 'rejected' } : w));
      showToast(`Penarikan berhasil di${action === 'approve' ? 'setujui' : 'tolak'}.`, 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal memproses penarikan.', 'error');
    }
  };

  const handleTakedownAsset = async (id) => {
    if (!window.confirm('Yakin ingin melakukan takedown (hapus permanen) pada aset ini?')) return;
    try {
      await api.delete(`/admin/assets/${id}`);
      setAssets(assets.filter(a => a.id !== id));
      showToast('Aset berhasil dihapus dari platform.', 'success');
    } catch (e) {
      showToast(e.response?.data?.error || 'Gagal menghapus aset.', 'error');
    }
  };

  const getRoleLabel = (role) => {
    const map = { 'ADMIN': 'Admin', 'CREATOR': 'Kreator', 'BUYER': 'Pembeli' };
    return map[role] || role;
  };

  if (loading) {
    return (
      <div className="main-content" style={{ textAlign: 'center', paddingTop: '10rem' }}>
        <div className="text-gradient" style={{ fontSize: '1.5rem' }}>Memuat Panel Admin...</div>
      </div>
    );
  }

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

      <div style={{ marginBottom: '3rem' }}>
        <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>Panel Kontrol Admin</h1>
        <p style={{ color: 'var(--text-muted)' }}>Analitik dan manajemen platform Azetqu</p>
      </div>

      {/* Stat Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '2rem', marginBottom: '3rem' }}>
        <div className="glass-panel" style={{ padding: '2rem', borderLeft: '4px solid var(--primary)' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Pendapatan Platform</p>
          <h2 style={{ fontSize: '2rem' }}>{formatRp(stats.platform_revenue || stats.total_revenue)}</h2>
        </div>
        <div className="glass-panel" style={{ padding: '2rem', borderLeft: '4px solid var(--secondary)' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Total Pengguna</p>
          <h2 style={{ fontSize: '2rem' }}>{users.length}</h2>
        </div>
        <div className="glass-panel" style={{ padding: '2rem', borderLeft: '4px solid var(--success)' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Total Aset</p>
          <h2 style={{ fontSize: '2rem' }}>{assets.length}</h2>
        </div>
        <div className="glass-panel" style={{ padding: '2rem', borderLeft: '4px solid var(--warning)' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '0.5rem' }}>Penarikan Tertunda</p>
          <h2 style={{ fontSize: '2rem' }}>{withdrawals.filter(w => (w.status || 'pending').toLowerCase() === 'pending').length}</h2>
        </div>
      </div>

      {/* Tabs navigation */}
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem', flexWrap: 'wrap' }}>
        <button className={`btn ${activeTab === 'users' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setActiveTab('users')}>Pengguna</button>
        <button className={`btn ${activeTab === 'assets' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setActiveTab('assets')}>Manajemen Aset</button>
        <button className={`btn ${activeTab === 'withdrawals' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setActiveTab('withdrawals')}>Persetujuan Penarikan</button>
        <button className={`btn ${activeTab === 'orders' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setActiveTab('orders')}>Riwayat Transaksi</button>
      </div>

      {/* Tab Contents */}
      {activeTab === 'users' && (
        <div className="glass-panel animate-fade-in">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <h3>Manajemen Pengguna ({users.length})</h3>
            <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
              <input 
                type="text" 
                className="input-field" 
                placeholder="Cari nama/email..." 
                value={userSearchQuery}
                onChange={e => setUserSearchQuery(e.target.value)}
                style={{ padding: '0.5rem 1rem', width: '200px' }}
              />
              <select 
                className="input-field" 
                value={userFilterRole}
                onChange={e => setUserFilterRole(e.target.value)}
                style={{ padding: '0.5rem 1rem', width: '150px' }}
              >
                <option value="ALL">Semua Peran</option>
                <option value="BUYER">Pembeli</option>
                <option value="CREATOR">Penjual (Kreator)</option>
                <option value="ADMIN">Admin</option>
              </select>
            </div>
          </div>

          {users.length === 0 ? (
            <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>Tidak ada data pengguna.</p>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                    <th style={{ padding: '1rem' }}>Pengguna</th>
                    <th style={{ padding: '1rem' }}>Peran</th>
                    <th style={{ padding: '1rem' }}>Status</th>
                    <th style={{ padding: '1rem' }}>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {users.filter(u => {
                    if (userFilterRole !== 'ALL' && u.role !== userFilterRole) return false;
                    if (userSearchQuery) {
                      const q = userSearchQuery.toLowerCase();
                      const name = (u.name || '').toLowerCase();
                      const email = (u.email || '').toLowerCase();
                      if (!name.includes(q) && !email.includes(q)) return false;
                    }
                    return true;
                  }).map((user) => (
                    <tr key={user.id} style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                      <td style={{ padding: '1rem' }}>
                        <p style={{ fontWeight: '500' }}>{user.name || user.email}</p>
                        {user.name && <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{user.email}</p>}
                      </td>
                      <td style={{ padding: '1rem' }}>
                        <span className={`badge ${user.role === 'ADMIN' ? 'badge-primary' : user.role === 'CREATOR' ? 'badge-warning' : 'badge-success'}`}>
                          {getRoleLabel(user.role)}
                        </span>
                      </td>
                      <td style={{ padding: '1rem' }}>
                        <span className={`badge ${user.status === 'ACTIVE' ? 'badge-success' : 'badge-danger'}`}>
                          {user.status === 'ACTIVE' ? 'Aktif' : 'Diblokir'}
                        </span>
                      </td>
                      <td style={{ padding: '1rem' }}>
                        {user.role !== 'ADMIN' && (
                          <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <button 
                              onClick={() => handleSuspend(user.id, user.status)}
                              className={`btn ${user.status === 'ACTIVE' ? 'btn-danger' : 'btn-outline'}`} 
                              style={{ padding: '0.25rem 0.75rem', fontSize: '0.8rem' }}
                            >
                              {user.status === 'ACTIVE' ? 'Blokir' : 'Buka Blokir'}
                            </button>
                            <button 
                              onClick={() => handleDeleteUser(user.id)}
                              className="btn btn-outline" 
                              style={{ padding: '0.25rem 0.75rem', fontSize: '0.8rem', color: 'var(--danger)', borderColor: 'var(--danger)' }}
                            >
                              Hapus
                            </button>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {activeTab === 'assets' && (
        <div className="glass-panel animate-fade-in">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <h3>Manajemen Aset ({assets.length})</h3>
            <input 
              type="text" 
              className="input-field" 
              placeholder="Cari judul/kategori..." 
              value={assetSearchQuery}
              onChange={e => setAssetSearchQuery(e.target.value)}
              style={{ padding: '0.5rem 1rem', width: '250px' }}
            />
          </div>
          {assets.length === 0 ? (
            <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>Tidak ada aset di platform.</p>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                    <th style={{ padding: '1rem' }}>Judul Aset</th>
                    <th style={{ padding: '1rem' }}>Kreator ID</th>
                    <th style={{ padding: '1rem' }}>Harga</th>
                    <th style={{ padding: '1rem' }}>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {assets.filter(a => {
                    if (!assetSearchQuery) return true;
                    const q = assetSearchQuery.toLowerCase();
                    const title = (a.title || a.name || '').toLowerCase();
                    const cat = (a.category || '').toLowerCase();
                    return title.includes(q) || cat.includes(q);
                  }).map((asset) => (
                    <tr key={asset.id} style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                      <td style={{ padding: '1rem' }}>
                        <p style={{ fontWeight: '500' }}>{asset.title || asset.name}</p>
                        <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{asset.category}</p>
                      </td>
                      <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>
                        {asset.created_by?.substring(0, 8)}...
                      </td>
                      <td style={{ padding: '1rem' }}>{formatRp(asset.price)}</td>
                      <td style={{ padding: '1rem' }}>
                        <button 
                          onClick={() => handleTakedownAsset(asset.id)}
                          className="btn btn-danger" 
                          style={{ padding: '0.25rem 0.75rem', fontSize: '0.8rem' }}
                        >
                          Takedown
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {activeTab === 'withdrawals' && (
        <div className="glass-panel animate-fade-in">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
            <h3>Persetujuan Penarikan ({withdrawals.length})</h3>
            <input 
              type="text" 
              className="input-field" 
              placeholder="Cari ID Kreator/Bank..." 
              value={withdrawalSearchQuery}
              onChange={e => setWithdrawalSearchQuery(e.target.value)}
              style={{ padding: '0.5rem 1rem', width: '250px' }}
            />
          </div>
          <p style={{ color: 'var(--text)', fontSize: '0.85rem', marginBottom: '1.5rem', background: 'var(--warning)', padding: '1rem', borderRadius: '8px', border: 'var(--border-width) solid var(--border)', boxShadow: '4px 4px 0px var(--border)' }}>
            <strong>PENTING:</strong> Karena sistem belum terhubung ke Payment Gateway Pencairan, Anda diwajibkan untuk mentransfer dana secara manual via Bank ke rekening kreator <strong>sebelum</strong> menekan tombol "Setujui".
          </p>
          {(() => {
            const searchQ = withdrawalSearchQuery.toLowerCase();
            const filteredWithdrawals = withdrawals.filter(w => {
              if (!searchQ) return true;
              const uID = (w.creator_id || w.user_id || '').toLowerCase();
              const bank = (w.bank_code || '').toLowerCase();
              return uID.includes(searchQ) || bank.includes(searchQ);
            });

            const pending = filteredWithdrawals.filter(w => (w.status || 'pending').toLowerCase() === 'pending');
            const history = filteredWithdrawals
              .filter(w => (w.status || 'pending').toLowerCase() !== 'pending')
              .sort((a, b) => {
                // Gunakan id atau timestamp untuk sorting terbaru di atas
                // Jika tidak ada timestamp, anggap ID baru nilainya lebih besar (terutama jika integer)
                if (b.created_at && a.created_at) return new Date(b.created_at) - new Date(a.created_at);
                if (b.updated_at && a.updated_at) return new Date(b.updated_at) - new Date(a.updated_at);
                return b.id > a.id ? 1 : -1;
              });

            return (
              <>
                <h4 style={{ marginBottom: '1rem' }}>Menunggu Persetujuan ({pending.length})</h4>
                {pending.length === 0 ? (
                  <p style={{ color: 'var(--text-muted)', marginBottom: '3rem' }}>Tidak ada penarikan tertunda.</p>
                ) : (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '1rem', marginBottom: '3rem' }}>
                    {pending.map(wd => (
                      <div key={wd.id} style={{ background: 'var(--surface-hover)', padding: '1.5rem', borderRadius: '12px', border: 'var(--border-width) solid var(--border)' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
                          <div>
                            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Kreator</span>
                            <div style={{ fontWeight: '600', fontSize: '1rem' }}>{users.find(u => u.id === (wd.creator_id || wd.user_id))?.name || users.find(u => u.id === (wd.creator_id || wd.user_id))?.email || 'Anonim'}</div>
                            <div style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: 'var(--text-muted)' }}>{(wd.creator_id || wd.user_id)?.substring(0, 8)}...</div>
                          </div>
                          <div style={{ textAlign: 'right' }}>
                            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Jumlah</span>
                            <div style={{ color: 'var(--success)', fontWeight: '700', fontSize: '1.2rem' }}>{formatRp(wd.amount)}</div>
                          </div>
                        </div>
                        
                        <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '1.5rem', background: 'var(--surface)', padding: '0.5rem', borderRadius: '8px', border: 'var(--border-width) solid var(--border)' }}>
                          Bank: {wd.bank_code || 'BCA'} - {wd.account_number || '1234567890'}
                        </div>

                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          <button onClick={() => handleWithdrawalAction(wd.id, 'approve')} className="btn btn-primary" style={{ flex: 1, padding: '0.5rem', fontSize: '0.9rem' }}>Setujui</button>
                          <button onClick={() => handleWithdrawalAction(wd.id, 'reject')} className="btn btn-outline" style={{ flex: 1, padding: '0.5rem', fontSize: '0.9rem' }}>Tolak</button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                <h4 style={{ marginBottom: '1rem', paddingTop: '2rem', borderTop: 'var(--border-width) solid var(--border)' }}>Riwayat Penarikan ({history.length})</h4>
                {history.length === 0 ? (
                  <p style={{ color: 'var(--text-muted)' }}>Belum ada riwayat penarikan.</p>
                ) : (
                  <div style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                      <thead>
                        <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                          <th style={{ padding: '1rem' }}>Nama Kreator</th>
                          <th style={{ padding: '1rem' }}>ID Kreator</th>
                          <th style={{ padding: '1rem' }}>Bank Tujuan</th>
                          <th style={{ padding: '1rem' }}>Jumlah</th>
                          <th style={{ padding: '1rem' }}>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {history.map((wd) => {
                          const cId = wd.creator_id || wd.user_id;
                          const user = users.find(u => u.id === cId) || {};
                          return (
                          <tr key={wd.id} style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                            <td style={{ padding: '1rem', fontWeight: 'bold' }}>{user.name || user.email || 'Anonim'}</td>
                            <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>{cId?.substring(0, 8)}...</td>
                            <td style={{ padding: '1rem' }}>
                              <p style={{ fontWeight: '500' }}>{wd.bank_code || 'BCA'}</p>
                              <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{wd.account_number || '-'}</p>
                            </td>
                            <td style={{ padding: '1rem', fontWeight: 'bold' }}>{formatRp(wd.amount)}</td>
                            <td style={{ padding: '1rem' }}>
                              <span className={`badge ${wd.status?.toLowerCase() === 'approved' ? 'badge-success' : 'badge-danger'}`}>
                                {wd.status?.toLowerCase() === 'approved' ? 'Disetujui' : 'Ditolak'}
                              </span>
                            </td>
                          </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </>
            );
          })()}
        </div>
      )}

      {activeTab === 'orders' && (
        <div className="glass-panel animate-fade-in">
          <h3 style={{ marginBottom: '1.5rem' }}>Semua Riwayat Transaksi ({orders.length})</h3>
          {orders.length === 0 ? (
            <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>Tidak ada transaksi di platform.</p>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                    <th style={{ padding: '1rem' }}>ID Transaksi</th>
                    <th style={{ padding: '1rem' }}>Aset ID</th>
                    <th style={{ padding: '1rem' }}>Pembeli ID</th>
                    <th style={{ padding: '1rem' }}>Total Harga</th>
                    <th style={{ padding: '1rem' }}>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map((order) => (
                    <tr key={order.id} style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                      <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>{order.id?.substring(0, 8)}...</td>
                      <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>{order.asset_id?.substring(0, 8)}...</td>
                      <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>{order.user_id?.substring(0, 8)}...</td>
                      <td style={{ padding: '1rem' }}>{formatRp(order.total_price)}</td>
                      <td style={{ padding: '1rem' }}>
                        <span className={`badge ${order.status === 'paid' || order.status === 'completed' ? 'badge-success' : 'badge-warning'}`}>
                          {order.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

    </div>
  );
};

export default AdminDashboard;
