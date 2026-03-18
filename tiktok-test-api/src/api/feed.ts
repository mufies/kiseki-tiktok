import axiosInstance from './axios';
import type { FeedResponse, Video } from '../types';

export const feedAPI = {
  // Get personalized feed for user
  getFeed: async (userId: string, limit: number = 20): Promise<FeedResponse> => {
    const response = await axiosInstance.get<FeedResponse>(`/feed/${userId}`, {
      params: { limit },
    });
    return response.data;
  },

  // Get trending videos
  getTrending: async (limit: number = 20): Promise<Video[]> => {
    const response = await axiosInstance.get<Video[]>('/trending', {
      params: { limit },
    });
    return response.data;
  },
};
