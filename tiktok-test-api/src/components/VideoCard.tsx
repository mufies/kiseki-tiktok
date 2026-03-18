import { useRef, useEffect, useState } from 'react';
import { Heart, Bookmark, Volume2, VolumeX, AlertCircle, MessageCircle, Share2, X, Send } from 'lucide-react';
import { eventAPI } from '../api/event';
import { interactionAPI } from '../api/interaction';
import { videoAPI } from '../api/video';
import { useAuth } from '../context/AuthContext';
import type { Video } from '../types';

interface VideoCardProps {
  video: Video;
  isActive: boolean;
  onVideoView?: (videoId: string) => void;
}

interface Comment {
  id: string;
  author: string;
  content: string;
  timestamp: string;
}

export default function VideoCard({ video, isActive, onVideoView }: VideoCardProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isLiked, setIsLiked] = useState(video.interactions?.is_liked ?? false);
  const [isBookmarked, setIsBookmarked] = useState(video.interactions?.is_bookmarked ?? false);
  const [isMuted, setIsMuted] = useState(true);
  const [watchPercentage, setWatchPercentage] = useState(0);
  const [videoUrl, setVideoUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [showCommentModal, setShowCommentModal] = useState(false);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentInput, setCommentInput] = useState('');
  const [loadingComments, setLoadingComments] = useState(false);
  const { user } = useAuth();
  const commentsEndRef = useRef<HTMLDivElement>(null);

  // Local state to track counts for real-time updates
  const [likeCount, setLikeCount] = useState(video.interactions?.like_count ?? 0);
  const [bookmarkCount, setBookmarkCount] = useState(video.interactions?.bookmark_count ?? 0);
  const [commentCount, setCommentCount] = useState(video.interactions?.comment_count ?? 0);

  const watchEventSentRef = useRef(false);

  // Fetch video URL with presigned URL
  useEffect(() => {
    const loadVideo = async () => {
      try {
        const videoId = video.id || video.video_id;
        if (!videoId) return;
        const videoData = await videoAPI.getVideo(videoId);
        setVideoUrl(videoData.streamUrl || null);
      } catch (error) {
        console.error('Failed to load video:', error);
      } finally {
        setLoading(false);
      }
    };

    loadVideo();
  }, [video.id, video.video_id]);

  // Auto-play/pause based on visibility
  useEffect(() => {
    const videoElement = videoRef.current;
    if (!videoElement) return;

    if (isActive) {
      try {
        videoElement.play().catch((error) => {
          console.error('Autoplay failed:', error);
        });
      } catch (error) {
        console.error('Play error:', error);
      }
    } else {
      videoElement.pause();
    }
  }, [isActive, videoUrl]);

  // Load comments when modal opens
  useEffect(() => {
    if (!showCommentModal) return;

    const loadComments = async () => {
      try {
        setLoadingComments(true);
        const videoId = video.id || video.video_id;
        if (!videoId) return;

        const fetchedComments = await interactionAPI.getComments(videoId);
        setComments(
          (fetchedComments as any).map((comment: any) => ({
            id: comment.id,
            author: comment.username,
            content: comment.content, timestamp: new Date(comment.createdAt).toLocaleString('en-US', {
              timeStyle: 'short',
              dateStyle: 'short',
            }),
          }))
        );
      } catch (error) {
        console.error('Failed to load comments:', error);
      } finally {
        setLoadingComments(false);
      }
    };

    loadComments();
  }, [showCommentModal, video.id, video.video_id]);

  // Track watch percentage
  useEffect(() => {
    const videoElement = videoRef.current;
    if (!videoElement) return;

    const handleTimeUpdate = () => {
      const percent = (videoElement.currentTime / videoElement.duration) * 100;
      setWatchPercentage(percent);
    };

    videoElement.addEventListener('timeupdate', handleTimeUpdate);
    return () => videoElement.removeEventListener('timeupdate', handleTimeUpdate);
  }, [videoUrl]);

  // Send watch event when user scrolls away or video ends
  useEffect(() => {
    return () => {
      // Send watch event on unmount (when scrolling away)
      if (!watchEventSentRef.current && watchPercentage > 0 && user) {
        sendWatchEvent();
        watchEventSentRef.current = true;
      }
    };
  }, []);

  const sendWatchEvent = async () => {
    if (!user) return;

    try {
      const videoId = video.id || video.video_id;
      if (!videoId) return;
      const userId = user.id || user.user_id;
      if (!userId) return;

      await eventAPI.sendWatchEvent(
        userId,
        videoId,
        Math.round(watchPercentage),
        isLiked
      );
      console.log(`Watch event sent: ${Math.round(watchPercentage)}%`);
      if (onVideoView) {
        onVideoView(videoId);
      }
    } catch (error) {
      console.error('Failed to send watch event:', error);
    }
  };

  const handleLike = async () => {
    try {
      const videoId = video.id || video.video_id;
      if (!videoId) return;

      // Optimistic update
      const newIsLiked = !isLiked;
      setIsLiked(newIsLiked);
      setLikeCount((prev) => (newIsLiked ? prev + 1 : prev - 1));

      await interactionAPI.toggleLike(videoId);
    } catch (error) {
      console.error('Failed to like:', error);
      // Revert on error
      setIsLiked(!isLiked);
      setLikeCount((prev) => (isLiked ? prev + 1 : prev - 1));
    }
  };

  const handleBookmark = async () => {
    try {
      const videoId = video.id || video.video_id;
      if (!videoId) return;

      // Optimistic update
      const newIsBookmarked = !isBookmarked;
      setIsBookmarked(newIsBookmarked);
      setBookmarkCount((prev) => (newIsBookmarked ? prev + 1 : prev - 1));

      await interactionAPI.toggleBookmark(videoId);
    } catch (error) {
      console.error('Failed to bookmark:', error);
      // Revert on error
      setIsBookmarked(!isBookmarked);
      setBookmarkCount((prev) => (isBookmarked ? prev + 1 : prev - 1));
    }
  };

  const handleVideoEnd = () => {
    if (!watchEventSentRef.current && user) {
      sendWatchEvent();
      watchEventSentRef.current = true;
    }
  };

  const handleAddComment = async () => {
    if (!commentInput.trim() || !user) return;

    try {
      const videoId = video.id || video.video_id;
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
      setCommentCount((prev) => prev + 1);

      // Call API
      const apiResponse = await interactionAPI.addComment(videoId, commentInput);
      console.log('Comment added:', apiResponse);

      // Auto-scroll
      setTimeout(() => {
        commentsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    } catch (error) {
      console.error('Failed to add comment:', error);
      // Revert on error
      setComments(comments.slice(0, -1));
      setCommentCount((prev) => (prev > 0 ? prev - 1 : 0));
    }
  };

  if (loading) {
    return (
      <div className="relative w-full h-full bg-black flex items-center justify-center">
        <div className="text-white text-center animate-pulse">
          <div className="w-10 h-10 border-4 border-gray-600 border-t-white rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-sm text-gray-400 font-medium">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="relative w-full h-full bg-black flex items-center justify-center group">
      {/* Mute indicator */}
      {isActive && (
        <div className="absolute top-4 left-1/2 transform -translate-x-1/2 z-20 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
          <span className="bg-black/70 text-white text-xs px-3 py-1 rounded-full">
            {isMuted ? '🔇 Muted' : '🔊 Unmuted'}
          </span>
        </div>
      )}

      {/* Main video container */}
      <div className="relative w-full h-full">
        {videoUrl ? (
          <video
            ref={videoRef}
            src={videoUrl}
            muted={isMuted}
            loop
            playsInline
            onEnded={handleVideoEnd}
            onClick={() => setIsMuted(!isMuted)}
            className="w-full h-full object-contain cursor-pointer"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-black">
            <div className="text-gray-400 flex flex-col items-center">
              <AlertCircle size={40} className="mb-2" />
              <p>Failed to load video</p>
            </div>
          </div>
        )}
      </div>

      {/* Right sidebar with action buttons */}
      <div className="absolute right-4 bottom-24 z-20 flex flex-col gap-4">
        {/* Like button */}
        <button
          onClick={handleLike}
          className="flex flex-col items-center gap-1 text-white hover:scale-110 transition-transform"
        >
          <div className="w-12 h-12 rounded-full flex flex-col items-center justify-center drop-shadow-md">
            <Heart size={34} fill={isLiked ? '#ef4444' : 'transparent'} color={isLiked ? '#ef4444' : 'white'} />
          </div>
          <span className="text-xs font-semibold drop-shadow-md">
            {likeCount > 999 ? Math.floor(likeCount / 1000) + 'k' : likeCount}
          </span>
        </button>

        {/* Comment button */}
        <button
          onClick={() => setShowCommentModal(true)}
          className="flex flex-col items-center gap-1 text-white hover:scale-110 transition-transform"
        >
          <div className="w-12 h-12 rounded-full flex flex-col items-center justify-center drop-shadow-md">
            <MessageCircle size={32} color="white" fill="white" className="scale-x-[-1]" />
          </div>
          <span className="text-xs font-semibold drop-shadow-md">
            {commentCount > 999 ? Math.floor(commentCount / 1000) + 'k' : commentCount}
          </span>
        </button>

        {/* Bookmark button */}
        <button
          onClick={handleBookmark}
          className="flex flex-col items-center gap-1 text-white hover:scale-110 transition-transform"
        >
          <div className="w-12 h-12 rounded-full flex flex-col items-center justify-center drop-shadow-md">
            <Bookmark size={30} fill={isBookmarked ? '#eab308' : 'transparent'} color={isBookmarked ? '#eab308' : 'white'} />
          </div>
          <span className="text-xs font-semibold drop-shadow-md">
            {bookmarkCount > 999 ? Math.floor(bookmarkCount / 1000) + 'k' : bookmarkCount}
          </span>
        </button>

        {/* Share button */}
        <button
          className="flex flex-col items-center gap-1 text-white hover:scale-110 transition-transform"
        >
          <div className="w-12 h-12 rounded-full flex flex-col items-center justify-center drop-shadow-md">
            <Share2 size={32} color="white" fill="white" />
          </div>
          <span className="text-xs font-semibold drop-shadow-md">Share</span>
        </button>

        {/* Mute button */}
        <button
          onClick={() => setIsMuted(!isMuted)}
          className="flex flex-col items-center gap-1 text-white hover:scale-110 transition-transform mt-2"
        >
          <div className="w-12 h-12 rounded-full bg-black/40 backdrop-blur-sm flex items-center justify-center transition border border-white/20">
            {isMuted ? <VolumeX size={20} /> : <Volume2 size={20} />}
          </div>
        </button>
      </div>

      {/* Bottom overlay with video metadata */}
      <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent pt-24 pb-4 px-4 z-10 pointer-events-none">
        <div className="text-white max-w-[80%] pointer-events-auto">
          {/* User handling / Title */}
          <h3 className="text-base font-bold mb-1 line-clamp-1">@{video.ownerId || 'user'}</h3>

          {/* Description */}
          <p className="text-sm text-gray-100 mb-2 line-clamp-2">{video.description || video.title}</p>

          {/* Hashtags */}
          {video.hashtags && video.hashtags.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-2">
              {(Array.isArray(video.hashtags) ? video.hashtags : []).map((tag, idx) => (
                <span key={idx} className="text-sm font-semibold hover:underline cursor-pointer">
                  #{tag}
                </span>
              ))}
            </div>
          )}

          {/* Music/Sound track ticker (fake for now to emulate tiktok) */}
          <div className="flex items-center gap-2 mb-2">
            <div className="animate-pulse">🎵</div>
            <span className="text-sm">Original sound - {video.title}</span>
          </div>

          {/* Interaction stats */}
          {video.interactions && (
            <div className="flex items-center gap-4 text-xs text-white/70 mb-2">
              {video.interactions.view_count !== undefined && (
                <span>👁️ {video.interactions.view_count > 999 ? Math.floor(video.interactions.view_count / 1000) + 'k' : video.interactions.view_count} views</span>
              )}
              {video.interactions.like_count !== undefined && (
                <span>❤️ {video.interactions.like_count} likes</span>
              )}
            </div>
          )}

          {/* Watch progress */}
          <div className="text-xs text-white/50 py-1">
            Watched: {Math.round(watchPercentage)}%
          </div>
        </div>
      </div>

      {/* Comment Modal */}
      {showCommentModal && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-end">
          {/* Close on background click */}
          <div className="absolute inset-0" onClick={() => setShowCommentModal(false)} />

          {/* Modal Content - Slide up from bottom */}
          <div className="relative w-full bg-zinc-900 rounded-t-3xl max-h-[80vh] flex flex-col" style={{ animation: 'slideUp 0.3s ease-out' }}>
            {/* Drag handle */}
            <div className="w-full flex justify-center pt-3 pb-2">
              <div className="w-12 h-1 bg-zinc-700 rounded-full"></div>
            </div>

            {/* Header */}
            <div className="px-6 py-4 border-b border-zinc-800 flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-white">Comments</h2>
                <p className="text-sm text-zinc-400">{comments.length} comments</p>
              </div>
              <button onClick={() => setShowCommentModal(false)} className="text-zinc-400 hover:text-white transition">
                <X size={24} />
              </button>
            </div>

            {/* Comments List */}
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4 scrollbar-hide">
              {loadingComments ? (
                <div className="text-center text-zinc-500 py-8">
                  <div className="w-6 h-6 border-3 border-zinc-600 border-t-purple-500 rounded-full animate-spin mx-auto mb-3"></div>
                  <p>Loading comments...</p>
                </div>
              ) : comments.length === 0 ? (
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
                  autoFocus
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
        </div>
      )}
    </div>
  );
}
