import React, { useEffect } from 'react';
import { useAuth } from '../hooks/useAuth';
import Avatar from '../components/Avatar';
import './ProfilePage.css';

const ProfilePage: React.FC = () => {
  const { user, loading, logout, refreshUser } = useAuth();
  const [loggingOut, setLoggingOut] = React.useState(false);

  useEffect(() => {
    if (user) {
      refreshUser();
    }
  }, []);

  const handleLogout = async () => {
    setLoggingOut(true);
    try {
      await logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setLoggingOut(false);
    }
  };

  if (loading) {
    return (
      <div className="profile-page">
        <div className="profile-container">
          <div className="loading">Loading...</div>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="profile-page">
        <div className="profile-container">
          <div className="not-authenticated">
            <p>Please log in to view your profile.</p>
          </div>
        </div>
      </div>
    );
  }

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="profile-page">
      <div className="profile-container">
        <div className="profile-header">
          <Avatar
            avatarUrl={user.avatarUrl}
            username={user.username}
            email={user.email}
            displayName={user.displayName}
            size={120}
          />
          <h1>{user.displayName}</h1>
          <p className="profile-username">@{user.username}</p>
        </div>

        <div className="profile-content">
          <div className="profile-section">
            <h2>Account Information</h2>
            <div className="info-item">
              <span className="info-label">Email</span>
              <span className="info-value">{user.email}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Username</span>
              <span className="info-value">{user.username}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Display Name</span>
              <span className="info-value">{user.displayName}</span>
            </div>
          </div>

          <div className="profile-section">
            <h2>Account Details</h2>
            <div className="info-item">
              <span className="info-label">Member since</span>
              <span className="info-value">{formatDate(user.createdAt)}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Last updated</span>
              <span className="info-value">{formatDate(user.updatedAt)}</span>
            </div>
          </div>

          <button
            className="logout-button"
            onClick={handleLogout}
            disabled={loggingOut}
          >
            {loggingOut ? 'Logging out...' : 'Log Out'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
