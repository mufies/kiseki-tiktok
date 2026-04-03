import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Radio, Users, Eye } from 'lucide-react';
import { streamAPI } from '../api/stream';
import type { Stream } from '../api/stream';
import { userAPI } from '../api/user';

interface StreamWithUser extends Stream {
  streamerUsername?: string;
}

export default function LiveStreams() {
  const navigate = useNavigate();
  const [streams, setStreams] = useState<StreamWithUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadLiveStreams();

    // Poll for updates every 10 seconds
    const interval = setInterval(loadLiveStreams, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadLiveStreams = async () => {
    try {
      setError(null);
      const response = await streamAPI.getLiveStreams(50, 0);

      // Fetch usernames for each stream
      const streamsWithUsers = await Promise.all(
        response.streams.map(async (stream) => {
          try {
            const user = await userAPI.getUserProfile(stream.user_id);
            return { ...stream, streamerUsername: user.username };
          } catch {
            return { ...stream, streamerUsername: 'Unknown' };
          }
        })
      );

      setStreams(streamsWithUsers);
    } catch (err) {
      console.error('Failed to load live streams:', err);
      setError('Failed to load live streams');
    } finally {
      setLoading(false);
    }
  };

  const getTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return 'just now';
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="text-white text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500 mx-auto mb-4"></div>
          <p className="text-lg">Loading live streams...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black">
      {/* Header */}
      <div className="bg-gray-900 border-b border-gray-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center gap-4">
          <button
            onClick={() => navigate('/')}
            className="text-white hover:text-gray-300 transition-colors"
          >
            <ArrowLeft className="w-6 h-6" />
          </button>
          <div className="flex items-center gap-2">
            <Radio className="w-6 h-6 text-red-500" />
            <h1 className="text-white text-2xl font-bold">Live Streams</h1>
          </div>
          <div className="ml-auto text-gray-400 text-sm">
            {streams.length} {streams.length === 1 ? 'stream' : 'streams'} live
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-6 py-8">
        {error && (
          <div className="bg-red-900/30 border border-red-700/50 rounded-lg p-4 mb-6">
            <p className="text-red-300">{error}</p>
          </div>
        )}

        {streams.length === 0 ? (
          <div className="text-center py-20">
            <div className="w-20 h-20 rounded-full bg-gray-800 flex items-center justify-center mx-auto mb-4">
              <Radio className="w-10 h-10 text-gray-600" />
            </div>
            <h2 className="text-white text-2xl font-bold mb-2">No Live Streams</h2>
            <p className="text-gray-400 mb-6">
              No one is streaming right now. Be the first to go live!
            </p>
            <button
              onClick={() => navigate('/go-live')}
              className="px-6 py-3 bg-red-600 text-white rounded-lg font-semibold hover:bg-red-700 transition-colors inline-flex items-center gap-2"
            >
              <Radio className="w-5 h-5" />
              Go Live
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {streams.map((stream) => (
              <div
                key={stream.id}
                onClick={() => navigate(`/stream/${stream.streamerUsername || stream.user_id}`)}
                className="bg-gray-900 rounded-lg overflow-hidden cursor-pointer hover:ring-2 hover:ring-purple-500 transition-all group"
              >
                {/* Thumbnail */}
                <div className="relative aspect-video bg-gray-800">
                  {stream.thumbnail_url ? (
                    <img
                      src={stream.thumbnail_url}
                      alt={stream.title}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center">
                      <Eye className="w-12 h-12 text-gray-600" />
                    </div>
                  )}

                  {/* Live badge */}
                  <div className="absolute top-3 left-3 px-2 py-1 bg-red-600 rounded flex items-center gap-1 text-white text-xs font-bold">
                    <Radio className="w-3 h-3 animate-pulse" />
                    LIVE
                  </div>

                  {/* Viewer count */}
                  <div className="absolute bottom-3 left-3 px-2 py-1 bg-black/75 rounded flex items-center gap-1 text-white text-xs">
                    <Users className="w-3 h-3" />
                    {stream.viewer_count.toLocaleString()}
                  </div>
                </div>

                {/* Stream info */}
                <div className="p-4">
                  <div className="flex gap-3">
                    {/* Avatar */}
                    <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-bold">
                      {(stream.streamerUsername || 'U').charAt(0).toUpperCase()}
                    </div>

                    {/* Details */}
                    <div className="flex-1 min-w-0">
                      <h3 className="text-white font-semibold text-sm mb-1 truncate group-hover:text-purple-400 transition-colors">
                        {stream.title}
                      </h3>
                      <p className="text-gray-400 text-xs mb-1">
                        @{stream.streamerUsername || 'Unknown'}
                      </p>
                      {stream.started_at && (
                        <p className="text-gray-500 text-xs">
                          Started {getTimeAgo(stream.started_at)}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
