import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Radio } from 'lucide-react';
import VideoFeed from '../components/VideoFeed';

export default function Home() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Failed to logout:', error);
    }
  };

  return (
    <div className="min-h-screen bg-black relative">
      <header className="absolute top-0 left-0 right-0 z-50 pointer-events-none p-4 pt-6">
        <div className="max-w-7xl mx-auto flex items-center justify-between pointer-events-auto">
          <h2 className="text-xl font-bold text-white drop-shadow-md">For You</h2>
          {user && (
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate('/go-live')}
                className="px-4 py-2 bg-red-600 text-white font-semibold text-sm rounded-lg flex items-center gap-2 drop-shadow-lg hover:bg-red-700 transition"
              >
                <Radio className="w-4 h-4" />
                Go Live
              </button>
              <button
                onClick={() => navigate('/upload')}
                className="text-white font-semibold text-sm flex items-center gap-1 drop-shadow-md hover:text-gray-300 transition"
              >
                + Upload
              </button>
              <button
                onClick={() => navigate('/profile')}
                className="text-white font-semibold text-sm drop-shadow-md hover:text-gray-300 transition"
              >
                Profile
              </button>
              <button
                onClick={handleLogout}
                className="text-white font-semibold text-sm drop-shadow-md hover:text-gray-300 transition"
              >
                Logout
              </button>
            </div>
          )}
        </div>
      </header>

      <main className="h-[100dvh]">
        <VideoFeed />
      </main>
    </div>
  );
}
