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
  const currentShip = SHIP_TYPES[currentShipIndex];

  // Debug logging
  console.log('ShipPlacement render:', {
    currentShipIndex,
    currentShip: currentShip?.name,
    shipsPlaced: ships.length,
    isVertical
  });

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
      console.log('Cannot place ship at', x, y);
      return;
    }

    console.log('Placing ship at', x, y, 'vertical:', isVertical);

    // Create a new board
    const newBoard = board.map((row: string[]) => [...row]);

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

    // Update state
    setBoard(newBoard);
    setShips([...ships, newShip]);
    setCurrentShipIndex(currentShipIndex + 1);
  };

  const handleCellClick = (x: number, y: number) => {
    console.log('Cell clicked:', x, y);

    // Prevent multiple clicks and ensure we have a valid ship to place
    if (currentShipIndex >= SHIP_TYPES.length || !currentShip) {
      console.log('No more ships to place');
      return;
    }

    // Only place if the position is valid
    if (canPlaceShip(x, y, currentShip.size, isVertical)) {
      placeShip(x, y);
    } else {
      console.log('Invalid placement position');
    }
  };

  // Simple cell class function - we'll add hover effects with CSS
  const getCellClass = (x: number, y: number, cellState: string): string => {
    return `game-cell ${cellState}`;
  };

  const resetPlacement = () => {
    setBoard(Array(10).fill(null).map(() => Array(10).fill('empty')));
    setShips([]);
    setCurrentShipIndex(0);
  };

  const handleOrientationChange = (vertical: boolean) => {
    setIsVertical(vertical);
  };

  const allShipsPlaced = currentShipIndex >= SHIP_TYPES.length;

  return (
    <div>
      <h2>Place Your Ships</h2>
      
      <div style={{ marginBottom: '1rem' }}>
        {!allShipsPlaced ? (
          <div>
            <p>
              <strong>Step {currentShipIndex + 1} of {SHIP_TYPES.length}:</strong> Place your <strong>{currentShip.name}</strong> (Size: {currentShip.size} cells)
            </p>
            <p style={{ fontSize: '0.9em', color: '#666' }}>
              Click on the board to place your ship. Use the checkbox below to change orientation.
            </p>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input
                type="checkbox"
                checked={isVertical}
                onChange={(e) => handleOrientationChange(e.target.checked)}
              />
              <span>Vertical orientation {isVertical ? '(↓)' : '(→)'}</span>
            </label>
          </div>
        ) : (
          <div>
            <p style={{ color: 'green', fontWeight: 'bold' }}>✅ All ships placed! Ready to start the game.</p>
          </div>
        )}
      </div>

      <div className="game-board">
        {board.map((row, y) =>
          row.map((cell, x) => (
            <button
              key={`${x}-${y}`}
              className={getCellClass(x, y, cell)}
              onClick={() => handleCellClick(x, y)}
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
        <h3>Ship Placement Progress:</h3>
        <div className="ship-placement-progress">
          {SHIP_TYPES.map((ship, index) => (
            <div
              key={ship.type}
              className={`ship-placement-step ${
                index < currentShipIndex ? 'completed' :
                index === currentShipIndex ? 'current' : 'pending'
              }`}
            >
              {ship.name} ({ship.size})
            </div>
          ))}
        </div>

        <div style={{ marginTop: '1rem', fontSize: '0.9em', color: '#666' }}>
          <p><strong>Instructions:</strong></p>
          <ul style={{ margin: '0.5rem 0', paddingLeft: '1.5rem' }}>
            <li>Click any empty cell on the board to place your ship</li>
            <li>Use the orientation checkbox to place ships vertically or horizontally</li>
            <li>Ships cannot overlap or go outside the board</li>
            <li>Green cells show valid placement, red cells show invalid placement</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ShipPlacement;
