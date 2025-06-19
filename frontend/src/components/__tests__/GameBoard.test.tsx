import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import GameBoard from '../GameBoard';
import { CellState } from '../../types';

describe('GameBoard', () => {
  const mockOnCellClick = jest.fn();

  const createBoard = (cellState: CellState = 'empty'): CellState[][] => {
    return Array(10).fill(null).map(() => Array(10).fill(cellState));
  };

  beforeEach(() => {
    mockOnCellClick.mockClear();
  });

  it('renders a 10x10 grid', () => {
    const board = createBoard();
    render(<GameBoard board={board} onCellClick={mockOnCellClick} />);
    
    const cells = screen.getAllByRole('button');
    expect(cells).toHaveLength(100);
  });

  it('calls onCellClick when a cell is clicked', () => {
    const board = createBoard();
    render(<GameBoard board={board} onCellClick={mockOnCellClick} />);
    
    const firstCell = screen.getAllByRole('button')[0];
    fireEvent.click(firstCell);
    
    expect(mockOnCellClick).toHaveBeenCalledWith(0, 0);
  });

  it('applies correct CSS classes based on cell state', () => {
    const board = createBoard();
    board[0][0] = 'ship';
    board[0][1] = 'hit';
    board[0][2] = 'miss';
    
    render(<GameBoard board={board} onCellClick={mockOnCellClick} />);
    
    const cells = screen.getAllByRole('button');
    expect(cells[0]).toHaveClass('game-cell', 'ship');
    expect(cells[1]).toHaveClass('game-cell', 'hit');
    expect(cells[2]).toHaveClass('game-cell', 'miss');
  });

  it('disables cells when disabled prop is true', () => {
    const board = createBoard();
    render(<GameBoard board={board} onCellClick={mockOnCellClick} disabled={true} />);
    
    const cells = screen.getAllByRole('button');
    cells.forEach(cell => {
      expect(cell).toBeDisabled();
    });
  });

  it('disables hit and miss cells', () => {
    const board = createBoard();
    board[0][0] = 'hit';
    board[0][1] = 'miss';
    
    render(<GameBoard board={board} onCellClick={mockOnCellClick} />);
    
    const cells = screen.getAllByRole('button');
    expect(cells[0]).toBeDisabled(); // hit cell
    expect(cells[1]).toBeDisabled(); // miss cell
    expect(cells[2]).not.toBeDisabled(); // empty cell
  });

  it('shows correct title attribute for cells', () => {
    const board = createBoard();
    render(<GameBoard board={board} onCellClick={mockOnCellClick} />);
    
    const firstCell = screen.getAllByRole('button')[0];
    expect(firstCell).toHaveAttribute('title', '0, 0');
    
    const lastCell = screen.getAllByRole('button')[99];
    expect(lastCell).toHaveAttribute('title', '9, 9');
  });
});
