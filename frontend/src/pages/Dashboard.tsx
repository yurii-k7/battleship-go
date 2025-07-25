import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { gameAPI, userAPI } from '../services/api';
import { websocketService } from '../services/websocket';
import { Game, Score } from '../types';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [games, setGames] = useState<Game[]>([]);
  const [stats, setStats] = useState<Score | null>(null);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadDashboardData();
    
    // Connect to WebSocket for real-time updates
    websocketService.connect();
    
    // Listen for new game creation events
    websocketService.on('new_game_created', (message) => {
      console.log('New game created:', message);
      loadDashboardData(); // Refresh the games list
    });

    // Listen for game updates (like when someone joins)
    websocketService.on('game_update', (message) => {
      console.log('Game updated:', message);
      loadDashboardData(); // Refresh the games list
    });

    return () => {
      websocketService.disconnect();
    };
  }, []);

  const loadDashboardData = async () => {
    try {
      const [gamesData, statsData] = await Promise.all([
        gameAPI.getGames(),
        userAPI.getStats()
      ]);
      setGames(gamesData || []);
      setStats(statsData);
    } catch (err: unknown) {
      setError('Failed to load dashboard data');
      console.error(err);
      // Set empty arrays on error to prevent null reference errors
      setGames([]);
    } finally {
      setLoading(false);
    }
  };

  const createNewGame = async () => {
    setCreating(true);
    setError('');

    try {
      const newGame = await gameAPI.createGame();
      setGames([newGame, ...(games || [])]);
    } catch (err: unknown) {
      setError('Failed to create game');
      console.error(err);
    } finally {
      setCreating(false);
    }
  };

  const joinGame = async (gameId: number) => {
    try {
      await gameAPI.joinGame(gameId);
      loadDashboardData(); // Refresh the games list
    } catch (err: unknown) {
      setError('Failed to join game');
      console.error(err);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getGameStatusClass = (status: string) => {
    return `game-status ${status}`;
  };

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  return (
    <div className="dashboard">
      <h1>Welcome back, {user?.username}!</h1>
      
      {error && <div className="error-message">{error}</div>}
      
      {/* Player Stats */}
      {stats && (
        <div className="game-list">
          <h2>Your Stats</h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: '1rem', textAlign: 'center' }}>
            <div>
              <strong>{stats.wins}</strong>
              <div>Wins</div>
            </div>
            <div>
              <strong>{stats.losses}</strong>
              <div>Losses</div>
            </div>
            <div>
              <strong>{stats.hits}</strong>
              <div>Hits</div>
            </div>
            <div>
              <strong>{stats.misses}</strong>
              <div>Misses</div>
            </div>
            <div>
              <strong>{stats.points}</strong>
              <div>Points</div>
            </div>
          </div>
        </div>
      )}
      
      {/* Game Actions */}
      <div className="game-list">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h2>Games</h2>
          <button onClick={createNewGame} disabled={creating} className="btn-primary">
            {creating ? 'Creating...' : 'Create New Game'}
          </button>
        </div>
        
        {!games || games.length === 0 ? (
          <p>No games yet. Create your first game!</p>
        ) : (
          <div>
            {games.map((game) => (
              <div key={game.id} className="game-item">
                <div>
                  <strong>Game #{game.id}</strong>
                  <div style={{ fontSize: '0.9em', color: '#666' }}>
                    Created: {formatDate(game.created_at)}
                  </div>
                </div>
                
                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                  <span className={getGameStatusClass(game.status)}>
                    {game.status.toUpperCase()}
                  </span>
                  
                  {game.status === 'waiting' && game.player1_id !== user?.id && (
                    <button onClick={() => joinGame(game.id)}>
                      Join Game
                    </button>
                  )}
                  
                  {(game.status === 'active' || 
                    (game.status === 'waiting' && game.player1_id === user?.id)) && (
                    <Link to={`/game/${game.id}`}>
                      <button>Enter Game</button>
                    </Link>
                  )}
                  
                  {game.status === 'finished' && (
                    <Link to={`/game/${game.id}`}>
                      <button>View Game</button>
                    </Link>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;
