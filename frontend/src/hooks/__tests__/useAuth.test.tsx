import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { AuthProvider, useAuth } from '../useAuth';
import * as api from '../../services/api';

// Mock the API
jest.mock('../../services/api');
const mockedApi = api as jest.Mocked<typeof api>;

// Test component that uses the auth hook
const TestComponent: React.FC = () => {
  const { user, loading, login, register, logout } = useAuth();

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      {user ? (
        <div>
          <span>Welcome {user.username}</span>
          <button onClick={logout}>Logout</button>
        </div>
      ) : (
        <div>
          <span>Not logged in</span>
          <button onClick={() => login('testuser', 'password')}>Login</button>
          <button onClick={() => register('testuser', 'test@example.com', 'password')}>Register</button>
        </div>
      )}
    </div>
  );
};

describe('useAuth', () => {
  beforeEach(() => {
    // Clear localStorage
    localStorage.clear();
    // Reset mocks
    jest.clearAllMocks();
  });

  it('shows loading state initially', () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('shows not logged in state when no user', async () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Not logged in')).toBeInTheDocument();
    });
  });

  it('restores user from localStorage', async () => {
    const mockUser = {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2023-01-01T00:00:00Z'
    };

    localStorage.setItem('token', 'mock-token');
    localStorage.setItem('user', JSON.stringify(mockUser));

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Welcome testuser')).toBeInTheDocument();
    });
  });

  it('handles login successfully', async () => {
    const mockUser = {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2023-01-01T00:00:00Z'
    };

    mockedApi.authAPI.login.mockResolvedValue({
      user: mockUser,
      token: 'mock-token'
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Not logged in')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Login'));

    await waitFor(() => {
      expect(screen.getByText('Welcome testuser')).toBeInTheDocument();
    });

    expect(localStorage.getItem('token')).toBe('mock-token');
    expect(localStorage.getItem('user')).toBe(JSON.stringify(mockUser));
  });

  it('handles register successfully', async () => {
    const mockUser = {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2023-01-01T00:00:00Z'
    };

    mockedApi.authAPI.register.mockResolvedValue({
      user: mockUser,
      token: 'mock-token'
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Not logged in')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Register'));

    await waitFor(() => {
      expect(screen.getByText('Welcome testuser')).toBeInTheDocument();
    });

    expect(localStorage.getItem('token')).toBe('mock-token');
    expect(localStorage.getItem('user')).toBe(JSON.stringify(mockUser));
  });

  it('handles logout', async () => {
    const mockUser = {
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      created_at: '2023-01-01T00:00:00Z',
      updated_at: '2023-01-01T00:00:00Z'
    };

    localStorage.setItem('token', 'mock-token');
    localStorage.setItem('user', JSON.stringify(mockUser));

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Welcome testuser')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Logout'));

    await waitFor(() => {
      expect(screen.getByText('Not logged in')).toBeInTheDocument();
    });

    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('user')).toBeNull();
  });
});
