import { useState, useEffect, Fragment } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const formatRp = (val) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(Number(val) || 0);
};

const BuyerDashboard = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [expandedOrder, setExpandedOrder] = useState(null);
  const [downloadLinks, setDownloadLinks] = useState({});
  const [assetData, setAssetData] = useState({});

  useEffect(() => {
    const fetchOrders = async () => {
      // Loading state handled by initial state
      try {
        const res = await api.get('/orders/my-orders');
        const list = res.data?.orders || res.data || [];
        const validList = Array.isArray(list) ? list : [];
        setOrders(validList);

        const uniqueAssetIds = [...new Set(validList.map(o => o.asset_id))].filter(Boolean);
        for (const assetId of uniqueAssetIds) {
          try {
            const assetRes = await api.get(`/assets/${assetId}`);
            setAssetData(prev => ({ ...prev, [assetId]: assetRes.data?.title || assetRes.data?.name || `Aset ${assetId.substring(0,6)}` }));
          } catch(err) {
             console.error("Gagal memuat info aset", assetId);
          }
        }
      } catch (e) {
        console.error("Failed to fetch orders", e);
        setOrders([]);
      } finally {
        setLoading(false);
      }
    };
    
    fetchOrders();
    const interval = setInterval(fetchOrders, 15000); // Auto-refresh setiap 15 detik
    return () => clearInterval(interval);
  }, []);

  const getStatusLabel = (status) => {
    const map = {
      'pending': 'Menunggu',
      'paid': 'Lunas',
      'completed': 'Selesai',
      'cancelled': 'Dibatalkan',
    };
    return map[status?.toLowerCase()] || status;
  };

  const toggleExpand = async (orderId, assetId) => {
    if (expandedOrder === orderId) {
      setExpandedOrder(null);
    } else {
      setExpandedOrder(orderId);
      // Fetch real download link if not cached
      if (!downloadLinks[orderId]) {
        try {
          const res = await api.get(`/assets/${assetId}/download-link`);
          setDownloadLinks(prev => ({ ...prev, [orderId]: res.data.download_url }));
        } catch (e) {
          setDownloadLinks(prev => ({ ...prev, [orderId]: `https://example.com/access/${assetId}` }));
        }
      }
    }
  };

  return (
    <div className="main-content animate-fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: '3rem' }}>
        <div>
          <h1 style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>Pustaka Saya</h1>
          <p style={{ color: 'var(--text-muted)' }}>Kelola aset yang telah Anda beli</p>
        </div>
      </div>

      <div className="glass-panel">
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)', color: 'var(--text-muted)' }}>
                <th style={{ padding: '1rem' }}>ID Pesanan</th>
                <th style={{ padding: '1rem' }}>Nama Aset</th>
                <th style={{ padding: '1rem' }}>Jumlah</th>
                <th style={{ padding: '1rem' }}>Total Harga</th>
                <th style={{ padding: '1rem' }}>Status</th>
                <th style={{ padding: '1rem' }}>Tanggal</th>
                <th style={{ padding: '1rem' }}>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order, i) => (
                <Fragment key={order.id}>
                  <tr style={{ borderBottom: expandedOrder === order.id ? 'none' : 'var(--border-width) solid var(--border)' }} className={`stagger-${(i % 3) + 1} animate-fade-in`}>
                    <td style={{ padding: '1rem', fontFamily: 'monospace', fontSize: '0.8rem' }}>{order.id?.substring(0, 8)}...</td>
                    <td style={{ padding: '1rem', fontWeight: '500' }}>
                      <Link to={`/assets/${order.asset_id}`} style={{ color: 'var(--primary)', textDecoration: 'none' }}>
                        {assetData[order.asset_id] ? assetData[order.asset_id] : 'Memuat...'}
                      </Link>
                    </td>
                    <td style={{ padding: '1rem' }}>{order.quantity || 1}</td>
                    <td style={{ padding: '1rem', fontWeight: '500' }}>{formatRp(order.total_price)}</td>
                    <td style={{ padding: '1rem' }}>
                      <span className={`badge ${order.status?.toLowerCase() === 'paid' || order.status?.toLowerCase() === 'completed' ? 'badge-success' : order.status?.toLowerCase() === 'cancelled' ? 'badge-danger' : 'badge-warning'}`}>
                        {getStatusLabel(order.status)}
                      </span>
                    </td>
                    <td style={{ padding: '1rem', color: 'var(--text-muted)', fontSize: '0.875rem' }}>
                      {order.created_at ? new Date(order.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-'}
                    </td>
                    <td style={{ padding: '1rem' }}>
                      {(order.status?.toLowerCase() === 'paid' || order.status?.toLowerCase() === 'completed') ? (
                        <button 
                          onClick={() => toggleExpand(order.id, order.asset_id)}
                          className="btn btn-outline" 
                          style={{ padding: '0.25rem 0.75rem', fontSize: '0.85rem' }}
                        >
                          {expandedOrder === order.id ? 'Tutup' : 'Lihat Akses'}
                        </button>
                      ) : (
                        <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>Belum lunas</span>
                      )}
                    </td>
                  </tr>
                  
                  {/* Expanded Dropdown Row */}
                  {expandedOrder === order.id && (
                    <tr style={{ borderBottom: 'var(--border-width) solid var(--border)' }}>
                      <td colSpan="7" style={{ padding: '0 1rem 1.5rem 1rem' }}>
                        <div style={{ background: 'var(--cyan)', padding: '1.5rem', borderRadius: '8px', border: 'var(--border-width) solid var(--border)', boxShadow: '4px 4px 0px var(--border)' }}>
                          <h4 style={{ marginBottom: '0.5rem', color: 'var(--text)' }}>Akses File Aset</h4>
                          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginBottom: '1rem' }}>
                            Berikut adalah tautan aman untuk mengakses file atau proyek aset ini. Pastikan Anda menyimpannya dengan baik.
                          </p>
                          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', background: 'var(--surface)', border: 'var(--border-width) solid var(--border)', padding: '0.75rem 1rem', borderRadius: '6px', boxShadow: 'inset 2px 2px 0px rgba(0,0,0,0.1)' }}>
                            <span style={{ fontFamily: 'monospace', color: 'var(--text)', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {downloadLinks[order.id] ? downloadLinks[order.id] : 'Memuat tautan...'}
                            </span>
                            {downloadLinks[order.id] && (
                              <a 
                                href={downloadLinks[order.id].startsWith('http') ? downloadLinks[order.id] : `https://${downloadLinks[order.id]}`}
                                target="_blank" 
                                rel="noreferrer"
                                className="btn btn-primary"
                                style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}
                              >
                                Buka Tautan
                              </a>
                            )}
                          </div>
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
            </tbody>
          </table>
          
          {loading && <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>Memuat pustaka...</div>}
          
          {!loading && orders.length === 0 && (
            <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>
              Anda belum melakukan pembelian. <Link to="/" style={{ color: 'var(--primary)' }}>Jelajahi marketplace</Link>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default BuyerDashboard;
