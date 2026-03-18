import axiosInstance from './axios';
import type { WatchEvent } from '../types';

export const eventAPI = {
  // Send watch event
  sendWatchEvent: async (
    userId: string,
    videoId: string,
    watchPct: number,
    liked: boolean
  ): Promise<WatchEvent> => {
    const response = await axiosInstance.post<WatchEvent>('/events/watch', {
      userId,
      videoId,
      watchPct,
      liked,
      timestamp: new Date().toISOString(),
    });
    return response.data;
  },
};
