import React, { useState, useEffect } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import BottomTabBar, { TabId } from './components/BottomTabBar';
import AppHeader from './components/AppHeader';
import ParticleBackground from './components/ParticleBackground';
import WelcomePage from './pages/WelcomePage';
import LoginPage from './pages/LoginPage';
import ProfilePage from './pages/ProfilePage';
import './styles/globals.css';

const AppContent: React.FC = () => {
  const [currentTab, setCurrentTab] = useState<TabId>('home');
  const { user, loading, logout } = useAuth();

  useEffect(() => {
    if (loading) return;
    if (user && currentTab === 'auth') {
      setCurrentTab('profile');
    }
    if (!user && currentTab === 'profile') {
      setCurrentTab('home');
    }
  }, [user, loading, currentTab]);

  const handleTabChange = (tab: TabId) => {
    if (tab === 'profile' && !user) {
      setCurrentTab('auth');
      return;
    }
    if (tab === 'auth' && user) {
      logout();
      setCurrentTab('home');
      return;
    }
    setCurrentTab(tab);
  };

  const handleHeaderLoginClick = () => {
    setCurrentTab('auth');
  };

  const handleHeaderLogoutClick = () => {
    logout();
    setCurrentTab('home');
  };

  return (
    <div className="app">
      <ParticleBackground />
      <AppHeader
        isLoggedIn={!!user}
        onLoginClick={handleHeaderLoginClick}
        onLogoutClick={handleHeaderLogoutClick}
      />
      <main className="app-content">
        {currentTab === 'home' && <WelcomePage />}
        {currentTab === 'auth' && <LoginPage />}
        {currentTab === 'profile' && <ProfilePage />}
      </main>
      <BottomTabBar
        currentTab={currentTab}
        onTabChange={handleTabChange}
        isLoggedIn={!!user}
      />
    </div>
  );
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
};

export default App;
