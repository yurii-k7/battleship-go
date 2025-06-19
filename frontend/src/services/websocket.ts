import { io, Socket } from 'socket.io-client';
import { WebSocketMessage } from '../types';

class WebSocketService {
  private socket: Socket | null = null;
  private listeners: Map<string, Function[]> = new Map();

  connect(gameId?: number): void {
    const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
    const token = localStorage.getItem('token');
    
    this.socket = io(WS_URL, {
      auth: {
        token,
      },
      query: {
        gameId: gameId?.toString(),
      },
    });

    this.socket.on('connect', () => {
      console.log('Connected to WebSocket');
      if (gameId) {
        this.joinGame(gameId);
      }
    });

    this.socket.on('disconnect', () => {
      console.log('Disconnected from WebSocket');
    });

    this.socket.on('message', (message: WebSocketMessage) => {
      this.handleMessage(message);
    });

    this.socket.on('error', (error: any) => {
      console.error('WebSocket error:', error);
    });
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
    this.listeners.clear();
  }

  joinGame(gameId: number): void {
    if (this.socket) {
      this.socket.emit('message', {
        type: 'join_game',
        data: gameId,
      });
    }
  }

  sendChatMessage(gameId: number, message: string): void {
    if (this.socket) {
      this.socket.emit('message', {
        type: 'chat',
        game_id: gameId,
        message,
      });
    }
  }

  sendMove(gameId: number, x: number, y: number): void {
    if (this.socket) {
      this.socket.emit('message', {
        type: 'move',
        game_id: gameId,
        data: { x, y },
      });
    }
  }

  sendShipPlacement(gameId: number, ships: any[]): void {
    if (this.socket) {
      this.socket.emit('message', {
        type: 'ship_placement',
        game_id: gameId,
        data: ships,
      });
    }
  }

  on(event: string, callback: Function): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)!.push(callback);
  }

  off(event: string, callback: Function): void {
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
    return this.socket?.connected || false;
  }
}

export const websocketService = new WebSocketService();
export default websocketService;
