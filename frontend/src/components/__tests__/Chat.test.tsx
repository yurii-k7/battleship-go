import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import Chat from '../Chat';
import { ChatMessage } from '../../types';

describe('Chat', () => {
  const mockOnSendMessage = jest.fn();
  const currentUserId = 1;

  const mockMessages: ChatMessage[] = [
    {
      id: 1,
      game_id: 1,
      player_id: 1,
      message: 'Hello!',
      created_at: '2023-01-01T10:00:00Z'
    },
    {
      id: 2,
      game_id: 1,
      player_id: 2,
      message: 'Hi there!',
      created_at: '2023-01-01T10:01:00Z'
    }
  ];

  beforeEach(() => {
    mockOnSendMessage.mockClear();
  });

  it('renders chat messages', () => {
    render(
      <Chat 
        messages={mockMessages} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    expect(screen.getByText('Hello!')).toBeInTheDocument();
    expect(screen.getByText('Hi there!')).toBeInTheDocument();
  });

  it('shows "You" for current user messages', () => {
    render(
      <Chat 
        messages={mockMessages} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    expect(screen.getByText('You')).toBeInTheDocument();
    expect(screen.getByText('Opponent')).toBeInTheDocument();
  });

  it('sends message when form is submitted', async () => {
    render(
      <Chat 
        messages={[]} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    const input = screen.getByPlaceholderText('Type a message...');
    const sendButton = screen.getByText('Send');
    
    fireEvent.change(input, { target: { value: 'Test message' } });
    fireEvent.click(sendButton);
    
    expect(mockOnSendMessage).toHaveBeenCalledWith('Test message');
    
    await waitFor(() => {
      expect(input).toHaveValue('');
    });
  });

  it('disables send button when input is empty', () => {
    render(
      <Chat 
        messages={[]} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    const sendButton = screen.getByText('Send');
    expect(sendButton).toBeDisabled();
  });

  it('enables send button when input has text', () => {
    render(
      <Chat 
        messages={[]} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    const input = screen.getByPlaceholderText('Type a message...');
    const sendButton = screen.getByText('Send');
    
    fireEvent.change(input, { target: { value: 'Test' } });
    expect(sendButton).not.toBeDisabled();
  });

  it('shows empty state when no messages', () => {
    render(
      <Chat 
        messages={[]} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    expect(screen.getByText('No messages yet. Start the conversation!')).toBeInTheDocument();
  });

  it('prevents sending empty messages', () => {
    render(
      <Chat 
        messages={[]} 
        onSendMessage={mockOnSendMessage} 
        currentUserId={currentUserId} 
      />
    );
    
    const input = screen.getByPlaceholderText('Type a message...');
    const sendButton = screen.getByText('Send');
    
    fireEvent.change(input, { target: { value: '   ' } }); // Only whitespace
    fireEvent.click(sendButton);
    
    expect(mockOnSendMessage).not.toHaveBeenCalled();
  });
});
