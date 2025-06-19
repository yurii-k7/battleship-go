import React, { useState, useEffect } from 'react';
import { leaderboardAPI } from '../services/api';
import { LeaderboardEntry } from '../types';

const Leaderboard: React.FC = () => {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadLeaderboard();
  }, []);

  const loadLeaderboard = async () => {
    try {
      const data = await leaderboardAPI.getLeaderboard();
      setLeaderboard(data);
    } catch (err: any) {
      setError('Failed to load leaderboard');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const calculateWinRate = (wins: number, losses: number): string => {
    const total = wins + losses;
    if (total === 0) return '0%';
    return `${Math.round((wins / total) * 100)}%`;
  };

  const calculateAccuracy = (hits: number, misses: number): string => {
    const total = hits + misses;
    if (total === 0) return '0%';
    return `${Math.round((hits / total) * 100)}%`;
  };

  if (loading) {
    return <div className="loading">Loading leaderboard...</div>;
  }

  return (
    <div className="leaderboard">
      <h1>🏆 Leaderboard</h1>
      
      {error && <div className="error-message">{error}</div>}
      
      <div className="leaderboard-table">
        <table>
          <thead>
            <tr>
              <th>Rank</th>
              <th>Player</th>
              <th>Points</th>
              <th>Wins</th>
              <th>Losses</th>
              <th>Win Rate</th>
              <th>Hits</th>
              <th>Misses</th>
              <th>Accuracy</th>
            </tr>
          </thead>
          <tbody>
            {leaderboard.length === 0 ? (
              <tr>
                <td colSpan={9} style={{ textAlign: 'center', padding: '2rem' }}>
                  No players on the leaderboard yet.
                </td>
              </tr>
            ) : (
              leaderboard.map((entry, index) => (
                <tr key={entry.id}>
                  <td>
                    <strong>#{index + 1}</strong>
                    {index === 0 && ' 🥇'}
                    {index === 1 && ' 🥈'}
                    {index === 2 && ' 🥉'}
                  </td>
                  <td>
                    <strong>{entry.username}</strong>
                  </td>
                  <td>
                    <strong>{entry.points}</strong>
                  </td>
                  <td>{entry.wins}</td>
                  <td>{entry.losses}</td>
                  <td>{calculateWinRate(entry.wins, entry.losses)}</td>
                  <td>{entry.hits}</td>
                  <td>{entry.misses}</td>
                  <td>{calculateAccuracy(entry.hits, entry.misses)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      
      <div style={{ marginTop: '2rem', padding: '1rem', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
        <h3>Scoring System</h3>
        <ul>
          <li><strong>Win:</strong> +100 points</li>
          <li><strong>Hit:</strong> +10 points</li>
          <li><strong>Sinking a ship:</strong> +20 bonus points</li>
          <li><strong>Perfect game (no misses):</strong> +50 bonus points</li>
        </ul>
      </div>
    </div>
  );
};

export default Leaderboard;
