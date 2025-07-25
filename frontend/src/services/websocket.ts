import { WebSocketMessage } from '../types';

class WebSocketService {
  private socket: WebSocket | null = null;
  private listeners: Map<string, ((message: WebSocketMessage) => void)[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  connect(gameId?: number): void {
    const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
    const token = localStorage.getItem('token');

    // Build WebSocket URL with query parameters
    const url = new URL('/ws', WS_URL.replace('ws://', 'http://').replace('wss://', 'https://'));
    if (token) {
      url.searchParams.append('token', token);
    }
    if (gameId) {
      url.searchParams.append('gameId', gameId.toString());
    }

    // Convert back to WebSocket URL
    const wsUrl = url.toString().replace('http://', 'ws://').replace('https://', 'wss://');

    this.socket = new WebSocket(wsUrl);

    this.socket.onopen = () => {
      console.log('Connected to WebSocket');
      this.reconnectAttempts = 0;
      if (gameId) {
        this.joinGame(gameId);
      }
    };

    this.socket.onclose = (event) => {
      console.log('Disconnected from WebSocket', event);
      this.handleReconnect(gameId);
    };

    this.socket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private handleReconnect(gameId?: number): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      setTimeout(() => {
        this.connect(gameId);
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.listeners.clear();
  }

  joinGame(gameId: number): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.sendMessage({
        type: 'join_game',
        data: gameId,
      });
    }
  }

  sendChatMessage(gameId: number, message: string): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.sendMessage({
        type: 'chat',
        game_id: gameId,
        message,
      });
    }
  }

  sendMove(gameId: number, x: number, y: number): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.sendMessage({
        type: 'move',
        game_id: gameId,
        data: { x, y },
      });
    }
  }

  sendShipPlacement(gameId: number, ships: unknown[]): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.sendMessage({
        type: 'ship_placement',
        game_id: gameId,
        data: ships,
      });
    }
  }

  private sendMessage(message: unknown): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    }
  }

  on(event: string, callback: (message: WebSocketMessage) => void): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      eventListeners.push(callback);
    }
  }

  off(event: string, callback: (message: WebSocketMessage) => void): void {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      const index = eventListeners.indexOf(callback);
      if (index > -1) {
        eventListeners.splice(index, 1);
      }
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    const listeners = this.listeners.get(message.type);
    if (listeners) {
      listeners.forEach(callback => callback(message));
    }

    // Also emit to generic 'message' listeners
    const messageListeners = this.listeners.get('message');
    if (messageListeners) {
      messageListeners.forEach(callback => callback(message));
    }
  }

  isConnected(): boolean {
    return this.socket?.readyState === WebSocket.OPEN;
  }
}

export const websocketService = new WebSocketService();
export default websocketService;
