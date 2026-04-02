import { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';

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

  useEffect(() => {
    if (!videoRef.current || !hlsUrl) return;

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
        if (autoplay) {
          video.play().catch((err) => {
            console.error('Autoplay failed:', err);
            setError('Click to play');
          });
        }
      });

      hls.on(Hls.Events.ERROR, (_, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              setError('Network error - attempting to recover');
              hls.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              setError('Media error - attempting to recover');
              hls.recoverMediaError();
              break;
            default:
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

  return (
    <div className="relative w-full aspect-video bg-black rounded-lg overflow-hidden">
      <video
        ref={videoRef}
        className="w-full h-full"
        controls
        poster={poster}
        onClick={handleClick}
        playsInline
      />

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
