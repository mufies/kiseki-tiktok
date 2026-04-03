import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Radio, Users } from 'lucide-react';
import { streamAPI } from '../api/stream';
import type { Stream } from '../api/stream';
import { userAPI } from '../api/user';
import type { UserProfile } from '../api/user';
import { useAuth } from '../context/AuthContext';
import StreamPlayer from '../components/StreamPlayer';
import StreamChat from '../components/StreamChat';

export default function WatchStream() {
  const { username } = useParams<{ username: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [stream, setStream] = useState<Stream | null>(null);
  const [hlsUrl, setHlsUrl] = useState('');
  const [viewerCount, setViewerCount] = useState(0);

  // Load user and stream data
  useEffect(() => {
    if (!username) {
      setError('Username is required');
      setLoading(false);
      return;
    }

    const loadStream = async () => {
      try {
        setLoading(true);
        setError(null);

        // Get user by username
        const profile = await userAPI.getUserByUsername(username);
        setUserProfile(profile);

        // Get user's streams
        const streams = await streamAPI.getUserStreams(profile.user_id || profile.id);

        // Find live stream
        const liveStream = streams.find((s) => s.status === 'live');

        if (!liveStream) {
          setError('User is not currently live');
          setLoading(false);
          return;
        }

        setStream(liveStream);

        // Get playback URL
        const playback = await streamAPI.getPlaybackUrl(liveStream.id);
        setHlsUrl(playback.hls_url);

        // Join stream to increment viewer count
        await streamAPI.joinStream(liveStream.id);
        setViewerCount(liveStream.viewer_count + 1);

        setLoading(false);
      } catch (err) {
        console.error('Failed to load stream:', err);
        setError('Failed to load stream. User may not exist or is offline.');
        setLoading(false);
      }
    };

    loadStream();

    // Leave stream when component unmounts
    return () => {
      if (stream) {
        streamAPI.leaveStream(stream.id).catch(console.error);
      }
    };
  }, [username]);

  // Poll for viewer count updates
  useEffect(() => {
    if (!stream) return;

    const updateViewerCount = async () => {
      try {
        const updated = await streamAPI.getStream(stream.id);
        setViewerCount(updated.viewer_count);

        // If stream ended, show offline message
        if (updated.status !== 'live') {
          setError('Stream has ended');
        }
      } catch (error) {
        console.error('Failed to update viewer count:', error);
      }
    };

    const interval = setInterval(updateViewerCount, 5000);
    return () => clearInterval(interval);
  }, [stream]);

  if (loading) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="text-white text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500 mx-auto mb-4"></div>
          <p className="text-lg">Loading stream...</p>
        </div>
      </div>
    );
  }

  if (error || !stream || !userProfile || !currentUser) {
    return (
      <div className="min-h-screen bg-black">
        {/* Header */}
        <div className="bg-gray-900 border-b border-gray-800 px-6 py-4">
          <div className="max-w-7xl mx-auto flex items-center gap-4">
            <button
              onClick={() => navigate('/')}
              className="text-white hover:text-gray-300 transition-colors"
            >
              <ArrowLeft className="w-6 h-6" />
            </button>
            <h1 className="text-white text-xl font-semibold">@{username}</h1>
          </div>
        </div>

        {/* Offline message */}
        <div className="max-w-3xl mx-auto px-6 py-20">
          <div className="bg-gray-900 rounded-lg p-12 text-center">
            <div className="w-20 h-20 rounded-full bg-gray-800 flex items-center justify-center mx-auto mb-6">
              <Radio className="w-10 h-10 text-gray-600" />
            </div>
            <h2 className="text-white text-2xl font-bold mb-2">User is Offline</h2>
            <p className="text-gray-400 mb-6">
              {error || `@${username} is not currently live. Check back later!`}
            </p>
            <button
              onClick={() => navigate('/')}
              className="px-6 py-3 bg-purple-600 text-white rounded-lg font-semibold hover:bg-purple-700 transition-colors"
            >
              Back to Home
            </button>
          </div>
        </div>
      </div>
    );
  }

  const startedTime = stream.started_at ? new Date(stream.started_at) : null;
  const timeAgo = startedTime ? getTimeAgo(startedTime) : '';

  return (
    <div className="min-h-screen bg-black">
      {/* Header */}
      <div className="bg-gray-900 border-b border-gray-800 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/')}
              className="text-white hover:text-gray-300 transition-colors"
            >
              <ArrowLeft className="w-6 h-6" />
            </button>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-bold">
                {userProfile.username.charAt(0).toUpperCase()}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-white text-lg font-semibold">@{userProfile.username}</h1>
                  <div className="flex items-center gap-1 px-2 py-0.5 bg-red-600 rounded text-white text-xs font-bold">
                    <Radio className="w-3 h-3" />
                    LIVE
                  </div>
                </div>
                <p className="text-gray-400 text-sm">{stream.title}</p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 text-white">
            <Users className="w-5 h-5 text-purple-500" />
            <span className="font-semibold">{viewerCount.toLocaleString()}</span>
            <span className="text-gray-400 text-sm">viewers</span>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto p-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Stream player */}
          <div className="lg:col-span-2 space-y-4">
            {hlsUrl && hlsUrl.startsWith('http') ? (
              <StreamPlayer hlsUrl={hlsUrl} poster={stream.thumbnail_url} />
            ) : (
              <div className="aspect-video bg-gray-800 rounded-lg flex items-center justify-center">
                <div className="text-center text-white">
                  <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500 mx-auto mb-4"></div>
                  <p className="text-lg">Loading stream...</p>
                </div>
              </div>
            )}

            {/* Stream info */}
            <div className="bg-gray-900 rounded-lg p-4 text-white">
              <h2 className="text-xl font-bold mb-2">{stream.title}</h2>
              {stream.description && (
                <p className="text-gray-400 mb-3">{stream.description}</p>
              )}
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2 text-gray-400">
                  <span>@{userProfile.username}</span>
                  <span>•</span>
                  <span>Started {timeAgo}</span>
                </div>
                <button
                  onClick={() => navigate(`/user/${userProfile.user_id || userProfile.id}`)}
                  className="px-4 py-2 bg-purple-600 rounded-lg font-semibold hover:bg-purple-700 transition-colors"
                >
                  View Profile
                </button>
              </div>
            </div>
          </div>

          {/* Chat */}
          <div className="lg:col-span-1 h-[600px]">
            <StreamChat
              streamId={stream.id}
              currentUserId={currentUser.user_id || currentUser.id}
              currentUsername={currentUser.username}
              streamOwnerId={stream.user_id}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function getTimeAgo(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return 'just now';
}
