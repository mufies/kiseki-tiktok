export interface User {
  id: string;
  user_id?: string;
  username: string;
  email: string;
}

export interface VideoInteractions {
  like_count?: number;
  comment_count?: number;
  bookmark_count?: number;
  view_count?: number;
  is_liked?: boolean;
  is_bookmarked?: boolean;
}

export interface Video {
  id?: string;
  video_id?: string;
  ownerId?: string;
  title: string;
  description?: string;
  mimeType?: string;
  size?: number;
  hashtags?: string[];
  createdAt?: string;
  updatedAt?: string;
  score?: number; // From feed service
  interactions?: VideoInteractions; // From feed service
}

export interface VideoResponse {
  video: Video;
  streamUrl: string;
  expiresAt: string;
}

export interface PresignedURLResponse {
  streamUrl: string;
  expiresAt: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken?: string;
  user?: User;
}

export interface FeedResponse {
  videos: Video[];
}

export interface WatchEvent {
  userId: string;
  videoId: string;
  watchPct: number;
  liked: boolean;
  timestamp: string;
}
