import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import FileUpload from '../components/FileUpload';
import './LoginPage.css';

const LoginPage: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);

  // Login form state
  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');

  // Registration form state
  const [regEmail, setRegEmail] = useState('');
  const [regUsername, setRegUsername] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regDisplayName, setRegDisplayName] = useState('');

  const { login, register } = useAuth();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await login(loginEmail, loginPassword);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      let avatarUrl: string | undefined;
      if (avatarFile) {
        // Convert file to base64 data URL
        avatarUrl = await new Promise<string>((resolve, reject) => {
          const reader = new FileReader();
          reader.onloadend = () => resolve(reader.result as string);
          reader.onerror = reject;
          reader.readAsDataURL(avatarFile);
        });
      }

      await register(regEmail, regUsername, regPassword, regDisplayName, avatarUrl);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleFileSelect = (file: File | null) => {
    setAvatarFile(file);
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-header">
          <h1>{isLogin ? 'Login' : 'Register'}</h1>
          <p className="login-subtitle">
            {isLogin
              ? 'Welcome back! Please sign in to your account.'
              : 'Create a new account to get started.'}
          </p>
        </div>

        <div className="login-toggle">
          <button
            type="button"
            className={`toggle-button ${isLogin ? 'active' : ''}`}
            onClick={() => {
              setIsLogin(true);
              setError(null);
            }}
          >
            Login
          </button>
          <button
            type="button"
            className={`toggle-button ${!isLogin ? 'active' : ''}`}
            onClick={() => {
              setIsLogin(false);
              setError(null);
            }}
          >
            Register
          </button>
        </div>

        {error && <div className="error-message">{error}</div>}

        {isLogin ? (
          <form onSubmit={handleLogin} className="login-form">
            <div className="form-group">
              <label htmlFor="login-email">Email</label>
              <input
                id="login-email"
                type="email"
                value={loginEmail}
                onChange={(e) => setLoginEmail(e.target.value)}
                required
                placeholder="Enter your email"
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="login-password">Password</label>
              <input
                id="login-password"
                type="password"
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
                required
                placeholder="Enter your password"
                disabled={loading}
              />
            </div>

            <button type="submit" className="submit-button" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleRegister} className="login-form">
            <div className="form-group">
              <label htmlFor="reg-email">Email</label>
              <input
                id="reg-email"
                type="email"
                value={regEmail}
                onChange={(e) => setRegEmail(e.target.value)}
                required
                placeholder="Enter your email"
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="reg-username">Username</label>
              <input
                id="reg-username"
                type="text"
                value={regUsername}
                onChange={(e) => setRegUsername(e.target.value)}
                required
                minLength={3}
                maxLength={50}
                placeholder="Choose a username"
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="reg-display-name">Display Name</label>
              <input
                id="reg-display-name"
                type="text"
                value={regDisplayName}
                onChange={(e) => setRegDisplayName(e.target.value)}
                required
                placeholder="Enter your display name"
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label htmlFor="reg-password">Password</label>
              <input
                id="reg-password"
                type="password"
                value={regPassword}
                onChange={(e) => setRegPassword(e.target.value)}
                required
                minLength={6}
                placeholder="Enter your password (min 6 characters)"
                disabled={loading}
              />
            </div>

            <div className="form-group">
              <label>Avatar (Optional)</label>
              <FileUpload
                onFileSelect={handleFileSelect}
                currentFile={avatarFile}
                maxSizeMB={5}
              />
            </div>

            <button type="submit" className="submit-button" disabled={loading}>
              {loading ? 'Creating account...' : 'Create Account'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default LoginPage;
