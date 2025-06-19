import React from 'react';
import { CellState } from '../types';

interface GameBoardProps {
  board: CellState[][];
  onCellClick: (x: number, y: number) => void;
  disabled?: boolean;
}

const GameBoard: React.FC<GameBoardProps> = ({ board, onCellClick, disabled = false }) => {
  const getCellClass = (cellState: CellState): string => {
    return `game-cell ${cellState}`;
  };

  return (
    <div className="game-board">
      {board.map((row, y) =>
        row.map((cell, x) => (
          <button
            key={`${x}-${y}`}
            className={getCellClass(cell)}
            onClick={() => onCellClick(x, y)}
            disabled={disabled || cell === 'hit' || cell === 'miss'}
            title={`${x}, ${y}`}
          />
        ))
      )}
    </div>
  );
};

export default GameBoard;
