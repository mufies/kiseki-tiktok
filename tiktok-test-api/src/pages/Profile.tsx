import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { userAPI, type UserProfile } from '../api/user';
import { videoAPI } from '../api/video';
import { interactionAPI, type LikedVideo } from '../api/interaction';
import { type Video } from '../types';
import { Settings, Bookmark, Heart, Grid, LayoutGrid, ChevronLeft, X, Volume2, VolumeX, Send } from 'lucide-react';

interface Comment {
  id: string;
  author: string;
  text: string;
  timestamp: string;
}


export default function Profile() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [videos, setVideos] = useState<Video[]>([]);
  const [likedVideos, setLikedVideos] = useState<LikedVideo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({ username: '', avatarUrl: '' });
  const [activeTab, setActiveTab] = useState<'videos' | 'liked'>('videos');
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null);
  const [selectedVideoUrl, setSelectedVideoUrl] = useState<string | null>(null);
  const [isMuted, setIsMuted] = useState(true);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentInput, setCommentInput] = useState('');
  const [showComments, setShowComments] = useState(false);
  const [currentLikedIndex, setCurrentLikedIndex] = useState<number>(-1);
  const videoRef = useRef<HTMLVideoElement>(null);
  const commentsEndRef = useRef<HTMLDivElement>(null);
  const videoContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadProfile();
  }, [user]);

  useEffect(() => {
    if (activeTab === 'liked' && profile) {
      loadLikedVideos();
    }
  }, [activeTab, profile]);

  const loadProfile = async () => {
    if (!user) return;

    try {
      setLoading(true);
      setError(null);

      const userProfile = await userAPI.getCurrentProfile();
      setProfile(userProfile);
      setEditForm({
        username: userProfile.username || '',
        avatarUrl: userProfile.avatarUrl || '',
      });

      const userVideos = await videoAPI.getUserVideos(userProfile.id);
      setVideos(userVideos);
    } catch (err) {
      console.error('Failed to load profile:', err);
      setError('Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  const loadLikedVideos = async () => {
    if (!profile) return;

    try {
      setError(null);
      const liked = await interactionAPI.getUserLikes(profile.id);
      setLikedVideos(liked);
    } catch (err) {
      console.error('Failed to load liked videos:', err);
      setError('Failed to load liked videos');
    }
  };

  const handleUpdateProfile = async () => {
    try {
      const updatedProfile = await userAPI.updateProfile(editForm);
      setProfile(updatedProfile);
      setIsEditing(false);
    } catch (err) {
      console.error('Failed to update profile:', err);
      setError('Failed to update profile');
    }
  };

  const handleDeleteVideo = async (videoId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm('Are you sure you want to delete this video?')) return;

    try {
      await videoAPI.deleteVideo(videoId);
      setVideos(videos.filter(v => (v.id || v.video_id) !== videoId));
    } catch (err) {
      console.error('Failed to delete video:', err);
      setError('Failed to delete video');
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Failed to logout:', error);
    }
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
      // Mock comments for demo (in real app, these would come from API)
      setComments([
        {
          id: '1',
          author: 'user1',
          text: 'Great video! 🔥',
          timestamp: '2 hours ago',
        },
        {
          id: '2',
          author: 'user2',
          text: 'Love this content',
          timestamp: '1 hour ago',
        },
        {
          id: '3',
          author: 'user3',
          text: 'Amazing! Please share more',
          timestamp: '30 minutes ago',
        },
      ]);
    } catch (err) {
      console.error('Failed to load video:', err);
      setError('Failed to load video');
    }
  };

  const handleLikedVideoClick = async (likedVideo: LikedVideo, index: number) => {
    if (!likedVideo.available) {
      setError('This video is no longer available');
      return;
    }

    try {
      const videoData = await videoAPI.getVideo(likedVideo.videoId);
      // Convert LikedVideo to Video format for display
      const video: Video = {
        id: likedVideo.videoId,
        video_id: likedVideo.videoId,
        title: likedVideo.title,
        hashtags: likedVideo.hashtags,
        description: '',
      };
      setSelectedVideo(video);
      setSelectedVideoUrl(videoData.streamUrl || null);
      setCurrentLikedIndex(index);
      setIsMuted(true);
      setShowComments(false);
      setCommentInput('');
      setComments([
        {
          id: '1',
          author: 'user1',
          text: 'Great video! 🔥',
          timestamp: '2 hours ago',
        },
        {
          id: '2',
          author: 'user2',
          text: 'Love this content',
          timestamp: '1 hour ago',
        },
        {
          id: '3',
          author: 'user3',
          text: 'Amazing! Please share more',
          timestamp: '30 minutes ago',
        },
      ]);
    } catch (err) {
      console.error('Failed to load liked video:', err);
      setError('Failed to load video');
    }
  };

  const navigateToNextLikedVideo = async () => {
    const availableVideos = likedVideos.filter(v => v.available);
    if (currentLikedIndex >= 0 && currentLikedIndex < availableVideos.length - 1) {
      const nextVideo = availableVideos[currentLikedIndex + 1];
      await handleLikedVideoClick(nextVideo, currentLikedIndex + 1);
    }
  };

  const navigateToPrevLikedVideo = async () => {
    const availableVideos = likedVideos.filter(v => v.available);
    if (currentLikedIndex > 0) {
      const prevVideo = availableVideos[currentLikedIndex - 1];
      await handleLikedVideoClick(prevVideo, currentLikedIndex - 1);
    }
  };

  const closeVideoViewer = () => {
    setSelectedVideo(null);
    setSelectedVideoUrl(null);
    setShowComments(false);
    setCurrentLikedIndex(-1);
    if (videoRef.current) {
      videoRef.current.pause();
    }
  };

  // Handle wheel scroll for navigation
  const handleWheel = (e: React.WheelEvent) => {
    if (currentLikedIndex < 0 || showComments) return;

    if (e.deltaY > 0) {
      // Scroll down - next video
      navigateToNextLikedVideo();
    } else if (e.deltaY < 0) {
      // Scroll up - previous video
      navigateToPrevLikedVideo();
    }
  };

  // Touch handling for mobile swipe
  const [touchStart, setTouchStart] = useState<number>(0);
  const [touchEnd, setTouchEnd] = useState<number>(0);

  const handleTouchStart = (e: React.TouchEvent) => {
    setTouchStart(e.targetTouches[0].clientY);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    setTouchEnd(e.targetTouches[0].clientY);
  };

  const handleTouchEnd = () => {
    if (currentLikedIndex < 0 || showComments) return;
    if (!touchStart || !touchEnd) return;

    const distance = touchStart - touchEnd;
    const isSwipe = Math.abs(distance) > 50;

    if (isSwipe) {
      if (distance > 0) {
        // Swipe up - next video
        navigateToNextLikedVideo();
      } else {
        // Swipe down - previous video
        navigateToPrevLikedVideo();
      }
    }

    setTouchStart(0);
    setTouchEnd(0);
  };

  const handleAddComment = () => {
    if (!commentInput.trim() || !user) return;

    const newComment: Comment = {
      id: Date.now().toString(),
      author: user.username,
      text: commentInput,
      timestamp: 'just now',
    };

    setComments([...comments, newComment]);
    setCommentInput('');

    // Auto-scroll to bottom
    setTimeout(() => {
      commentsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, 100);
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
        <p className="text-xl mb-4 text-red-400">{error || 'Profile not found'}</p>
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
        <button onClick={() => navigate('/')} className="text-white hover:text-zinc-300 transition">
          <ChevronLeft size={28} />
        </button>
        <h1 className="font-bold text-lg">{profile.username || 'Profile'}</h1>
        <button onClick={handleLogout} className="text-zinc-400 hover:text-white transition" title="Logout">
          <Settings size={22} />
        </button>
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

        {/* Edit Profile Logic */}
        {isEditing ? (
          <div className="w-full bg-zinc-900 border border-zinc-800 p-4 rounded-xl mb-4">
            <div className="flex flex-col gap-3 mb-4">
              <input
                type="text"
                placeholder="Username"
                className="bg-black border border-zinc-800 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-purple-500"
                value={editForm.username}
                onChange={e => setEditForm({ ...editForm, username: e.target.value })}
              />
              <input
                type="text"
                placeholder="Avatar Image URL (Optional)"
                className="bg-black border border-zinc-800 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-purple-500"
                value={editForm.avatarUrl}
                onChange={e => setEditForm({ ...editForm, avatarUrl: e.target.value })}
              />
            </div>
            <div className="flex gap-2">
              <button
                className="flex-1 bg-red-500 hover:bg-red-600 text-white py-2 rounded font-semibold text-sm transition"
                onClick={handleUpdateProfile}
              >
                Save
              </button>
              <button
                className="flex-1 bg-zinc-800 hover:bg-zinc-700 text-white py-2 rounded font-semibold text-sm transition"
                onClick={() => setIsEditing(false)}
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <div className="flex gap-2 w-full max-w-xs justify-center">
            <button
              className="bg-zinc-800 hover:bg-zinc-700 px-8 py-2 rounded-md font-semibold text-sm transition"
              onClick={() => setIsEditing(true)}
            >
              Edit profile
            </button>
            <button
              className="bg-zinc-800 hover:bg-zinc-700 px-4 py-2 rounded-md transition"
              title="Bookmarks"
            >
              <Bookmark size={18} />
            </button>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="flex border-b border-zinc-800 w-full mb-1">
        <button
          className={`flex-1 flex justify-center py-3 border-b-2 transition ${activeTab === 'videos' ? 'border-white text-white' : 'border-transparent text-zinc-500 hover:text-white'}`}
          onClick={() => setActiveTab('videos')}
        >
          <LayoutGrid size={24} />
        </button>
        <button
          className={`flex-1 flex justify-center py-3 border-b-2 transition ${activeTab === 'liked' ? 'border-white text-white' : 'border-transparent text-zinc-500 hover:text-white'}`}
          onClick={() => setActiveTab('liked')}
        >
          <Heart size={24} />
        </button>
      </div>

      {/* Video Grid */}
      <div className="max-w-4xl mx-auto w-full">
        {activeTab === 'videos' ? (
          videos.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-zinc-500">
              <Grid size={48} className="mb-4 opacity-50" />
              <p>No videos yet</p>
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-0.5 sm:gap-1">
              {videos.map(video => (
                <div
                  key={video.id || video.video_id}
                  onClick={() => handleVideoClick(video)}
                  className="relative aspect-[3/4] bg-zinc-900 group cursor-pointer overflow-hidden hover:opacity-80 transition-opacity"
                >
                  {/* Placeholder/thumbnail */}
                  <div className="w-full h-full flex items-center justify-center text-zinc-700 font-bold bg-gradient-to-br from-zinc-800 to-zinc-900">
                    <span className="truncate w-3/4 text-center text-xs opacity-50">{video.title}</span>
                  </div>

                  {/* Views/Plays overlay */}
                  <div className="absolute bottom-1 left-2 text-xs font-semibold flex items-center gap-1 drop-shadow-md text-white/70">
                    <span>▶</span> {Math.floor(Math.random() * 5000) + 100}
                  </div>

                  {/* Delete button (visible on hover) */}
                  <button
                    className="absolute top-2 right-2 bg-black/60 hover:bg-red-600 p-1.5 rounded text-white opacity-0 group-hover:opacity-100 transition z-10"
                    onClick={(e) => handleDeleteVideo(video.id || video.video_id || '', e)}
                    title="Delete Video"
                  >
                    🗑️
                  </button>
                </div>
              ))}
            </div>
          )
        ) : (
          (() => {
            const availableVideos = likedVideos.filter(v => v.available);
            return availableVideos.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 text-zinc-500">
                <Heart size={48} className="mb-4 opacity-50" />
                <p>No liked videos yet</p>
              </div>
            ) : (
              <div className="grid grid-cols-3 gap-0.5 sm:gap-1">
                {availableVideos.map((likedVideo, index) => (
                  <div
                    key={likedVideo.interactionId}
                    onClick={() => handleLikedVideoClick(likedVideo, index)}
                    className="relative aspect-[3/4] bg-zinc-900 group cursor-pointer overflow-hidden hover:opacity-80 transition-opacity"
                  >
                    {/* Placeholder/thumbnail */}
                    <div className="w-full h-full flex items-center justify-center text-zinc-700 font-bold bg-gradient-to-br from-zinc-800 to-zinc-900">
                      <span className="truncate w-3/4 text-center text-xs opacity-50">
                        {likedVideo.title}
                      </span>
                    </div>

                    {/* Hashtags overlay */}
                    {likedVideo.hashtags.length > 0 && (
                      <div className="absolute top-2 left-2 text-xs font-semibold text-purple-400 drop-shadow-md">
                        #{likedVideo.hashtags[0]}
                      </div>
                    )}

                    {/* Liked date overlay */}
                    <div className="absolute bottom-1 left-2 text-xs font-semibold flex items-center gap-1 drop-shadow-md text-white/70">
                      <Heart size={12} fill="currentColor" />
                      <span>{new Date(likedVideo.likedAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                ))}
              </div>
            );
          })()
        )}
      </div>

      {/* Full-screen Video Viewer Modal */}
      {selectedVideo && selectedVideoUrl && (
        <div
          ref={videoContainerRef}
          className="fixed inset-0 bg-black z-50 flex flex-col"
          onWheel={handleWheel}
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        >
          {/* Close button */}
          <button
            onClick={closeVideoViewer}
            className="absolute top-4 left-4 z-50 text-white hover:text-gray-300 transition"
          >
            <X size={32} />
          </button>

          {/* Navigation indicators */}
          {currentLikedIndex >= 0 && (
            <div className="absolute top-4 right-4 z-50 text-white text-sm bg-black/60 px-3 py-1 rounded-full">
              {currentLikedIndex + 1} / {likedVideos.filter(v => v.available).length}
            </div>
          )}

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
              <div className="w-full flex justify-center pt-2 pb-1">
                <div className="w-12 h-1 bg-zinc-700 rounded-full"></div>
              </div>

              {/* Comments Header */}
              <div className="px-4 py-3 border-b border-zinc-800">
                <h2 className="text-lg font-bold text-white">Comments</h2>
                <p className="text-sm text-zinc-400">{comments.length} comments</p>
              </div>

              {/* Comments List */}
              <div className="flex-1 overflow-y-auto scrollbar-hide px-4 py-3 space-y-3">
                {comments.length === 0 ? (
                  <div className="text-center text-zinc-500 py-6">
                    <p>No comments yet</p>
                    <p className="text-xs mt-2">Be the first to comment!</p>
                  </div>
                ) : (
                  <>
                    {comments.map((comment) => (
                      <div key={comment.id} className="flex gap-2">
                        {/* Avatar */}
                        <div className="w-8 h-8 rounded-full bg-purple-600 flex items-center justify-center flex-shrink-0 text-white text-xs font-bold">
                          {comment.author.charAt(0).toUpperCase()}
                        </div>

                        {/* Comment Content */}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-white text-sm">@{comment.author}</span>
                            <span className="text-xs text-zinc-500">{comment.timestamp}</span>
                          </div>
                          <p className="text-white text-sm mt-0.5 break-words line-clamp-2">{comment.text}</p>
                        </div>
                      </div>
                    ))}
                    <div ref={commentsEndRef} />
                  </>
                )}
              </div>

              {/* Comment Input */}
              <div className="px-4 py-3 border-t border-zinc-800 bg-zinc-950">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={commentInput}
                    onChange={(e) => setCommentInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
                    placeholder="Add comment..."
                    className="flex-1 bg-zinc-800 text-white text-sm px-3 py-2 rounded-full focus:outline-none focus:ring-2 focus:ring-purple-500 placeholder-zinc-500"
                  />
                  <button
                    onClick={handleAddComment}
                    disabled={!commentInput.trim()}
                    className="bg-purple-600 hover:bg-purple-700 disabled:bg-zinc-700 disabled:cursor-not-allowed text-white p-2 rounded-full transition flex items-center justify-center flex-shrink-0"
                  >
                    <Send size={18} />
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
