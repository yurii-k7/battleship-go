import React, { useState } from 'react';
import { Ship } from '../types';

interface ShipPlacementProps {
  onShipsPlaced: (ships: Ship[]) => void;
}

const SHIP_TYPES = [
  { type: 'carrier', size: 5, name: 'Carrier' },
  { type: 'battleship', size: 4, name: 'Battleship' },
  { type: 'cruiser', size: 3, name: 'Cruiser' },
  { type: 'submarine', size: 3, name: 'Submarine' },
  { type: 'destroyer', size: 2, name: 'Destroyer' },
] as const;

const ShipPlacement: React.FC<ShipPlacementProps> = ({ onShipsPlaced }) => {
  const [board, setBoard] = useState<string[][]>(
    Array(10).fill(null).map(() => Array(10).fill('empty'))
  );
  const [ships, setShips] = useState<Ship[]>([]);
  const [currentShipIndex, setCurrentShipIndex] = useState(0);
  const [isVertical, setIsVertical] = useState(false);
  const [dragStart, setDragStart] = useState<{ x: number; y: number } | null>(null);

  const currentShip = SHIP_TYPES[currentShipIndex];

  const canPlaceShip = (x: number, y: number, size: number, vertical: boolean): boolean => {
    // Check bounds
    if (vertical) {
      if (y + size > 10) return false;
    } else {
      if (x + size > 10) return false;
    }

    // Check for overlaps
    for (let i = 0; i < size; i++) {
      const checkX = vertical ? x : x + i;
      const checkY = vertical ? y + i : y;
      
      if (board[checkY][checkX] !== 'empty') {
        return false;
      }
    }

    return true;
  };

  const placeShip = (x: number, y: number) => {
    if (!currentShip || !canPlaceShip(x, y, currentShip.size, isVertical)) {
      return;
    }

    const newBoard = board.map(row => [...row]);
    const newShip: Ship = {
      type: currentShip.type,
      size: currentShip.size,
      start_x: x,
      start_y: y,
      end_x: isVertical ? x : x + currentShip.size - 1,
      end_y: isVertical ? y + currentShip.size - 1 : y,
      is_vertical: isVertical,
    };

    // Place ship on board
    for (let i = 0; i < currentShip.size; i++) {
      const placeX = isVertical ? x : x + i;
      const placeY = isVertical ? y + i : y;
      newBoard[placeY][placeX] = 'ship';
    }

    setBoard(newBoard);
    setShips([...ships, newShip]);
    setCurrentShipIndex(currentShipIndex + 1);
  };

  const handleCellClick = (x: number, y: number) => {
    if (currentShipIndex < SHIP_TYPES.length) {
      placeShip(x, y);
    }
  };

  const handleCellMouseEnter = (x: number, y: number) => {
    if (currentShipIndex >= SHIP_TYPES.length) return;

    const newBoard = board.map(row => [...row]);
    
    // Clear previous preview
    for (let row = 0; row < 10; row++) {
      for (let col = 0; col < 10; col++) {
        if (newBoard[row][col] === 'preview') {
          newBoard[row][col] = 'empty';
        }
      }
    }

    // Show preview if placement is valid
    if (canPlaceShip(x, y, currentShip.size, isVertical)) {
      for (let i = 0; i < currentShip.size; i++) {
        const previewX = isVertical ? x : x + i;
        const previewY = isVertical ? y + i : y;
        if (newBoard[previewY][previewX] === 'empty') {
          newBoard[previewY][previewX] = 'preview';
        }
      }
    }

    setBoard(newBoard);
  };

  const clearPreview = () => {
    const newBoard = board.map(row => [...row]);
    for (let row = 0; row < 10; row++) {
      for (let col = 0; col < 10; col++) {
        if (newBoard[row][col] === 'preview') {
          newBoard[row][col] = 'empty';
        }
      }
    }
    setBoard(newBoard);
  };

  const resetPlacement = () => {
    setBoard(Array(10).fill(null).map(() => Array(10).fill('empty')));
    setShips([]);
    setCurrentShipIndex(0);
  };

  const getCellClass = (cellState: string): string => {
    return `game-cell ${cellState}`;
  };

  const allShipsPlaced = currentShipIndex >= SHIP_TYPES.length;

  return (
    <div>
      <h2>Place Your Ships</h2>
      
      <div style={{ marginBottom: '1rem' }}>
        {!allShipsPlaced ? (
          <div>
            <p>
              Place your <strong>{currentShip.name}</strong> (Size: {currentShip.size})
            </p>
            <label>
              <input
                type="checkbox"
                checked={isVertical}
                onChange={(e) => setIsVertical(e.target.checked)}
              />
              Vertical orientation
            </label>
          </div>
        ) : (
          <p>All ships placed! Ready to start the game.</p>
        )}
      </div>

      <div 
        className="game-board"
        onMouseLeave={clearPreview}
      >
        {board.map((row, y) =>
          row.map((cell, x) => (
            <button
              key={`${x}-${y}`}
              className={getCellClass(cell)}
              onClick={() => handleCellClick(x, y)}
              onMouseEnter={() => handleCellMouseEnter(x, y)}
              disabled={allShipsPlaced}
            />
          ))
        )}
      </div>

      <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem' }}>
        <button onClick={resetPlacement}>Reset</button>
        {allShipsPlaced && (
          <button 
            onClick={() => onShipsPlaced(ships)}
            className="btn-primary"
          >
            Confirm Ship Placement
          </button>
        )}
      </div>

      <div style={{ marginTop: '1rem' }}>
        <h3>Ships to place:</h3>
        <ul>
          {SHIP_TYPES.map((ship, index) => (
            <li 
              key={ship.type}
              style={{ 
                textDecoration: index < currentShipIndex ? 'line-through' : 'none',
                color: index < currentShipIndex ? '#666' : 'inherit'
              }}
            >
              {ship.name} (Size: {ship.size})
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default ShipPlacement;
