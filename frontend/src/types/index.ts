export interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface Game {
  id: number;
  player1_id: number;
  player2_id?: number;
  status: 'waiting' | 'active' | 'finished';
  current_turn?: number;
  winner_id?: number;
  created_at: string;
  updated_at: string;
}

export interface Ship {
  id?: number;
  game_id?: number;
  player_id?: number;
  type: 'carrier' | 'battleship' | 'cruiser' | 'submarine' | 'destroyer';
  size: number;
  start_x: number;
  start_y: number;
  end_x: number;
  end_y: number;
  is_vertical: boolean;
  is_sunk?: boolean;
}

export interface Move {
  id: number;
  game_id: number;
  player_id: number;
  x: number;
  y: number;
  is_hit: boolean;
  ship_id?: number;
  created_at: string;
}

export interface ChatMessage {
  id: number;
  game_id: number;
  player_id: number;
  message: string;
  created_at: string;
}

export interface Score {
  id: number;
  player_id: number;
  wins: number;
  losses: number;
  hits: number;
  misses: number;
  points: number;
}

export interface LeaderboardEntry extends Score {
  username: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface GameState {
  game: Game | null;
  playerShips: Ship[];
  opponentShips: Ship[];
  moves: Move[];
  chatMessages: ChatMessage[];
  isMyTurn: boolean;
  gameBoard: CellState[][];
  opponentBoard: CellState[][];
}

export type CellState = 'empty' | 'ship' | 'hit' | 'miss' | 'sunk';

export interface Position {
  x: number;
  y: number;
}

export interface WebSocketMessage {
  type: 'chat' | 'move' | 'game_update' | 'join_game' | 'ship_placement';
  game_id?: number;
  user_id?: number;
  data?: any;
  message?: string;
}
