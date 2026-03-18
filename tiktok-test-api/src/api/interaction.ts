import axiosInstance from './axios';

interface InteractionResponse {
  success: boolean;
  message?: string;
}

interface Comment {
  id: string;
  text: string;
  userId: string;
  createdAt: string;
}

interface LikesCount {
  count: number;
}

export interface LikedVideo {
  interactionId: number;
  likedAt: string;
  videoId: string;
  title: string;
  hashtags: string[];
  categories: string[];
  available: boolean;
}

export const interactionAPI = {
  // Like/unlike video
  toggleLike: async (videoId: string): Promise<InteractionResponse> => {
    const response = await axiosInstance.post<InteractionResponse>(`/interactions/videos/${videoId}/like`);
    return response.data;
  },

  // Bookmark/unbookmark video
  toggleBookmark: async (videoId: string): Promise<InteractionResponse> => {
    const response = await axiosInstance.post<InteractionResponse>(`/interactions/videos/${videoId}/bookmarked`);
    return response.data;
  },

  // Record view
  recordView: async (videoId: string): Promise<InteractionResponse> => {
    const response = await axiosInstance.post<InteractionResponse>(`/interactions/videos/${videoId}/view`);
    return response.data;
  },

  // Add comment
  addComment: async (videoId: string, content: string): Promise<Comment> => {
    const response = await axiosInstance.post<Comment>(`/interactions/videos/${videoId}/comment`, {
      content,
    });
    return response.data;
  },

  // Get likes count
  getLikesCount: async (videoId: string): Promise<LikesCount> => {
    const response = await axiosInstance.get<LikesCount>(`/interactions/videos/${videoId}/likes`);
    return response.data;
  },

  // Get comments
  getComments: async (videoId: string): Promise<Comment[]> => {
    const response = await axiosInstance.get<Comment[]>(`/interactions/videos/${videoId}/comments`);
    return response.data;
  },

  getUserLikes: async (userId: string): Promise<LikedVideo[]> => {
    const response = await axiosInstance.get<LikedVideo[]>(`/interactions/videos/users/${userId}/liked-videos`);
    return response.data;
  },
};
