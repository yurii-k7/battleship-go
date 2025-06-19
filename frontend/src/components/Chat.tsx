import React, { useState, useEffect, useRef } from 'react';
import { ChatMessage } from '../types';

interface ChatProps {
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  currentUserId: number;
}

const Chat: React.FC<ChatProps> = ({ messages, onSendMessage, currentUserId }) => {
  const [newMessage, setNewMessage] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (newMessage.trim()) {
      onSendMessage(newMessage.trim());
      setNewMessage('');
    }
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="chat-container">
      <h3 style={{ padding: '1rem', margin: 0, borderBottom: '1px solid #eee' }}>
        Game Chat
      </h3>
      
      <div className="chat-messages">
        {messages.length === 0 ? (
          <p style={{ color: '#666', textAlign: 'center', margin: '2rem 0' }}>
            No messages yet. Start the conversation!
          </p>
        ) : (
          messages.map((message) => (
            <div 
              key={message.id} 
              className={`chat-message ${message.player_id === currentUserId ? 'own-message' : ''}`}
            >
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center',
                marginBottom: '0.25rem'
              }}>
                <strong>
                  {message.player_id === currentUserId ? 'You' : 'Opponent'}
                </strong>
                <span style={{ fontSize: '0.8em', color: '#666' }}>
                  {formatTime(message.created_at)}
                </span>
              </div>
              <div>{message.message}</div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
      
      <form onSubmit={handleSubmit} className="chat-input">
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Type a message..."
          maxLength={500}
        />
        <button type="submit" disabled={!newMessage.trim()}>
          Send
        </button>
      </form>
    </div>
  );
};

export default Chat;
