import { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';
import { Settings } from 'lucide-react';

interface StreamPlayerProps {
  hlsUrl: string;
  poster?: string;
  autoplay?: boolean;
}

export default function StreamPlayer({ hlsUrl, poster, autoplay = true }: StreamPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [showQualityMenu, setShowQualityMenu] = useState(false);
  const [currentQuality, setCurrentQuality] = useState<number>(-1);
  const [qualities, setQualities] = useState<Array<{ level: number; height: number; bitrate: number }>>([]);

  useEffect(() => {
    if (!videoRef.current || !hlsUrl) return;

    // Validate HLS URL
    if (!hlsUrl.startsWith('http')) {
      setError('Invalid stream URL. Expected HTTP/HTTPS URL.');
      setLoading(false);
      return;
    }

    const video = videoRef.current;

    // Check if HLS is natively supported (Safari)
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = hlsUrl;
      video.addEventListener('loadedmetadata', () => setLoading(false));
      if (autoplay) {
        video.play().catch((err) => {
          console.error('Autoplay failed:', err);
          setError('Click to play');
        });
      }
    }
    // Use hls.js for other browsers
    else if (Hls.isSupported()) {
      const hls = new Hls({
        enableWorker: true,
        lowLatencyMode: true,
      });

      hlsRef.current = hls;

      hls.loadSource(hlsUrl);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        setLoading(false);

        // Get available quality levels
        if (hls.levels && hls.levels.length > 1) {
          const levelData = hls.levels.map((level, index) => ({
            level: index,
            height: level.height,
            bitrate: level.bitrate,
          }));
          setQualities(levelData);
          setCurrentQuality(hls.currentLevel);
        }

        if (autoplay) {
          video.play().catch((err) => {
            console.error('Autoplay failed:', err);
            setError('Click to play');
          });
        }
      });

      hls.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
        setCurrentQuality(data.level);
      });

      hls.on(Hls.Events.ERROR, (_, data) => {
        console.error('[StreamPlayer] HLS.js error:', data);

        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              console.error('[StreamPlayer] Network error details:', data.details);

              // Check if it's a 404 (file not found)
              if (data.details === 'manifestLoadError' || data.response?.code === 404) {
                setError(
                  'Stream not ready yet. The HLS files are still being generated. ' +
                  'Wait a few seconds and try refreshing, or check if OBS is still streaming.'
                );
                setLoading(false);
              } else {
                setError('Network error - attempting to recover');
                hls.startLoad();
              }
              break;

            case Hls.ErrorTypes.MEDIA_ERROR:
              console.error('[StreamPlayer] Media error details:', data.details);
              setError('Media error - attempting to recover');
              hls.recoverMediaError();
              break;

            default:
              console.error('[StreamPlayer] Fatal error:', data.details);
              setError('Fatal error - cannot play stream');
              hls.destroy();
              break;
          }
        }
      });
    } else {
      setError('HLS is not supported in this browser');
      setLoading(false);
    }

    return () => {
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [hlsUrl, autoplay]);

  const handleClick = () => {
    if (videoRef.current && error === 'Click to play') {
      videoRef.current.play();
      setError(null);
    }
  };

  const handleQualityChange = (level: number) => {
    if (hlsRef.current) {
      hlsRef.current.currentLevel = level;
      setCurrentQuality(level);
      setShowQualityMenu(false);
    }
  };

  const getQualityLabel = (quality: { level: number; height: number; bitrate: number }) => {
    if (quality.height >= 1080) return '1080p';
    if (quality.height >= 720) return '720p';
    if (quality.height >= 480) return '480p';
    if (quality.height >= 360) return '360p';
    return `${quality.height}p`;
  };

  return (
    <div className="relative w-full aspect-video bg-black rounded-lg overflow-hidden group">
      <video
        ref={videoRef}
        className="w-full h-full"
        controls
        poster={poster}
        onClick={handleClick}
        playsInline
      />

      {/* Quality Selector */}
      {qualities.length > 1 && (
        <div className="absolute top-4 right-4 z-10">
          <button
            onClick={() => setShowQualityMenu(!showQualityMenu)}
            className="p-2 bg-black/75 hover:bg-black/90 rounded-lg transition-colors backdrop-blur-sm"
            title="Quality Settings"
          >
            <Settings className="w-5 h-5 text-white" />
          </button>

          {showQualityMenu && (
            <div className="absolute right-0 mt-2 w-40 bg-gray-900 rounded-lg shadow-xl border border-gray-700 overflow-hidden">
              <div className="px-3 py-2 border-b border-gray-700">
                <p className="text-white text-xs font-semibold">Video Quality</p>
              </div>
              <div className="py-1">
                <button
                  onClick={() => handleQualityChange(-1)}
                  className={`w-full px-3 py-2 text-left text-sm transition-colors ${
                    currentQuality === -1
                      ? 'bg-purple-600 text-white'
                      : 'text-gray-300 hover:bg-gray-800'
                  }`}
                >
                  Auto {currentQuality === -1 && '✓'}
                </button>
                {qualities.map((quality) => (
                  <button
                    key={quality.level}
                    onClick={() => handleQualityChange(quality.level)}
                    className={`w-full px-3 py-2 text-left text-sm transition-colors ${
                      currentQuality === quality.level
                        ? 'bg-purple-600 text-white'
                        : 'text-gray-300 hover:bg-gray-800'
                    }`}
                  >
                    {getQualityLabel(quality)} {currentQuality === quality.level && '✓'}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {loading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-white"></div>
        </div>
      )}

      {error && error !== 'Click to play' && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-75">
          <div className="text-white text-center px-4">
            <p className="text-lg font-semibold mb-2">Playback Error</p>
            <p className="text-sm opacity-75">{error}</p>
          </div>
        </div>
      )}

      {error === 'Click to play' && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 cursor-pointer">
          <div className="w-16 h-16 rounded-full bg-white bg-opacity-90 flex items-center justify-center">
            <div className="w-0 h-0 border-t-8 border-t-transparent border-l-12 border-l-black border-b-8 border-b-transparent ml-1"></div>
          </div>
        </div>
      )}
    </div>
  );
}
