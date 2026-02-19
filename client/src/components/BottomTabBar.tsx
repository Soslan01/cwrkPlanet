import React from 'react';
import './BottomTabBar.css';

export type TabId = 'home' | 'auth' | 'profile';

interface BottomTabBarProps {
  currentTab: TabId;
  onTabChange: (tab: TabId) => void;
  isLoggedIn: boolean;
}

const BottomTabBar: React.FC<BottomTabBarProps> = ({ currentTab, onTabChange, isLoggedIn }) => {
  const authLabel = isLoggedIn ? 'Logout' : 'Login';
  const authIcon = isLoggedIn ? '🚪' : '🔐';

  return (
    <div className="bottom-tab-bar">
      <button
        className={`tab-button ${currentTab === 'home' ? 'active' : ''}`}
        onClick={() => onTabChange('home')}
        aria-label="Home"
      >
        <span className="tab-icon" aria-hidden>🏠</span>
        <span className="tab-label">Home</span>
      </button>
      <button
        className={`tab-button ${currentTab === 'auth' ? 'active' : ''}`}
        onClick={() => onTabChange('auth')}
        aria-label={authLabel}
      >
        <span className="tab-icon" aria-hidden>{authIcon}</span>
        <span className="tab-label">{authLabel}</span>
      </button>
      <button
        className={`tab-button ${currentTab === 'profile' ? 'active' : ''}`}
        onClick={() => onTabChange('profile')}
        aria-label="Profile"
      >
        <span className="tab-icon" aria-hidden>👤</span>
        <span className="tab-label">Profile</span>
      </button>
    </div>
  );
};

export default BottomTabBar;
