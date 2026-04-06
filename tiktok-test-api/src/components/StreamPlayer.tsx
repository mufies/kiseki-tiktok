import { useEffect, useRef, useState, useCallback } from 'react';
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
  const retryCountRef = useRef(0);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const mountedRef = useRef(true);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [retryMessage, setRetryMessage] = useState<string | null>(null);
  const [showQualityMenu, setShowQualityMenu] = useState(false);
  const [currentQuality, setCurrentQuality] = useState<number>(-1);
  const [qualities, setQualities] = useState<Array<{ level: number; height: number; bitrate: number }>>([]);

  const MAX_RETRIES = 30; // Retry for up to 60 seconds (30 * 2s)
  const RETRY_DELAY = 2000; // 2 seconds between retries

  const initHls = useCallback((video: HTMLVideoElement, url: string) => {
    console.log('[StreamPlayer] initHls called with URL:', url);

    // Cleanup existing HLS instance
    if (hlsRef.current) {
      console.log('[StreamPlayer] Destroying existing HLS instance');
      hlsRef.current.destroy();
      hlsRef.current = null;
    }

    const hls = new Hls({
      debug: false,
      enableWorker: true,
      lowLatencyMode: false,
      backBufferLength: 90,
      maxBufferLength: 30,
      maxMaxBufferLength: 60,
      manifestLoadingTimeOut: 10000,
      manifestLoadingMaxRetry: 2,
      manifestLoadingRetryDelay: 500,
      levelLoadingTimeOut: 10000,
      levelLoadingMaxRetry: 3,
      levelLoadingRetryDelay: 1000,
      fragLoadingTimeOut: 20000,
      fragLoadingMaxRetry: 6,
      fragLoadingRetryDelay: 1000,
    });

    hlsRef.current = hls;

    console.log('[StreamPlayer] Loading HLS source...');
    hls.loadSource(url);
    console.log('[StreamPlayer] Attaching media to video element...');
    hls.attachMedia(video);

    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      if (!mountedRef.current) return;
      console.log('[StreamPlayer] Manifest parsed successfully');
      setLoading(false);
      setRetryMessage(null);
      retryCountRef.current = 0;

      if (hls.levels && hls.levels.length > 0) {
        console.log('[StreamPlayer] Available levels:', hls.levels);
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
          if (!mountedRef.current) return;
          // Ignore AbortError from React Strict Mode double invocation
          if (err.name === 'AbortError') {
            console.log('[StreamPlayer] Play aborted (likely React Strict Mode)');
            return;
          }
          console.error('[StreamPlayer] Autoplay failed:', err);
          setError('Click to play');
        });
      }
    });

    hls.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
      if (!mountedRef.current) return;
      console.log('[StreamPlayer] Level switched to:', data.level);
      setCurrentQuality(data.level);
    });

    hls.on(Hls.Events.ERROR, (_, data) => {
      if (!mountedRef.current) return;
      console.error('[StreamPlayer] HLS.js error:', {
        type: data.type,
        details: data.details,
        fatal: data.fatal,
      });

      if (data.fatal) {
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            console.error('[StreamPlayer] Fatal network error:', data.details);

            // Retry on manifest/level load errors (stream might not be ready yet)
            if (
              data.details === 'manifestLoadError' ||
              data.details === 'levelLoadError' ||
              data.details === 'levelLoadTimeOut'
            ) {
              if (retryCountRef.current < MAX_RETRIES) {
                retryCountRef.current++;
                const message = `Waiting for stream... (${retryCountRef.current}/${MAX_RETRIES})`;
                console.log(`[StreamPlayer] ${message}`);
                setRetryMessage(message);

                retryTimeoutRef.current = setTimeout(() => {
                  if (mountedRef.current && videoRef.current) {
                    initHls(videoRef.current, url);
                  }
                }, RETRY_DELAY);
              } else {
                setError('Stream not available. Make sure OBS is streaming.');
                setLoading(false);
                setRetryMessage(null);
              }
            } else {
              console.log('[StreamPlayer] Attempting to recover from network error...');
              hls.startLoad();
            }
            break;

          case Hls.ErrorTypes.MEDIA_ERROR:
            console.error('[StreamPlayer] Fatal media error:', data.details);
            console.log('[StreamPlayer] Attempting to recover from media error...');
            hls.recoverMediaError();
            break;

          default:
            console.error('[StreamPlayer] Fatal error, cannot recover:', data.type, data.details);
            setError(`Playback error: ${data.details}`);
            setLoading(false);
            break;
        }
      }
    });

    hls.on(Hls.Events.MEDIA_ATTACHED, () => {
      console.log('[StreamPlayer] Media attached to video element');
    });

    hls.on(Hls.Events.LEVEL_LOADED, (_, data) => {
      console.log('[StreamPlayer] Level loaded:', data.level);
    });

    hls.on(Hls.Events.FRAG_LOADED, () => {
      console.log('[StreamPlayer] Fragment loaded successfully');
    });
  }, [autoplay]);

  useEffect(() => {
    mountedRef.current = true;
    retryCountRef.current = 0;

    if (!videoRef.current || !hlsUrl) return;

    if (!hlsUrl.startsWith('http')) {
      setError('Invalid stream URL. Expected HTTP/HTTPS URL.');
      setLoading(false);
      return;
    }

    const video = videoRef.current;

    // Delay 10 seconds before connecting to allow FFmpeg to create playlists
    const INITIAL_DELAY = 10000;
    console.log(`[StreamPlayer] Waiting ${INITIAL_DELAY / 1000}s for stream to be ready...`);
    setRetryMessage('Preparing stream... (waiting for transcoder)');

    const delayTimeout = setTimeout(() => {
      if (!mountedRef.current) {
        console.log('[StreamPlayer] Component unmounted, skipping init');
        return;
      }
      console.log('[StreamPlayer] Delay complete, initializing HLS...');
      setRetryMessage(null);

      // Check if HLS is natively supported (Safari only)
      // Use HLS.js for all other browsers as it's more reliable
      const canPlayNativeHls = video.canPlayType('application/vnd.apple.mpegurl');
      const isSafari = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

      if (canPlayNativeHls && isSafari && !Hls.isSupported()) {
        console.log('[StreamPlayer] Using native HLS support (Safari)');
        video.src = hlsUrl;
        video.addEventListener('loadedmetadata', () => {
          if (mountedRef.current) setLoading(false);
        });
        if (autoplay) {
          video.play().catch((err) => {
            if (!mountedRef.current) return;
            if (err.name === 'AbortError') return;
            console.error('Autoplay failed:', err);
            setError('Click to play');
          });
        }
      } else if (Hls.isSupported()) {
        console.log('[StreamPlayer] Using HLS.js, calling initHls with URL:', hlsUrl);
        initHls(video, hlsUrl);
      } else {
        setError('HLS is not supported in this browser');
        setLoading(false);
      }
    }, INITIAL_DELAY);

    return () => {
      mountedRef.current = false;
      clearTimeout(delayTimeout);
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
        retryTimeoutRef.current = null;
      }
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
    };
  }, [hlsUrl, autoplay, initHls]);

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
        muted={autoplay}
        crossOrigin="anonymous"
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
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-black bg-opacity-50">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-white"></div>
          {retryMessage && (
            <p className="mt-4 text-white text-sm">{retryMessage}</p>
          )}
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
