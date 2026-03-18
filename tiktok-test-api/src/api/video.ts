import axiosInstance from './axios';
import type { Video, VideoResponse, PresignedURLResponse } from '../types';

export const videoAPI = {
  // Upload video
  uploadVideo: async (
    file: File,
    title: string,
    description: string,
    hashtags: string[] = [],
    categories: string[] = []
  ): Promise<Video> => {
    const formData = new FormData();
    formData.append('video', file);
    formData.append('title', title);
    formData.append('description', description);

    hashtags.forEach(tag => formData.append('hashtags', tag));
    categories.forEach(cat => formData.append('categories', cat));

    const response = await axiosInstance.post<Video>('/api/videos/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // Get video by ID (includes presigned URL)
  getVideo: async (videoId: string): Promise<VideoResponse> => {
    const response = await axiosInstance.get<VideoResponse>(`/api/videos/${videoId}`);
    return response.data;
  },

  // Get presigned URL for video streaming
  getPresignedURL: async (videoId: string): Promise<PresignedURLResponse> => {
    const response = await axiosInstance.get<PresignedURLResponse>(`/api/videos/${videoId}/presigned-url`);
    return response.data;
  },

  // Get videos by user
  getUserVideos: async (userId: string): Promise<Video[]> => {
    const response = await axiosInstance.get<Video[]>(`/api/videos/user/${userId}`);
    return response.data;
  },

  // Update video
  updateVideo: async (
    videoId: string,
    title?: string,
    hashtags?: string[]
  ): Promise<Video> => {
    const payload: { title?: string; hashtags?: string[] } = {};
    if (title) payload.title = title;
    if (hashtags) payload.hashtags = hashtags;

    const response = await axiosInstance.patch<Video>(`/api/videos/${videoId}`, payload);
    return response.data;
  },

  // Delete video
  deleteVideo: async (videoId: string): Promise<void> => {
    await axiosInstance.delete(`/api/videos/${videoId}`);
  },
};
