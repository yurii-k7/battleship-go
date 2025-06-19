import React, { useState, useCallback } from 'react';
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
  const [isPlacing, setIsPlacing] = useState(false);

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

  const placeShip = useCallback((x: number, y: number) => {
    if (isPlacing || !currentShip || !canPlaceShip(x, y, currentShip.size, isVertical)) {
      return;
    }

    setIsPlacing(true);

    // Create a clean board without preview cells
    const newBoard = board.map(row =>
      row.map(cell => (cell === 'preview' || cell === 'invalid') ? 'empty' : cell)
    );

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

    // Update all state in a single batch
    setBoard(newBoard);
    setShips(prevShips => [...prevShips, newShip]);
    setCurrentShipIndex(prevIndex => prevIndex + 1);

    // Reset placing flag after a short delay
    setTimeout(() => setIsPlacing(false), 100);
  }, [board, currentShip, isVertical, isPlacing, canPlaceShip]);

  const handleCellClick = useCallback((x: number, y: number) => {
    // Prevent multiple clicks and ensure we have a valid ship to place
    if (isPlacing || currentShipIndex >= SHIP_TYPES.length || !currentShip) {
      return;
    }

    // Only place if the position is valid
    if (canPlaceShip(x, y, currentShip.size, isVertical)) {
      placeShip(x, y);
    }
  }, [isPlacing, currentShipIndex, currentShip, canPlaceShip, isVertical, placeShip]);

  const handleCellMouseEnter = useCallback((x: number, y: number) => {
    if (currentShipIndex >= SHIP_TYPES.length || !currentShip || isPlacing) return;

    // Use a more efficient approach to update preview
    setBoard(prevBoard => {
      const newBoard = prevBoard.map(row =>
        row.map(cell => (cell === 'preview' || cell === 'invalid') ? 'empty' : cell)
      );

      const isValidPlacement = canPlaceShip(x, y, currentShip.size, isVertical);

      // Show preview for valid placement or invalid indicator
      for (let i = 0; i < currentShip.size; i++) {
        const previewX = isVertical ? x : x + i;
        const previewY = isVertical ? y + i : y;

        // Check bounds
        if (previewX >= 0 && previewX < 10 && previewY >= 0 && previewY < 10) {
          if (newBoard[previewY][previewX] === 'empty') {
            newBoard[previewY][previewX] = isValidPlacement ? 'preview' : 'invalid';
          }
        }
      }

      return newBoard;
    });
  }, [currentShipIndex, currentShip, isPlacing, isVertical, canPlaceShip]);

  const clearPreview = () => {
    setBoard(prevBoard =>
      prevBoard.map(row =>
        row.map(cell => (cell === 'preview' || cell === 'invalid') ? 'empty' : cell)
      )
    );
  };

  const resetPlacement = () => {
    setBoard(Array(10).fill(null).map(() => Array(10).fill('empty')));
    setShips([]);
    setCurrentShipIndex(0);
    setIsPlacing(false);
  };

  const handleOrientationChange = (vertical: boolean) => {
    setIsVertical(vertical);
    // Clear any existing preview when orientation changes
    clearPreview();
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
                onChange={(e) => handleOrientationChange(e.target.checked)}
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
