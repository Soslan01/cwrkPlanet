import React from 'react';
import './WelcomePage.css';

const FEATURE_CARDS = [
  {
    icon: '👤',
    title: 'Your profile',
    description: 'Manage your account, avatar, and display name in one place.',
  },
  {
    icon: '🔐',
    title: 'Secure auth',
    description: 'Sign in or register with email. Your data stays safe.',
  },
  {
    icon: '🌓',
    title: 'Light & dark',
    description: 'Switch themes anytime for comfortable viewing day or night.',
  },
  {
    icon: '📱',
    title: 'Mobile-friendly',
    description: 'Use the app on any device with a clean, simple interface.',
  },
];

const WelcomePage: React.FC = () => {
  return (
    <div className="welcome-page">
      <section className="welcome-hero">
        <h1 className="welcome-title">
          Working towards your goals is hard.<br />
          <span className="welcome-title-accent">Not reaching them is harder.</span>
        </h1>
        <p className="welcome-subtitle">
          Get started with a single account. Sign in to access your profile and keep everything in sync.
        </p>
      </section>

      <section className="welcome-cards">
        <h2 className="welcome-cards-heading">What you can do here</h2>
        <div className="welcome-cards-grid">
          {FEATURE_CARDS.map((card, index) => (
            <article key={index} className="welcome-card">
              <span className="welcome-card-icon" aria-hidden>{card.icon}</span>
              <h3 className="welcome-card-title">{card.title}</h3>
              <p className="welcome-card-desc">{card.description}</p>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
};

export default WelcomePage;
