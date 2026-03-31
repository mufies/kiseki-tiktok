import { useState, useEffect, useRef } from 'react';
import { feedAPI } from '../api/feed';
import { useAuth } from '../context/AuthContext';
import type { Video } from '../types';
import VideoCard from './VideoCard';
import { interactionAPI } from '../api/interaction';

type FeedType = 'personalized' | 'trending';

export default function VideoFeed() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const feedType: FeedType = 'personalized';
  const [activeVideoIndex, setActiveVideoIndex] = useState(0);
  const { user } = useAuth();
  const videoRefs = useRef<(HTMLDivElement | null)[]>([]);
  const initialLoadDone = useRef(false);

  useEffect(() => {
    if (!user) return;

    // Only load once on mount
    if (!initialLoadDone.current) {
      initialLoadDone.current = true;
      loadFeed();
    }
  }, [user]);

  useEffect(() => {
    // Setup Intersection Observer for auto-play detection
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const index = Number(entry.target.getAttribute('data-index'));
            setActiveVideoIndex(index);
          }
        });
      },
      { threshold: 0.7 }
    );

    // Observe all video cards
    videoRefs.current.forEach((ref) => {
      if (ref) observer.observe(ref);
    });

    return () => {
      videoRefs.current.forEach((ref) => {
        if (ref) observer.unobserve(ref);
      });
    };
  }, [videos]);

  const loadFeed = async () => {
    if (!user) return;

    setLoading(true);
    setError(null);

    try {
      let feedData;
      const userId = user.id || user.user_id;
      if (!userId) {
        setError('User ID not found');
        return;
      }

      if (feedType === 'personalized') {
        const response = await feedAPI.getFeed(userId, 20);
        feedData = response.videos || response;
      } else {
        feedData = await feedAPI.getTrending(20);
      }

      // Ensure videos is an array and handle field mapping (video_id -> id)
      const videosList = Array.isArray(feedData) ? feedData : [];
      const mappedVideos = videosList.map((v: any) => ({
        ...v,
        id: v.id || v.video_id, // Handle both id and video_id from feed
      }));
      setVideos(mappedVideos);
    } catch (err) {
      console.error('Failed to load feed:', err);
      setError('Failed to load feed');
    } finally {
      setLoading(false);
    }
  };

  const handleVideoView = (videoId: string) => {
    interactionAPI.recordView(videoId).catch((err) => { console.error('Failed to record view:', err) });
  };

  if (!user) {
    return (
      <div className="h-screen flex items-center justify-center bg-black">
        <div className="text-white text-xl">Please login to view feed</div>
      </div>
    );
  }

  return (
    <div className="relative h-screen bg-black overflow-hidden">
      {/* Vertical scroll container with snap */}
      <div
        className="h-full overflow-y-scroll snap-y snap-mandatory scroll-smooth scrollbar-hide"
        style={{
          scrollBehavior: 'smooth',
          scrollSnapType: 'y mandatory',
        }}
      >
        {loading && (
          <div className="h-screen flex items-center justify-center bg-black">
            <div className="text-white text-center">
              <div className="text-3xl mb-3">🎬</div>
              <p>Loading videos...</p>
            </div>
          </div>
        )}

        {error && (
          <div className="h-screen flex items-center justify-center bg-black">
            <div className="text-white text-center">
              <p className="text-red-400 mb-4">{error}</p>
              <button
                onClick={loadFeed}
                className="bg-purple-600 hover:bg-purple-700 text-white font-semibold px-6 py-2 rounded-lg transition"
              >
                Retry
              </button>
            </div>
          </div>
        )}

        {!loading && !error && videos.length === 0 && (
          <div className="h-screen flex items-center justify-center bg-black">
            <div className="text-white text-center">
              <div className="text-3xl mb-3">📹</div>
              <p className="text-gray-400">No videos available</p>
            </div>
          </div>
        )}

        {!loading && !error && videos.length > 0 && (
          <>
            {videos.map((video, index) => (
              <div
                key={video.id || video.video_id || index}
                ref={(el) => {
                  videoRefs.current[index] = el;
                }}
                data-index={index}
                className="h-screen snap-start snap-always"
                style={{ scrollSnapStop: 'always' }}
              >
                <VideoCard
                  video={video}
                  isActive={index === activeVideoIndex}
                  onVideoView={handleVideoView}
                />
              </div>
            ))}
          </>
        )}
      </div>

      {/* Refresh button - fixed at bottom */}
      {!loading && (
        <button
          onClick={loadFeed}
          className="absolute bottom-6 left-1/2 transform -translate-x-1/2 bg-purple-600 hover:bg-purple-700 text-white font-semibold px-4 py-2 rounded-full transition z-10 text-sm"
        >
          🔄 Refresh
        </button>
      )}
    </div>
  );
}
