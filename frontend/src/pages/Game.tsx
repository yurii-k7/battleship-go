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
  const [gamePhase, setGamePhase] = useState<'waiting' | 'placing' | 'waiting_for_opponent' | 'playing' | 'finished'>('waiting');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [errorTimeout, setErrorTimeout] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (gameId) {
      loadGameData();
      connectWebSocket();
    }

    return () => {
      websocketService.disconnect();
      // Clear error timeout on unmount
      if (errorTimeout) {
        clearTimeout(errorTimeout);
      }
    };
  }, [gameId, errorTimeout]);

  // Update boards when ships or moves change
  useEffect(() => {
    if (ships.length > 0 || moves.length > 0) {
      updateBoards();
    }
  }, [ships, moves]);

  const setErrorWithTimeout = (message: string) => {
    // Clear existing timeout if any
    if (errorTimeout) {
      clearTimeout(errorTimeout);
    }
    
    setError(message);
    
    // Set new timeout to clear error after 3 seconds
    const timeout = setTimeout(() => {
      setError('');
      setErrorTimeout(null);
    }, 3000);
    
    setErrorTimeout(timeout);
  };

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
      setErrorWithTimeout('Failed to load game data');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const checkShipPlacementStatus = async (gameId: number) => {
    // Use the same logic as checkIfReadyToPlay for consistency
    await checkIfReadyToPlay();
  };

  const refreshGameState = async () => {
    try {
      const [gameData, movesData] = await Promise.all([
        gameAPI.getGame(parseInt(gameId!)),
        gameAPI.getGameMoves(parseInt(gameId!))
      ]);
      
      setGame(gameData);
      setMoves(movesData);
    } catch (err) {
      console.error('Failed to refresh game state:', err);
    }
  };

  const connectWebSocket = () => {
    websocketService.connect(parseInt(gameId!));

    websocketService.on('game_update', async (message: any) => {
      console.log('Game update received:', message);
      
      // Refresh game state first
      await refreshGameState();
      
      // Then check for phase transitions based on current phase
      const currentPhase = gamePhase;
      if (currentPhase === 'waiting' || currentPhase === 'waiting_for_opponent') {
        await checkIfReadyToPlay();
      }
    });

    websocketService.on('ship_placement_update', async (message: any) => {
      console.log('Ship placement update received:', message);
      
      // Check if both players have now placed ships
      await checkIfReadyToPlay();
    });

    websocketService.on('move', (message: any) => {
      setMoves(prev => [...prev, message.data]);
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

    console.log('Updating boards with', moves.length, 'moves for user', user?.id);

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
      console.log('Processing move:', move);
      if (move.player_id === user?.id) {
        // My move on opponent's board
        console.log('My move on opponent board:', move.x, move.y, move.is_hit ? 'hit' : 'miss');
        newOpponentBoard[move.y][move.x] = move.is_hit ? 'hit' : 'miss';
      } else {
        // Opponent's move on my board
        console.log('Opponent move on my board:', move.x, move.y, move.is_hit ? 'hit' : 'miss');
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
      // Check if both players have placed ships to determine game phase
      await checkIfReadyToPlay();
    } catch (err: any) {
      // If ships are already placed, just check if we can start playing
      if (err.response?.data?.error === 'ships already placed') {
        await checkIfReadyToPlay();
      } else {
        setErrorWithTimeout('Failed to place ships');
        console.error(err);
      }
    }
  };

  const checkIfReadyToPlay = async () => {
    try {
      const readyStatus = await gameAPI.checkGameReady(parseInt(gameId!));
      
      if (readyStatus.ready) {
        // Both players have placed ships, game is ready
        const myShips = await gameAPI.getShips(parseInt(gameId!));
        setShips(myShips);
        setGamePhase('playing');
      } else {
        // Check if current player has placed ships
        const myShips = await gameAPI.getShips(parseInt(gameId!));
        if (myShips.length === 5) {
          // Current player has placed ships, waiting for opponent
          setShips(myShips);
          setGamePhase('waiting_for_opponent');
        } else {
          // Current player still needs to place ships
          setGamePhase('placing');
        }
      }
    } catch (err) {
      console.error('Failed to check game readiness:', err);
      setGamePhase('placing');
    }
  };

  const handleCellClick = async (x: number, y: number) => {
    console.log('Cell clicked:', x, y, 'types:', typeof x, typeof y, 'gamePhase:', gamePhase, 'isMyTurn:', isMyTurn());
    
    // Validate coordinates
    if (typeof x !== 'number' || typeof y !== 'number' || x < 0 || x > 9 || y < 0 || y > 9) {
      console.error('Invalid coordinates:', x, y);
      setErrorWithTimeout('Invalid move coordinates');
      return;
    }
    
    if (gamePhase !== 'playing' || !isMyTurn()) {
      console.log('Move blocked - not playing or not my turn');
      return;
    }

    try {
      console.log('Making move with coordinates:', x, y);
      const move = await gameAPI.makeMove(parseInt(gameId!), x, y);
      console.log('Move successful:', move);
      setMoves(prev => [...prev, move]);
      websocketService.sendMove(parseInt(gameId!), x, y);
    } catch (err: any) {
      console.error('Move failed:', err);
      setErrorWithTimeout(`Failed to make move: ${err.response?.data?.error || err.message}`);
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

        {gamePhase === 'waiting_for_opponent' && (
          <div>
            <h2>Waiting for opponent to place ships...</h2>
            <p>You have placed all your ships. Please wait for your opponent to finish placing their ships.</p>
          </div>
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
