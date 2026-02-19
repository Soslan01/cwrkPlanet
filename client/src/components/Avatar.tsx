import React from 'react';
import './Avatar.css';

interface AvatarProps {
  avatarUrl?: string | null;
  username?: string;
  email?: string;
  displayName?: string;
  size?: number;
}

const Avatar: React.FC<AvatarProps> = ({
  avatarUrl,
  username,
  email,
  displayName,
  size = 64,
}) => {
  // Generate initials from displayName or username
  const getInitials = (): string => {
    if (displayName) {
      const parts = displayName.trim().split(/\s+/);
      if (parts.length >= 2) {
        return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
      }
      return displayName.substring(0, 2).toUpperCase();
    }
    if (username) {
      return username.substring(0, 2).toUpperCase();
    }
    return '??';
  };

  // Generate color from username or email hash
  const generateColor = (): string => {
    const str = username || email || 'default';
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    // Generate a pleasant color (avoid too dark or too light)
    const hue = Math.abs(hash) % 360;
    const saturation = 65 + (Math.abs(hash) % 20); // 65-85%
    const lightness = 45 + (Math.abs(hash) % 15); // 45-60%
    
    return `hsl(${hue}, ${saturation}%, ${lightness}%)`;
  };

  // Generate Gravatar identicon URL
  const generateGravatarUrl = (): string => {
    const str = email || username || 'default';
    // Simple hash function (not cryptographically secure, but fine for this use case)
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    // Convert to positive hex string
    const hexHash = Math.abs(hash).toString(16).padStart(8, '0');
    return `https://www.gravatar.com/avatar/${hexHash}?d=identicon&s=${size}`;
  };

  if (avatarUrl) {
    return (
      <img
        src={avatarUrl}
        alt={displayName || username || 'Avatar'}
        className="avatar"
        style={{ width: size, height: size }}
      />
    );
  }

  // Use Gravatar identicon as default (GitLab-style)
  return (
    <img
      src={generateGravatarUrl()}
      alt={displayName || username || 'Avatar'}
      className="avatar"
      style={{ width: size, height: size }}
      onError={(e) => {
        // Fallback to initials if Gravatar fails
        const target = e.target as HTMLImageElement;
        target.style.display = 'none';
        const parent = target.parentElement;
        if (parent) {
          const fallback = document.createElement('div');
          fallback.className = 'avatar avatar-initials';
          fallback.style.width = `${size}px`;
          fallback.style.height = `${size}px`;
          fallback.style.backgroundColor = generateColor();
          fallback.textContent = getInitials();
          parent.appendChild(fallback);
        }
      }}
    />
  );
};

export default Avatar;
