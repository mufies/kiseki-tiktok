// React Component Example for HLS Live Stream Player
// Install: npm install video.js @types/video.js

import React, { useEffect, useRef, useState } from 'react';
import videojs from 'video.js';
import 'video.js/dist/video-js.css';

interface LiveStreamPlayerProps {
  streamId: string;
  serverUrl?: string;
  autoplay?: boolean;
  onError?: (error: any) => void;
  onPlaying?: () => void;
  onEnded?: () => void;
}

export const LiveStreamPlayer: React.FC<LiveStreamPlayerProps> = ({
  streamId,
  serverUrl = 'http://localhost:8083',
  autoplay = true,
  onError,
  onPlaying,
  onEnded,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const playerRef = useRef<any>(null);
  const [isLive, setIsLive] = useState(false);
  const [viewerCount, setViewerCount] = useState(0);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!videoRef.current) return;

    // Initialize Video.js player
    const player = videojs(videoRef.current, {
      controls: true,
      autoplay,
      preload: 'auto',
      liveui: true,
      fluid: true,
      html5: {
        vhs: {
          overrideNative: true,
          enableLowInitialPlaylist: true,
          smoothQualityChange: true,
        },
      },
    });

    playerRef.current = player;

    // Set HLS source
    const hlsUrl = `${serverUrl}/hls/${streamId}/playlist.m3u8`;
    player.src({
      src: hlsUrl,
      type: 'application/x-mpegURL',
    });

    // Event listeners
    player.on('error', (e: any) => {
      const err = player.error();
      console.error('Player error:', err);
      setError(err.message);
      onError?.(err);
    });

    player.on('loadedmetadata', () => {
      console.log('Stream metadata loaded');
      setIsLive(true);
    });

    player.on('playing', () => {
      console.log('Stream is playing');
      setIsLive(true);
      onPlaying?.();
    });

    player.on('ended', () => {
      console.log('Stream ended');
      setIsLive(false);
      onEnded?.();
    });

    player.on('waiting', () => {
      console.log('Buffering...');
    });

    // Cleanup
    return () => {
      if (player) {
        player.dispose();
      }
    };
  }, [streamId, serverUrl, autoplay, onError, onPlaying, onEnded]);

  // Join stream (increment viewer count)
  useEffect(() => {
    const joinStream = async () => {
      try {
        const response = await fetch(`${serverUrl}/streams/${streamId}/viewers/join`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            user_id: 'current-user-id', // Replace with actual user ID
          }),
        });

        if (response.ok) {
          const data = await response.json();
          setViewerCount(data.viewer_count || 0);
        }
      } catch (error) {
        console.error('Failed to join stream:', error);
      }
    };

    joinStream();

    // Leave stream on unmount
    return () => {
      fetch(`${serverUrl}/streams/${streamId}/viewers/leave`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: 'current-user-id',
        }),
      }).catch(console.error);
    };
  }, [streamId, serverUrl]);

  // Poll viewer count every 10 seconds
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`${serverUrl}/streams/${streamId}`);
        if (response.ok) {
          const data = await response.json();
          setViewerCount(data.viewer_count || 0);
        }
      } catch (error) {
        console.error('Failed to fetch viewer count:', error);
      }
    }, 10000);

    return () => clearInterval(interval);
  }, [streamId, serverUrl]);

  return (
    <div className="live-stream-player">
      {/* Live indicator */}
      {isLive && (
        <div className="live-indicator">
          <span className="live-badge">🔴 LIVE</span>
          <span className="viewer-count">👁️ {viewerCount} watching</span>
        </div>
      )}

      {/* Video player */}
      <div data-vjs-player>
        <video
          ref={videoRef}
          className="video-js vjs-big-play-centered"
        />
      </div>

      {/* Error message */}
      {error && (
        <div className="error-message">
          <p>❌ Error: {error}</p>
          <button onClick={() => window.location.reload()}>
            Retry
          </button>
        </div>
      )}

      <style jsx>{`
        .live-stream-player {
          position: relative;
          width: 100%;
          max-width: 1280px;
          margin: 0 auto;
        }

        .live-indicator {
          position: absolute;
          top: 10px;
          left: 10px;
          z-index: 10;
          display: flex;
          gap: 10px;
          align-items: center;
        }

        .live-badge {
          background: #ff0000;
          color: white;
          padding: 4px 12px;
          border-radius: 4px;
          font-weight: bold;
          font-size: 14px;
          animation: pulse 2s infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.7; }
        }

        .viewer-count {
          background: rgba(0, 0, 0, 0.7);
          color: white;
          padding: 4px 12px;
          border-radius: 4px;
          font-size: 14px;
        }

        .error-message {
          padding: 20px;
          background: #ffebee;
          color: #c62828;
          border-radius: 8px;
          margin-top: 10px;
          text-align: center;
        }

        .error-message button {
          margin-top: 10px;
          padding: 8px 16px;
          background: #c62828;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }

        .error-message button:hover {
          background: #b71c1c;
        }
      `}</style>
    </div>
  );
};

// Example usage:
/*
import { LiveStreamPlayer } from './LiveStreamPlayer';

function App() {
  return (
    <div>
      <h1>Live Stream</h1>
      <LiveStreamPlayer
        streamId="550e8400-e29b-41d4-a716-446655440000"
        serverUrl="http://localhost:8083"
        autoplay={true}
        onPlaying={() => console.log('Started playing!')}
        onError={(error) => console.error('Stream error:', error)}
        onEnded={() => console.log('Stream ended')}
      />
    </div>
  );
}
*/
