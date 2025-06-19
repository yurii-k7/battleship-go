import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { gameAPI, chatAPI } from '../services/api';
import { websocketService } from '../services/websocket';
import { Game as GameType, Ship, Move, ChatMessage, CellState } from '../types';
import GameBoard from '../components/GameBoard';
import ShipPlacement from '../components/ShipPlacement';
import Chat from '../components/Chat';

const Game: React.FC = () => {
  const { gameId } = useParams<{ gameId: string }>();
  const { user } = useAuth();
  const [game, setGame] = useState<GameType | null>(null);
  const [ships, setShips] = useState<Ship[]>([]);
  const [moves, setMoves] = useState<Move[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [playerBoard, setPlayerBoard] = useState<CellState[][]>([]);
  const [opponentBoard, setOpponentBoard] = useState<CellState[][]>([]);
  const [gamePhase, setGamePhase] = useState<'waiting' | 'placing' | 'playing' | 'finished'>('waiting');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (gameId) {
      loadGameData();
      connectWebSocket();
    }

    return () => {
      websocketService.disconnect();
    };
  }, [gameId]);

  const loadGameData = async () => {
    try {
      const [gameData, movesData, chatData] = await Promise.all([
        gameAPI.getGame(parseInt(gameId!)),
        gameAPI.getGameMoves(parseInt(gameId!)),
        chatAPI.getMessages(parseInt(gameId!))
      ]);

      setGame(gameData);
      setMoves(movesData);
      setChatMessages(chatData);
      
      // Initialize boards
      initializeBoards();
      
      // Determine game phase
      if (gameData.status === 'waiting') {
        setGamePhase('waiting');
      } else if (gameData.status === 'active') {
        // Check if current user has placed ships
        await checkShipPlacementStatus(gameData.id);
      } else {
        setGamePhase('finished');
      }
      
    } catch (err: any) {
      setError('Failed to load game data');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const checkShipPlacementStatus = async (gameId: number) => {
    // For now, assume if game is active, we need to check if ships are placed
    // In a real implementation, you'd call an API to check ship placement status
    // For simplicity, we'll assume ships need to be placed when game becomes active
    setGamePhase('placing');
  };

  const connectWebSocket = () => {
    websocketService.connect(parseInt(gameId!));

    websocketService.on('game_update', (message: any) => {
      loadGameData();
    });

    websocketService.on('move', (message: any) => {
      setMoves(prev => [...prev, message.data]);
      updateBoards();
    });

    websocketService.on('chat', (message: any) => {
      setChatMessages(prev => [...prev, message.data]);
    });
  };

  const initializeBoards = () => {
    const emptyBoard: CellState[][] = Array(10).fill(null).map(() => 
      Array(10).fill('empty')
    );
    setPlayerBoard([...emptyBoard]);
    setOpponentBoard([...emptyBoard]);
  };

  const updateBoards = () => {
    // Update boards based on ships and moves
    // This is a simplified implementation
    const newPlayerBoard: CellState[][] = Array(10).fill(null).map(() => 
      Array(10).fill('empty')
    );
    const newOpponentBoard: CellState[][] = Array(10).fill(null).map(() => 
      Array(10).fill('empty')
    );

    // Place ships on player board
    ships.forEach(ship => {
      if (ship.is_vertical) {
        for (let y = ship.start_y; y <= ship.end_y; y++) {
          newPlayerBoard[y][ship.start_x] = 'ship';
        }
      } else {
        for (let x = ship.start_x; x <= ship.end_x; x++) {
          newPlayerBoard[ship.start_y][x] = 'ship';
        }
      }
    });

    // Apply moves to boards
    moves.forEach(move => {
      if (move.player_id === user?.id) {
        // My move on opponent's board
        newOpponentBoard[move.y][move.x] = move.is_hit ? 'hit' : 'miss';
      } else {
        // Opponent's move on my board
        newPlayerBoard[move.y][move.x] = move.is_hit ? 'hit' : 'miss';
      }
    });

    setPlayerBoard(newPlayerBoard);
    setOpponentBoard(newOpponentBoard);
  };

  const handleShipsPlaced = async (placedShips: Ship[]) => {
    try {
      await gameAPI.placeShips(parseInt(gameId!), placedShips);
      setShips(placedShips);
      setGamePhase('playing');
      updateBoards();
    } catch (err: any) {
      setError('Failed to place ships');
      console.error(err);
    }
  };

  const handleCellClick = async (x: number, y: number) => {
    if (gamePhase !== 'playing' || !isMyTurn()) {
      return;
    }

    try {
      const move = await gameAPI.makeMove(parseInt(gameId!), x, y);
      setMoves(prev => [...prev, move]);
      websocketService.sendMove(parseInt(gameId!), x, y);
      updateBoards();
    } catch (err: any) {
      setError('Failed to make move');
      console.error(err);
    }
  };

  const handleChatMessage = async (message: string) => {
    try {
      const chatMessage = await chatAPI.sendMessage(parseInt(gameId!), message);
      setChatMessages(prev => [...prev, chatMessage]);
      websocketService.sendChatMessage(parseInt(gameId!), message);
    } catch (err: any) {
      console.error('Failed to send chat message:', err);
    }
  };

  const isMyTurn = (): boolean => {
    if (!game || !user) return false;

    // During waiting phase, no one has a turn
    if (game.status === 'waiting') return false;

    // During active phase, check current_turn
    if (game.status === 'active') {
      return game.current_turn === user.id;
    }

    // Game is finished, no one has a turn
    return false;
  };

  const getTurnDisplayText = (): string => {
    if (!game || !user) return 'Loading...';

    switch (game.status) {
      case 'waiting':
        return 'Waiting for opponent';
      case 'active':
        if (game.current_turn === user.id) {
          return 'Your turn';
        } else if (game.current_turn) {
          return "Opponent's turn";
        } else {
          return 'Game starting...';
        }
      case 'finished':
        if (game.winner_id === user.id) {
          return 'You won!';
        } else if (game.winner_id) {
          return 'You lost';
        } else {
          return 'Game ended';
        }
      default:
        return 'Unknown status';
    }
  };

  if (loading) {
    return <div className="loading">Loading game...</div>;
  }

  if (!game) {
    return <div className="error-message">Game not found</div>;
  }

  return (
    <div className="game-container">
      <div className="game-board-container">
        <h1>Game #{game.id}</h1>
        
        {error && <div className="error-message">{error}</div>}
        
        <div style={{ marginBottom: '1rem' }}>
          <strong>Status:</strong> {game.status} |
          <strong> Turn:</strong> {getTurnDisplayText()}
        </div>

        {gamePhase === 'waiting' && (
          <div>
            <h2>Waiting for opponent to join...</h2>
            <p>Share this game ID with a friend: <strong>{game.id}</strong></p>
          </div>
        )}

        {gamePhase === 'placing' && (
          <ShipPlacement onShipsPlaced={handleShipsPlaced} />
        )}

        {gamePhase === 'playing' && (
          <div style={{ display: 'flex', gap: '2rem' }}>
            <div>
              <h3>Your Board</h3>
              <GameBoard 
                board={playerBoard}
                onCellClick={() => {}} // Player can't click their own board
                disabled={true}
              />
            </div>
            <div>
              <h3>Opponent's Board</h3>
              <GameBoard 
                board={opponentBoard}
                onCellClick={handleCellClick}
                disabled={!isMyTurn()}
              />
            </div>
          </div>
        )}

        {gamePhase === 'finished' && (
          <div>
            <h2>Game Finished!</h2>
            <p>
              Winner: {game.winner_id === user?.id ? 'You!' : 'Opponent'}
            </p>
          </div>
        )}
      </div>

      <Chat 
        messages={chatMessages}
        onSendMessage={handleChatMessage}
        currentUserId={user?.id || 0}
      />
    </div>
  );
};

export default Game;
