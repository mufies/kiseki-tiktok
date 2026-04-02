import { useEffect, useRef, useState } from 'react';
import type { ChatMessage } from '../types';

interface StreamChatProps {
  streamId: string;
  currentUserId: string;
  currentUsername: string;
  streamOwnerId: string;
}

export default function StreamChat({ streamId, currentUserId, currentUsername, streamOwnerId }: StreamChatProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const channelRef = useRef<BroadcastChannel | null>(null);

  // Initialize BroadcastChannel and load existing messages
  useEffect(() => {
    // Load messages from localStorage
    const loadMessages = () => {
      const storedMessages = localStorage.getItem(`stream_chat_${streamId}`);
      if (storedMessages) {
        try {
          setMessages(JSON.parse(storedMessages));
        } catch (error) {
          console.error('Failed to parse stored messages:', error);
        }
      }
    };

    loadMessages();

    // Create BroadcastChannel for real-time sync across tabs
    const channel = new BroadcastChannel(`stream_chat_${streamId}`);
    channelRef.current = channel;

    channel.onmessage = (event) => {
      if (event.data.type === 'new_message') {
        setMessages((prev) => {
          const newMessages = [...prev, event.data.message];
          // Keep only last 100 messages
          const limitedMessages = newMessages.slice(-100);
          localStorage.setItem(`stream_chat_${streamId}`, JSON.stringify(limitedMessages));
          return limitedMessages;
        });
      }
    };

    // Poll for updates every 2 seconds (fallback for cross-origin scenarios)
    const pollInterval = setInterval(() => {
      loadMessages();
    }, 2000);

    return () => {
      channel.close();
      clearInterval(pollInterval);
    };
  }, [streamId]);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();

    if (!inputValue.trim()) return;

    const newMessage: ChatMessage = {
      id: `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      streamId,
      userId: currentUserId,
      username: currentUsername,
      content: inputValue.trim(),
      timestamp: new Date().toISOString(),
      isStreamer: currentUserId === streamOwnerId,
    };

    // Add to local state
    setMessages((prev) => {
      const newMessages = [...prev, newMessage];
      const limitedMessages = newMessages.slice(-100);
      localStorage.setItem(`stream_chat_${streamId}`, JSON.stringify(limitedMessages));
      return limitedMessages;
    });

    // Broadcast to other tabs
    channelRef.current?.postMessage({
      type: 'new_message',
      message: newMessage,
    });

    setInputValue('');
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (seconds < 60) return 'just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex flex-col h-full bg-gray-900 text-white">
      {/* Header */}
      <div className="flex-shrink-0 px-4 py-3 border-b border-gray-700">
        <h3 className="font-semibold text-lg">Live Chat</h3>
        <p className="text-xs text-gray-400">{messages.length} messages</p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
        {messages.length === 0 ? (
          <div className="text-center text-gray-500 mt-8">
            <p>No messages yet</p>
            <p className="text-sm mt-1">Be the first to say something!</p>
          </div>
        ) : (
          messages.map((message) => (
            <div key={message.id} className="flex gap-2">
              {/* Avatar */}
              <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-xs font-bold">
                {message.username.charAt(0).toUpperCase()}
              </div>

              {/* Message content */}
              <div className="flex-1 min-w-0">
                <div className="flex items-baseline gap-2">
                  <span className={`font-semibold text-sm ${message.isStreamer ? 'text-red-500' : 'text-gray-200'}`}>
                    {message.username}
                    {message.isStreamer && (
                      <span className="ml-1 px-1.5 py-0.5 text-xs bg-red-500 text-white rounded">
                        STREAMER
                      </span>
                    )}
                  </span>
                  <span className="text-xs text-gray-500">{formatTimestamp(message.timestamp)}</span>
                </div>
                <p className="text-sm text-gray-300 break-words">{message.content}</p>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="flex-shrink-0 px-4 py-3 border-t border-gray-700">
        <form onSubmit={handleSendMessage} className="flex gap-2">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Send a message..."
            className="flex-1 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-purple-500 text-sm"
            maxLength={200}
          />
          <button
            type="submit"
            disabled={!inputValue.trim()}
            className="px-4 py-2 bg-purple-600 text-white rounded-lg font-semibold text-sm hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}
