import axios from 'axios';
import { AuthResponse, User, Game, Ship, Move, ChatMessage, LeaderboardEntry, Score } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  register: async (username: string, email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post('/auth/register', { username, email, password });
    return response.data;
  },

  login: async (username: string, password: string): Promise<AuthResponse> => {
    const response = await api.post('/auth/login', { username, password });
    return response.data;
  },
};

export const userAPI = {
  getProfile: async (): Promise<User> => {
    const response = await api.get('/user/profile');
    return response.data;
  },

  getStats: async (): Promise<Score> => {
    const response = await api.get('/user/stats');
    return response.data;
  },
};

export const gameAPI = {
  createGame: async (): Promise<Game> => {
    const response = await api.post('/games');
    return response.data;
  },

  joinGame: async (gameId: number): Promise<Game> => {
    const response = await api.post(`/games/${gameId}/join`);
    return response.data;
  },

  getGames: async (): Promise<Game[]> => {
    const response = await api.get('/games');
    return response.data;
  },

  getGame: async (gameId: number): Promise<Game> => {
    const response = await api.get(`/games/${gameId}`);
    return response.data;
  },

  placeShips: async (gameId: number, ships: Ship[]): Promise<void> => {
    await api.post(`/games/${gameId}/ships`, ships);
  },

  makeMove: async (gameId: number, x: number, y: number): Promise<Move> => {
    const response = await api.post(`/games/${gameId}/moves`, { x, y });
    return response.data;
  },

  getGameMoves: async (gameId: number): Promise<Move[]> => {
    const response = await api.get(`/games/${gameId}/moves`);
    return response.data;
  },
};

export const chatAPI = {
  sendMessage: async (gameId: number, message: string): Promise<ChatMessage> => {
    const response = await api.post(`/games/${gameId}/chat`, { message });
    return response.data;
  },

  getMessages: async (gameId: number): Promise<ChatMessage[]> => {
    const response = await api.get(`/games/${gameId}/chat`);
    return response.data;
  },
};

export const leaderboardAPI = {
  getLeaderboard: async (): Promise<LeaderboardEntry[]> => {
    const response = await api.get('/leaderboard');
    return response.data;
  },
};

export default api;
