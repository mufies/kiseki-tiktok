import { useState, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { userAPI, type UserProfile } from '../api/user';
import { videoAPI } from '../api/video';
import { interactionAPI } from '../api/interaction';
import { type Video } from '../types';
import { ChevronLeft, X, Volume2, VolumeX, Send, Grid } from 'lucide-react';

interface Comment {
  id: string;
  author: string;
  content: string;
  timestamp: string;
}

export default function UserProfile() {
  const navigate = useNavigate();
  const { userId } = useParams<{ userId: string }>();
  const { user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isFollowing, setIsFollowing] = useState(false);
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null);
  const [selectedVideoUrl, setSelectedVideoUrl] = useState<string | null>(null);
  const [isMuted, setIsMuted] = useState(true);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentInput, setCommentInput] = useState('');
  const [showComments, setShowComments] = useState(false);
  const videoRef = useRef<HTMLVideoElement>(null);
  const commentsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (userId && user) {
      // Check if viewing own profile
      const currentUserId = user.id || user.user_id;
      if (userId === currentUserId) {
        navigate('/profile');
        return;
      }
      loadUserProfile();
    }
  }, [userId, user, navigate]);

  const loadUserProfile = async () => {
    if (!userId || !user) return;

    try {
      setLoading(true);
      setError(null);

      // Load user profile
      const userProfile = await userAPI.getUserProfile(userId);
      setProfile(userProfile);

      // Load user's videos
      const userVideos = await videoAPI.getUserVideos(userId);
      setVideos(userVideos);

      // Check if current user is following this user
      const currentUserId = user.id || user.user_id;
      if (currentUserId) {
        const followingList = await userAPI.getFollowing(currentUserId);
        const isFollowingUser = followingList.some(
          (u) => (u.id || u.user_id) === userId
        );
        setIsFollowing(isFollowingUser);
      }
    } catch (err) {
      console.error('Failed to load user profile:', err);
      setError('Failed to load user profile');
    } finally {
      setLoading(false);
    }
  };

  const handleFollowToggle = async () => {
    if (!userId) return;

    try {
      if (isFollowing) {
        await userAPI.unfollowUser(userId);
        setIsFollowing(false);
        setProfile((prev) =>
          prev ? { ...prev, followerCount: (prev.followerCount || 0) - 1 } : null
        );
      } else {
        await userAPI.followUser(userId);
        setIsFollowing(true);
        setProfile((prev) =>
          prev ? { ...prev, followerCount: (prev.followerCount || 0) + 1 } : null
        );
      }
    } catch (err) {
      console.error('Failed to toggle follow:', err);
      setError('Failed to update follow status');
    }
  };

  const handleBlockUser = () => {
    alert('Block feature: Coming soon! This will block the user from interacting with you.');
  };

  const handleVideoClick = async (video: Video) => {
    try {
      const videoId = video.id || video.video_id;
      if (!videoId) return;
      const videoData = await videoAPI.getVideo(videoId);
      setSelectedVideo(video);
      setSelectedVideoUrl(videoData.streamUrl || null);
      setIsMuted(true);
      setShowComments(false);
      setCommentInput('');
      const videoComments = await interactionAPI.getComments(videoId);
      setComments(
        (videoComments as any).map((comment: any) => ({
          id: comment.id,
          author: comment.username,
          content: comment.content,
          timestamp: new Date(comment.createdAt).toLocaleString('en-US', {
            timeStyle: 'short',
            dateStyle: 'short',
          }),
        }))
      );
    } catch (err) {
      console.error('Failed to load video:', err);
      setError('Failed to load video');
    }
  };

  const closeVideoViewer = () => {
    setSelectedVideo(null);
    setSelectedVideoUrl(null);
    setShowComments(false);
    if (videoRef.current) {
      videoRef.current.pause();
    }
  };

  const handleAddComment = async () => {
    if (!commentInput.trim() || !user) return;

    try {
      const videoId = selectedVideo?.id || selectedVideo?.video_id;
      if (!videoId) return;

      // Optimistic update
      const newComment: Comment = {
        id: Date.now().toString(),
        author: user.username,
        content: commentInput,
        timestamp: 'just now',
      };

      setComments([...comments, newComment]);
      setCommentInput('');

      // Call API
      await interactionAPI.addComment(videoId, commentInput);

      // Auto-scroll to bottom
      setTimeout(() => {
        commentsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    } catch (error) {
      console.error('Failed to add comment:', error);
      // Revert on error
      setComments(comments.slice(0, -1));
    }
  };

  // Auto-play when video loads
  useEffect(() => {
    if (selectedVideoUrl && videoRef.current) {
      videoRef.current.play().catch(() => {
        // Autoplay failed, that's ok
      });
    }
  }, [selectedVideoUrl]);

  if (loading) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="w-10 h-10 border-4 border-gray-600 border-t-white rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen bg-black flex flex-col items-center justify-center text-white">
        <p className="text-xl mb-4 text-red-400">{error || 'User not found'}</p>
        <button
          onClick={() => navigate('/')}
          className="bg-zinc-800 hover:bg-zinc-700 px-6 py-2 rounded-md font-semibold transition"
        >
          Back to Home
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white pb-20">
      {/* Top Header */}
      <header className="sticky top-0 bg-black/80 backdrop-blur-md z-40 border-b border-zinc-800 flex items-center justify-between p-4">
        <button onClick={() => navigate(-1)} className="text-white hover:text-zinc-300 transition">
          <ChevronLeft size={28} />
        </button>
        <h1 className="font-bold text-lg">{profile.username || 'User Profile'}</h1>
        <div className="w-7"></div>
      </header>

      {/* Profile Info */}
      <div className="flex flex-col items-center pt-8 pb-4 px-4 max-w-2xl mx-auto">
        {/* Avatar */}
        <div className="w-28 h-28 rounded-full bg-zinc-800 mb-4 overflow-hidden border-2 border-zinc-800 flex items-center justify-center object-cover">
          {profile.avatarUrl ? (
            <img src={profile.avatarUrl} alt={profile.username} className="w-full h-full object-cover" />
          ) : (
            <span className="text-4xl text-zinc-500 uppercase">{profile.username?.charAt(0) || 'U'}</span>
          )}
        </div>

        <h2 className="text-xl font-bold mb-1">@{profile.username}</h2>
        <p className="text-zinc-400 text-sm mb-6">{profile.email}</p>

        {/* Stats */}
        <div className="flex gap-8 mb-6">
          <div className="flex flex-col items-center">
            <span className="font-bold text-lg">{profile.followingCount || 0}</span>
            <span className="text-zinc-400 text-sm">Following</span>
          </div>
          <div className="flex flex-col items-center">
            <span className="font-bold text-lg">{profile.followerCount || 0}</span>
            <span className="text-zinc-400 text-sm">Followers</span>
          </div>
          <div className="flex flex-col items-center">
            <span className="font-bold text-lg">{videos.length * 12}</span>
            <span className="text-zinc-400 text-sm">Likes</span>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2 w-full max-w-xs justify-center">
          <button
            className={`flex-1 px-8 py-2 rounded-md font-semibold text-sm transition ${
              isFollowing
                ? 'bg-zinc-800 hover:bg-zinc-700 text-white'
                : 'bg-red-500 hover:bg-red-600 text-white'
            }`}
            onClick={handleFollowToggle}
          >
            {isFollowing ? 'Following' : 'Follow'}
          </button>
          <button
            className="bg-zinc-800 hover:bg-zinc-700 px-6 py-2 rounded-md font-semibold text-sm transition"
            onClick={handleBlockUser}
          >
            Block
          </button>
        </div>
      </div>

      {/* Video Grid */}
      <div className="max-w-4xl mx-auto w-full">
        {videos.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-zinc-500">
            <Grid size={48} className="mb-4 opacity-50" />
            <p>No videos yet</p>
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-0.5 sm:gap-1">
            {videos.map((video) => (
              <div
                key={video.id || video.video_id}
                onClick={() => handleVideoClick(video)}
                className="relative aspect-[3/4] bg-zinc-900 group cursor-pointer overflow-hidden hover:opacity-80 transition-opacity"
              >
                {/* Thumbnail */}
                <div className="w-full h-full flex items-center justify-center text-zinc-700 font-bold bg-gradient-to-br from-zinc-800 to-zinc-900">
                  {video.videoThumbnail ? (
                    <img src={video.videoThumbnail} alt={video.title} className="w-full h-full object-cover" />
                  ) : (
                    <img src="/video-placeholder.png" alt={video.title} className="w-full h-full object-cover" />
                  )}
                </div>

                {/* Stats overlay */}
                <div className="absolute bottom-1 left-2 text-xs font-semibold text-white/70 drop-shadow-md">
                  <div className="flex items-center gap-2">
                    <span className="flex items-center gap-1">
                      ▶ {video.interactions?.view_count || 0}
                    </span>
                    <span className="flex items-center gap-1">
                      ❤ {video.interactions?.like_count || 0}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Full-screen Video Viewer Modal */}
      {selectedVideo && selectedVideoUrl && (
        <div className="fixed inset-0 bg-black z-50 flex flex-col">
          {/* Close button */}
          <button
            onClick={closeVideoViewer}
            className="absolute top-4 left-4 z-50 text-white hover:text-gray-300 transition"
          >
            <X size={32} />
          </button>

          {/* Video Container - Resize based on comments visibility */}
          <div className={`relative bg-black transition-all duration-300 ${showComments ? 'h-1/2' : 'h-full'}`}>
            <div className="w-full h-full flex items-center justify-center relative group">
              {selectedVideoUrl ? (
                <video
                  ref={videoRef}
                  src={selectedVideoUrl}
                  muted={isMuted}
                  loop
                  playsInline
                  className="w-full h-full object-contain cursor-pointer"
                  onClick={() => setIsMuted(!isMuted)}
                />
              ) : (
                <div className="text-white text-center">
                  <p>Failed to load video</p>
                </div>
              )}

              {/* Mute button */}
              <button
                onClick={() => setIsMuted(!isMuted)}
                className="absolute bottom-6 right-6 z-40 bg-black/60 hover:bg-black/80 text-white p-3 rounded-full transition"
              >
                {isMuted ? <VolumeX size={24} /> : <Volume2 size={24} />}
              </button>

              {/* Comments toggle button */}
              <button
                onClick={() => setShowComments(!showComments)}
                className="absolute bottom-6 right-20 z-40 bg-black/60 hover:bg-black/80 text-white p-3 rounded-full transition flex items-center gap-2"
              >
                <span>💬</span>
                <span className="text-sm font-semibold">{comments.length}</span>
              </button>

              {/* Video info overlay - bottom left */}
              <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black via-black/50 to-transparent pt-20 pb-6 px-6 z-10">
                <div className="max-w-lg">
                  <h3 className="text-xl font-bold mb-2 line-clamp-1">{selectedVideo.title}</h3>
                  {selectedVideo.description && (
                    <p className="text-gray-200 text-sm mb-2 line-clamp-1">{selectedVideo.description}</p>
                  )}
                  {selectedVideo.hashtags && selectedVideo.hashtags.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                      {selectedVideo.hashtags.slice(0, 2).map((tag, idx) => (
                        <span key={idx} className="text-purple-400 text-sm font-semibold">
                          #{tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Comments Panel - Slide up from bottom */}
          {showComments && (
            <div className="h-1/2 bg-zinc-900 border-t border-zinc-800 flex flex-col overflow-hidden" style={{ animation: 'slideUp 0.3s ease-out' }}>
              {/* Drag handle */}
              <div className="w-full flex justify-center pt-3 pb-2">
                <div className="w-12 h-1 bg-zinc-700 rounded-full"></div>
              </div>

              {/* Comments Header */}
              <div className="px-6 py-4 border-b border-zinc-800">
                <h2 className="text-lg font-bold text-white">Comments</h2>
                <p className="text-sm text-zinc-400">{comments.length} comments</p>
              </div>

              {/* Comments List */}
              <div className="flex-1 overflow-y-auto scrollbar-hide px-6 py-4 space-y-4">
                {comments.length === 0 ? (
                  <div className="text-center text-zinc-500 py-8">
                    <p>No comments yet</p>
                    <p className="text-xs mt-2">Be the first to comment!</p>
                  </div>
                ) : (
                  <>
                    {comments.map((comment) => (
                      <div key={comment.id} className="flex gap-3">
                        {/* Avatar */}
                        <div className="w-10 h-10 rounded-full bg-purple-600 flex items-center justify-center flex-shrink-0 text-white text-sm font-bold">
                          {comment.author.charAt(0).toUpperCase()}
                        </div>

                        {/* Comment Content */}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-white text-sm">@{comment.author}</span>
                            <span className="text-xs text-zinc-500">{comment.timestamp}</span>
                          </div>
                          <p className="text-white text-sm mt-1 break-words">{comment.content}</p>
                        </div>
                      </div>
                    ))}
                    <div ref={commentsEndRef} />
                  </>
                )}
              </div>

              {/* Comment Input */}
              <div className="px-6 py-4 border-t border-zinc-800 bg-zinc-950">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={commentInput}
                    onChange={(e) => setCommentInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
                    placeholder="Add comment..."
                    className="flex-1 bg-zinc-800 text-white text-sm px-4 py-2.5 rounded-full focus:outline-none focus:ring-2 focus:ring-purple-500 placeholder-zinc-500"
                  />
                  <button
                    onClick={handleAddComment}
                    disabled={!commentInput.trim()}
                    className="bg-purple-600 hover:bg-purple-700 disabled:bg-zinc-700 disabled:cursor-not-allowed text-white p-2.5 rounded-full transition flex items-center justify-center flex-shrink-0"
                  >
                    <Send size={20} />
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Error Toast */}
      {error && (
        <div className="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-red-600 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-3 z-50 animate-bounce">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="opacity-70 hover:opacity-100 font-bold">×</button>
        </div>
      )}
    </div>
  );
}
