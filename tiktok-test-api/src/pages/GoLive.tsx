import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Copy, Radio, Users, CheckCircle } from 'lucide-react';
import { streamAPI } from '../api/stream';
import type { Stream } from '../api/stream';
import { useAuth } from '../context/AuthContext';
import StreamPlayer from '../components/StreamPlayer';
import StreamChat from '../components/StreamChat';

const RTMP_SERVER = import.meta.env.VITE_RTMP_URL || 'rtmp://localhost:1935/live';

export default function GoLive() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [step, setStep] = useState<'setup' | 'live'>('setup');

  // Setup state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [stream, setStream] = useState<Stream | null>(null);
  const [creating, setCreating] = useState(false);
  const [copiedKey, setCopiedKey] = useState(false);
  const [copiedRtmp, setCopiedRtmp] = useState(false);
  const [instructionTab, setInstructionTab] = useState<'obs' | 'ffmpeg'>('obs');

  // Live state
  const [hlsUrl, setHlsUrl] = useState('');
  const [viewerCount, setViewerCount] = useState(0);

  // Poll for stream status when waiting for OBS connection
  useEffect(() => {
    if (!stream?.id || stream.status === 'live') return;

    const checkStreamStatus = async () => {
      try {
        const updated = await streamAPI.getStream(stream.id);
        if (updated.status === 'live' && stream.status !== 'live') {
          setStream(updated);
        }
      } catch (error) {
        console.error('Failed to check stream status:', error);
      }
    };

    const interval = setInterval(checkStreamStatus, 3000);
    return () => clearInterval(interval);
  }, [stream]);

  // Poll viewer count when live
  useEffect(() => {
    if (step !== 'live' || !stream) return;

    const updateViewerCount = async () => {
      try {
        const updated = await streamAPI.getStream(stream.id);
        setViewerCount(updated.viewer_count);
      } catch (error) {
        console.error('Failed to update viewer count:', error);
      }
    };

    const interval = setInterval(updateViewerCount, 5000);
    return () => clearInterval(interval);
  }, [step, stream]);

  const handleCreateStream = async () => {
    if (!title.trim()) {
      alert('Please enter a stream title');
      return;
    }

    if (!user) {
      alert('User not authenticated');
      return;
    }

    setCreating(true);
    try {
      const userId = user.user_id || user.id;
      const payload = {
        user_id: userId,
        title: title.trim(),
        description: description.trim() || undefined,
        save_vod: true,
      };

      console.log('Creating stream with payload:', payload);
      const newStream = await streamAPI.createStream(payload);
      console.log('Stream created:', newStream);
      setStream(newStream);
    } catch (error: any) {
      console.error('Failed to create stream:', error);
      console.error('Error response:', error?.response?.data);
      alert(`Failed to create stream: ${error?.response?.data?.error || error?.message || 'Unknown error'}`);
    } finally {
      setCreating(false);
    }
  };

  const handleStartLive = async () => {
    if (!stream) return;

    if (stream.status !== 'live') {
      alert('Please connect OBS and start streaming first');
      return;
    }

    try {
      // Get playback URL
      const playback = await streamAPI.getPlaybackUrl(stream.id);
      setHlsUrl(playback.hls_url);
      setStep('live');
    } catch (error) {
      console.error('Failed to start live:', error);
      alert('Failed to start live stream. Please try again.');
    }
  };

  const handleEndLive = async () => {
    if (!stream) return;

    const confirmed = confirm('Are you sure you want to end this stream?');
    if (!confirmed) return;

    try {
      await streamAPI.endStream(stream.id);
      navigate('/');
    } catch (error) {
      console.error('Failed to end stream:', error);
      alert('Failed to end stream. Please try again.');
    }
  };

  const copyToClipboard = async (text: string, type: 'key' | 'rtmp') => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'key') {
        setCopiedKey(true);
        setTimeout(() => setCopiedKey(false), 2000);
      } else {
        setCopiedRtmp(true);
        setTimeout(() => setCopiedRtmp(false), 2000);
      }
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  if (!user) return null;

  // Live view
  if (step === 'live' && stream) {
    return (
      <div className="min-h-screen bg-black">
        {/* Header */}
        <div className="bg-gray-900 border-b border-gray-800 px-6 py-4">
          <div className="max-w-7xl mx-auto flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Radio className="w-5 h-5 text-red-500 animate-pulse" />
                <span className="text-red-500 font-bold">LIVE</span>
              </div>
              <h1 className="text-white text-xl font-semibold">{stream.title}</h1>
            </div>
            <button
              onClick={handleEndLive}
              className="px-4 py-2 bg-red-600 text-white rounded-lg font-semibold hover:bg-red-700 transition-colors"
            >
              End Live
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="max-w-7xl mx-auto p-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Stream preview */}
            <div className="lg:col-span-2 space-y-4">
              <StreamPlayer hlsUrl={hlsUrl} poster={stream.thumbnail_url} />

              <div className="bg-gray-900 rounded-lg p-4 flex items-center gap-4 text-white">
                <Users className="w-5 h-5 text-purple-500" />
                <div>
                  <p className="text-sm text-gray-400">Current Viewers</p>
                  <p className="text-2xl font-bold">{viewerCount.toLocaleString()}</p>
                </div>
              </div>
            </div>

            {/* Chat */}
            <div className="lg:col-span-1 h-[600px]">
              <StreamChat
                streamId={stream.id}
                currentUserId={user.user_id || user.id}
                currentUsername={user.username}
                streamOwnerId={stream.user_id}
              />
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Setup view
  return (
    <div className="min-h-screen bg-black">
      {/* Header */}
      <div className="bg-gray-900 border-b border-gray-800 px-6 py-4">
        <div className="max-w-3xl mx-auto flex items-center gap-4">
          <button
            onClick={() => navigate('/')}
            className="text-white hover:text-gray-300 transition-colors"
          >
            <ArrowLeft className="w-6 h-6" />
          </button>
          <h1 className="text-white text-2xl font-bold">Go Live</h1>
        </div>
      </div>

      {/* Setup form */}
      <div className="max-w-3xl mx-auto p-6">
        {!stream ? (
          // Step 1: Create Stream Form
          <div className="bg-gray-900 rounded-lg p-6 space-y-6">
            <div className="text-center mb-6">
              <h2 className="text-white text-2xl font-bold mb-2">Create Your Stream</h2>
              <p className="text-gray-400">Fill in your stream details to generate a stream key</p>
            </div>

            {/* Title */}
            <div>
              <label className="block text-white font-semibold mb-2">
                Stream Title *
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Enter your stream title"
                className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                maxLength={100}
              />
            </div>

            {/* Description */}
            <div>
              <label className="block text-white font-semibold mb-2">
                Description (Optional)
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Tell viewers what your stream is about"
                className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500 resize-none"
                rows={3}
                maxLength={500}
              />
            </div>

            {/* Generate Stream Key button */}
            <button
              onClick={handleCreateStream}
              disabled={creating || !title.trim()}
              className="w-full px-6 py-4 bg-purple-600 text-white rounded-lg font-bold text-lg hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
            >
              {creating ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-t-2 border-b-2 border-white"></div>
                  Generating Stream Key...
                </>
              ) : (
                <>
                  <Radio className="w-5 h-5" />
                  Generate Stream Key
                </>
              )}
            </button>

            <div className="bg-blue-900/30 border border-blue-700/50 rounded-lg p-4 text-sm text-blue-300">
              <p><strong>Note:</strong> After generating your stream key, you'll receive instructions to connect your streaming software (OBS, FFmpeg, etc.)</p>
            </div>
          </div>
        ) : (
          // Step 2: Show Instructions
          <div className="space-y-6">
            <div className="bg-green-900/30 border border-green-700/50 rounded-lg p-4 text-center">
              <h3 className="text-green-400 font-bold text-lg mb-1">✓ Stream Key Generated!</h3>
              <p className="text-gray-300 text-sm">Follow the instructions below to connect your streaming software</p>
            </div>

            <div className="bg-gray-800 rounded-lg overflow-hidden">
              {/* Tabs */}
              <div className="flex border-b border-gray-700">
                <button
                  onClick={() => setInstructionTab('obs')}
                  className={`flex-1 px-6 py-3 font-semibold transition-colors ${
                    instructionTab === 'obs'
                      ? 'bg-gray-700 text-white'
                      : 'bg-gray-800 text-gray-400 hover:text-white'
                  }`}
                >
                  OBS Studio
                </button>
                <button
                  onClick={() => setInstructionTab('ffmpeg')}
                  className={`flex-1 px-6 py-3 font-semibold transition-colors ${
                    instructionTab === 'ffmpeg'
                      ? 'bg-gray-700 text-white'
                      : 'bg-gray-800 text-gray-400 hover:text-white'
                  }`}
                >
                  FFmpeg
                </button>
              </div>

              <div className="p-6 space-y-6">
                {instructionTab === 'obs' ? (
                  // OBS Instructions
                  <>
                    <div>
                      <h3 className="text-white text-lg font-semibold mb-4">Setup OBS Studio</h3>

                      {/* Step 1 */}
                      <div className="mb-6">
                        <div className="flex items-start gap-3 mb-3">
                          <div className="w-6 h-6 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold flex-shrink-0">
                            1
                          </div>
                          <div className="flex-1">
                            <p className="text-white font-semibold mb-1">Open OBS Studio Settings</p>
                            <p className="text-gray-400 text-sm">Go to Settings → Stream</p>
                          </div>
                        </div>
                      </div>

                      {/* Step 2 */}
                      <div className="mb-6">
                        <div className="flex items-start gap-3 mb-3">
                          <div className="w-6 h-6 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold flex-shrink-0">
                            2
                          </div>
                          <div className="flex-1">
                            <p className="text-white font-semibold mb-2">Configure Service</p>
                            <div className="bg-gray-900 rounded-lg p-3 space-y-2 text-sm">
                              <div className="flex items-center gap-2">
                                <CheckCircle className="w-4 h-4 text-green-500" />
                                <span className="text-gray-300">Service: <span className="text-white font-medium">Custom</span></span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Step 3 - Server */}
                      <div className="mb-6">
                        <div className="flex items-start gap-3 mb-3">
                          <div className="w-6 h-6 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold flex-shrink-0">
                            3
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center justify-between mb-2">
                              <p className="text-white font-semibold">Paste Server URL</p>
                              <button
                                onClick={() => copyToClipboard(RTMP_SERVER, 'rtmp')}
                                className="flex items-center gap-1 text-purple-500 hover:text-purple-400 text-sm"
                              >
                                <Copy className="w-4 h-4" />
                                {copiedRtmp ? 'Copied!' : 'Copy'}
                              </button>
                            </div>
                            <div className="px-4 py-3 bg-gray-900 rounded font-mono text-sm text-white break-all">
                              {RTMP_SERVER}
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Step 4 - Stream Key */}
                      <div className="mb-6">
                        <div className="flex items-start gap-3 mb-3">
                          <div className="w-6 h-6 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold flex-shrink-0">
                            4
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center justify-between mb-2">
                              <p className="text-white font-semibold">Paste Stream Key</p>
                              <button
                                onClick={() => copyToClipboard(stream.stream_key, 'key')}
                                className="flex items-center gap-1 text-purple-500 hover:text-purple-400 text-sm"
                              >
                                <Copy className="w-4 h-4" />
                                {copiedKey ? 'Copied!' : 'Copy'}
                              </button>
                            </div>
                            <div className="px-4 py-3 bg-gray-900 rounded font-mono text-sm text-white break-all">
                              {stream.stream_key}
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Step 5 */}
                      <div className="mb-6">
                        <div className="flex items-start gap-3">
                          <div className="w-6 h-6 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold flex-shrink-0">
                            5
                          </div>
                          <div className="flex-1">
                            <p className="text-white font-semibold mb-1">Start Streaming in OBS</p>
                            <p className="text-gray-400 text-sm">Click "Start Streaming" button in OBS</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </>
                ) : (
                  // FFmpeg Instructions
                  <>
                    <div>
                      <h3 className="text-white text-lg font-semibold mb-4">Stream with FFmpeg</h3>

                      <div className="mb-4">
                        <p className="text-gray-400 text-sm mb-4">
                          Use this command to stream a video file or camera input:
                        </p>

                        <div className="bg-gray-900 rounded-lg p-4">
                          <pre className="text-xs text-gray-300 overflow-x-auto">
{`ffmpeg -re -i video.mp4 \\
  -c:v libx264 -preset veryfast \\
  -maxrate 3000k -bufsize 6000k \\
  -pix_fmt yuv420p -g 50 \\
  -c:a aac -b:a 160k -ac 2 -ar 44100 \\
  -f flv ${RTMP_SERVER}/${stream.stream_key}`}
                          </pre>
                        </div>

                        <button
                          onClick={() => copyToClipboard(
                            `ffmpeg -re -i video.mp4 -c:v libx264 -preset veryfast -maxrate 3000k -bufsize 6000k -pix_fmt yuv420p -g 50 -c:a aac -b:a 160k -ac 2 -ar 44100 -f flv ${RTMP_SERVER}/${stream.stream_key}`,
                            'key'
                          )}
                          className="mt-3 flex items-center gap-2 text-purple-500 hover:text-purple-400 text-sm"
                        >
                          <Copy className="w-4 h-4" />
                          {copiedKey ? 'Copied!' : 'Copy Command'}
                        </button>
                      </div>

                      <div className="bg-blue-900/30 border border-blue-700/50 rounded-lg p-4">
                        <p className="text-blue-300 text-sm">
                          <strong>Note:</strong> Replace <code className="bg-blue-900/50 px-1 rounded">video.mp4</code> with your video file or camera input
                        </p>
                      </div>
                    </div>
                  </>
                )}

                {/* Connection Status */}
                <div className="pt-4 border-t border-gray-700">
                  <div className="flex items-center gap-3 mb-4">
                    <div className={`w-3 h-3 rounded-full ${stream.status === 'live' ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`}></div>
                    <div>
                      <p className="text-white font-semibold">
                        {stream.status === 'live' ? 'Connected! Ready to go live' : 'Waiting for connection...'}
                      </p>
                      <p className="text-gray-400 text-sm">
                        {stream.status === 'live'
                          ? 'Your stream is connected and ready to broadcast'
                          : 'Start streaming from OBS/FFmpeg to connect'}
                      </p>
                    </div>
                  </div>

                  {/* Connection Details */}
                  <div className="bg-gray-900 rounded-lg p-3 mb-4">
                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <p className="text-gray-500 text-xs mb-1">RTMP Port</p>
                        <p className="text-white font-mono">1935</p>
                      </div>
                      <div>
                        <p className="text-gray-500 text-xs mb-1">Status</p>
                        <p className={`font-semibold ${stream.status === 'live' ? 'text-green-500' : 'text-yellow-500'}`}>
                          {stream.status === 'live' ? 'Connected' : 'Waiting'}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                <button
                  onClick={handleStartLive}
                  disabled={stream.status !== 'live'}
                  className="w-full px-6 py-3 bg-red-600 text-white rounded-lg font-semibold hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                >
                  <Radio className="w-5 h-5" />
                  {stream.status === 'live' ? 'Start Broadcasting' : 'Waiting for Connection...'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
