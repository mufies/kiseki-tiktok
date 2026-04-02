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
  // Camel case variants from API
  likeCount?: number;
  commentCount?: number;
  bookmarkCount?: number;
  viewCount?: number;
  isLiked?: boolean;
  isBookmarked?: boolean;
}

export interface VideoOwner {
  user_id?: string;
  username: string;
  display_name?: string;
  profile_image_url?: string | null;
  followers_count?: number;
  following_count?: number;
  is_verified?: boolean;
  is_followed?: boolean;
  // Camel case variants from API
  userId?: string;
  displayName?: string;
  profileImageUrl?: string | null;
  followersCount?: number;
  followingCount?: number;
  isVerified?: boolean;
  isFollowed?: boolean;
}

export interface Video {
  id?: string;
  video_id?: string;
  ownerId?: string;
  title: string;
  description?: string;
  mimeType?: string;
  size?: number;
  videoThumbnail?: string;
  hashtags?: string[];
  createdAt?: string;
  updatedAt?: string;
  score?: number; // From feed service
  interactions?: VideoInteractions; // From feed service
  owner?: VideoOwner; // From feed service
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

export interface Stream {
  id: string;
  user_id: string;
  stream_key: string;
  title: string;
  description: string;
  thumbnail_url: string;
  status: 'offline' | 'live' | 'ending';
  viewer_count: number;
  started_at?: string;
  ended_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ChatMessage {
  id: string;
  streamId: string;
  userId: string;
  username: string;
  content: string;
  timestamp: string;
  isStreamer?: boolean;
}
