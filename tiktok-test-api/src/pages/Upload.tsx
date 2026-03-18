import { useRef, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { videoAPI } from '../api/video';
import { useAuth } from '../context/AuthContext';
import { UploadCloud, Film, X, ChevronLeft } from 'lucide-react';

export default function Upload() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [hashtags, setHashtags] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const MAX_FILE_SIZE = 100 * 1024 * 1024; // 100MB
  const MAX_TITLE_LENGTH = 100;
  const MAX_DESCRIPTION_LENGTH = 500;

  const processFile = (selectedFile: File) => {
    if (!selectedFile.type.startsWith('video/')) {
      setError('Please select a valid video file');
      return;
    }

    if (selectedFile.size > MAX_FILE_SIZE) {
      setError('Video file must be less than 100MB');
      return;
    }

    setFile(selectedFile);
    setError(null);

    const reader = new FileReader();
    reader.onload = (e) => {
      setPreview(e.target?.result as string);
    };
    reader.readAsDataURL(selectedFile);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) processFile(selectedFile);
  };

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const onDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const selectedFile = e.dataTransfer.files?.[0];
    if (selectedFile) processFile(selectedFile);
  }, []);

  const clearFile = () => {
    setFile(null);
    setPreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user) {
      setError('Please login to upload videos');
      return;
    }

    if (!file) {
      setError('Please select a video file');
      return;
    }

    if (!title.trim()) {
      setError('Title is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const hashtagList = hashtags
        .split(' ')
        .map(tag => tag.startsWith('#') ? tag.slice(1) : tag)
        .filter(tag => tag.length > 0);

      await videoAPI.uploadVideo(file, title, description, hashtagList);
      navigate('/profile');
    } catch (err) {
      console.error('Upload failed:', err);
      setError('Failed to upload video. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-black text-white selection:bg-purple-900 pb-20">
      <header className="sticky top-0 bg-black/80 backdrop-blur-md z-40 border-b border-zinc-800 flex items-center p-4 mb-8">
        <button onClick={() => navigate(-1)} className="text-zinc-400 hover:text-white transition mr-4">
          <ChevronLeft size={28} />
        </button>
        <h1 className="font-bold text-xl">Upload video</h1>
      </header>

      <div className="max-w-4xl mx-auto px-4">
        <div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-6 md:p-8 shadow-xl flex flex-col lg:flex-row gap-8">

          {/* Left: Upload Area */}
          <div className="w-full lg:w-1/2 flex flex-col gap-4">
            <h2 className="text-lg font-bold mb-2">Video file</h2>
            {!file ? (
              <div
                className={`flex-1 border-2 border-dashed rounded-2xl flex flex-col items-center justify-center p-8 text-center transition-all bg-black cursor-pointer group hover:bg-zinc-950 ${isDragging ? 'border-purple-500 bg-purple-900/10' : 'border-zinc-700 hover:border-zinc-500'
                  }`}
                onClick={() => fileInputRef.current?.click()}
                onDragOver={onDragOver}
                onDragLeave={onDragLeave}
                onDrop={onDrop}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="video/*"
                  onChange={handleFileChange}
                  className="hidden"
                />
                <UploadCloud size={48} className={`mb-4 transition-colors ${isDragging ? 'text-purple-400' : 'text-zinc-500 group-hover:text-purple-400'}`} />
                <p className="text-base font-semibold mb-2">Select video to upload</p>
                <p className="text-sm text-zinc-500 mb-6">Or drag and drop a file</p>

                <div className="text-xs text-zinc-600 space-y-1">
                  <p>MP4 or WebM</p>
                  <p>720x1280 resolution or higher</p>
                  <p>Up to 10 minutes</p>
                  <p>Less than 100MB</p>
                </div>

                <button className="mt-8 bg-purple-600 hover:bg-purple-700 text-white font-semibold py-2 px-8 rounded-lg shadow-md transition">
                  Select file
                </button>
              </div>
            ) : (
              <div className="flex-1 flex flex-col bg-black border border-zinc-800 rounded-2xl overflow-hidden relative">
                {preview ? (
                  <video
                    src={preview}
                    controls
                    className="w-full h-full object-contain max-h-[500px]"
                  />
                ) : (
                  <div className="flex-1 flex items-center justify-center p-8">
                    <Film size={48} className="text-zinc-600 mb-4" />
                  </div>
                )}

                {/* Overlay details */}
                <div className="absolute top-0 left-0 right-0 bg-gradient-to-b from-black/80 to-transparent p-4 flex justify-between items-start">
                  <div className="text-sm truncate pr-4 drop-shadow-md">
                    <p className="font-semibold truncate">{file.name}</p>
                    <p className="text-zinc-300">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); clearFile(); }}
                    className="bg-black/50 hover:bg-red-500/80 text-white p-1.5 rounded-full backdrop-blur-sm transition"
                    title="Remove file"
                  >
                    <X size={18} />
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Right: Form */}
          <form onSubmit={handleSubmit} className="w-full lg:w-1/2 flex flex-col gap-5">
            <div>
              <label className="block text-sm font-semibold mb-2">Caption <span className="text-red-500">*</span></label>
              <div className="relative">
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value.slice(0, MAX_TITLE_LENGTH))}
                  placeholder="What's your video about?"
                  className="w-full px-4 py-3 bg-black border border-zinc-700 rounded-xl text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 focus:ring-1 focus:ring-purple-500 transition text-sm"
                  maxLength={MAX_TITLE_LENGTH}
                />
                <span className="absolute right-3 top-3 text-zinc-500 text-xs font-mono">
                  {title.length}/{MAX_TITLE_LENGTH}
                </span>
              </div>
            </div>

            <div>
              <label className="block text-sm font-semibold mb-2">Description</label>
              <div className="relative">
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value.slice(0, MAX_DESCRIPTION_LENGTH))}
                  placeholder="Add more details..."
                  rows={4}
                  className="w-full px-4 py-3 bg-black border border-zinc-700 rounded-xl text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 focus:ring-1 focus:ring-purple-500 transition resize-none text-sm"
                  maxLength={MAX_DESCRIPTION_LENGTH}
                />
                <span className="absolute right-3 bottom-3 text-zinc-500 text-xs font-mono">
                  {description.length}/{MAX_DESCRIPTION_LENGTH}
                </span>
              </div>
            </div>

            <div>
              <label className="block text-sm font-semibold mb-2">Hashtags</label>
              <input
                type="text"
                value={hashtags}
                onChange={(e) => setHashtags(e.target.value)}
                placeholder="#dance #trending"
                className="w-full px-4 py-3 bg-black border border-zinc-700 rounded-xl text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 focus:ring-1 focus:ring-purple-500 transition text-sm"
              />
            </div>

            {error && (
              <div className="bg-red-500/10 border border-red-500/50 rounded-lg p-3 flex items-center gap-2 text-sm text-red-200">
                <div className="w-1.5 h-1.5 rounded-full bg-red-500"></div>
                {error}
              </div>
            )}

            <div className="mt-auto pt-6 flex gap-3">
              <button
                type="button"
                onClick={() => navigate(-1)}
                className="flex-1 bg-zinc-800 hover:bg-zinc-700 text-white font-semibold py-3 rounded-lg transition"
              >
                Discard
              </button>
              <button
                type="submit"
                disabled={loading || !file || !title.trim()}
                className={`flex-[2] font-semibold py-3 rounded-lg transition flex items-center justify-center gap-2 ${loading || !file || !title.trim()
                    ? 'bg-zinc-800 text-zinc-500 cursor-not-allowed'
                    : 'bg-purple-600 hover:bg-purple-500 text-white shadow-lg shadow-purple-900/20'
                  }`}
              >
                {loading && <div className="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" />}
                {loading ? 'Posting...' : 'Post'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
