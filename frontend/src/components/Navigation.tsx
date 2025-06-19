import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const Navigation: React.FC = () => {
  const { user, logout } = useAuth();

  return (
    <nav className="navigation">
      <Link to="/" className="nav-brand">
        ⚓ Battleship
      </Link>
      
      {user ? (
        <>
          <ul className="nav-links">
            <li><Link to="/dashboard">Dashboard</Link></li>
            <li><Link to="/leaderboard">Leaderboard</Link></li>
          </ul>
          
          <div className="nav-user">
            <span>Welcome, {user.username}!</span>
            <button onClick={logout}>Logout</button>
          </div>
        </>
      ) : (
        <ul className="nav-links">
          <li><Link to="/login">Login</Link></li>
          <li><Link to="/register">Register</Link></li>
        </ul>
      )}
    </nav>
  );
};

export default Navigation;
