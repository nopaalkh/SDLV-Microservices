import { Link, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

const Navbar = () => {
  const [scrolled, setScrolled] = useState(false);
  const navigate = useNavigate();
  
  // Dummy auth state for now
  const token = localStorage.getItem('token');
  const userRole = localStorage.getItem('role') || 'buyer';

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 50);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    navigate('/login');
  };

  return (
    <nav className={`navbar ${scrolled ? 'scrolled' : ''}`}>
      <Link to="/" className="text-gradient-primary" style={{ fontSize: '1.5rem', fontWeight: '800' }}>
        Azetqu
      </Link>
      
      <div className="nav-links">
        <Link to="/" className="nav-link">Marketplace</Link>
        
        {!token ? (
          <>
            <Link to="/login" className="nav-link">Log in</Link>
            <Link to="/register" className="btn btn-primary">Sign up</Link>
          </>
        ) : (
          <>
            {userRole === 'buyer' && <Link to="/dashboard/buyer" className="nav-link">My Library</Link>}
            {userRole === 'creator' && <Link to="/dashboard/creator" className="nav-link">Creator Studio</Link>}
            {userRole === 'admin' && <Link to="/dashboard/admin" className="nav-link">Admin Panel</Link>}
            
            <button onClick={handleLogout} className="btn btn-outline" style={{ padding: '0.5rem 1rem' }}>
              Log out
            </button>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
