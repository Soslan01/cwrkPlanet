import React from 'react';
import { useTheme } from '../context/ThemeContext';
import './AppHeader.css';

interface AppHeaderProps {
  isLoggedIn: boolean;
  onLoginClick: () => void;
  onLogoutClick: () => void;
}

const AppHeader: React.FC<AppHeaderProps> = ({ isLoggedIn, onLoginClick, onLogoutClick }) => {
  const { theme, toggleTheme } = useTheme();

  return (
    <header className="app-header">
      <div className="app-header-inner">
        <span className="app-header-title">Client</span>
        <div className="app-header-actions">
          <button
            type="button"
            className="app-header-btn app-header-btn-icon"
            onClick={toggleTheme}
            aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
            title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
          <button
            type="button"
            className="app-header-btn app-header-btn-auth"
            onClick={isLoggedIn ? onLogoutClick : onLoginClick}
            aria-label={isLoggedIn ? 'Logout' : 'Login'}
          >
            {isLoggedIn ? 'Logout' : 'Login'}
          </button>
        </div>
      </div>
    </header>
  );
};

export default AppHeader;
