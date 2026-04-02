import axiosInstance from './axios';

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

export interface CreateStreamRequest {
  user_id: string;
  title: string;
  description?: string;
  save_vod?: boolean;
}

export interface StreamPlayback {
  hls_url: string;
  rtmp_url?: string;
}

export interface LiveStreamsResponse {
  streams: Stream[];
  total: number;
  limit: number;
  offset: number;
}

export const streamAPI = {
  // Create new stream
  createStream: async (data: CreateStreamRequest): Promise<Stream> => {
    const response = await axiosInstance.post<{message: string; stream: Stream}>('/streams', data);
    return response.data.stream; // Extract stream from nested response
  },

  // Get stream by ID
  getStream: async (streamId: string): Promise<Stream> => {
    const response = await axiosInstance.get<{stream: Stream}>(`/streams/${streamId}`);
    return response.data.stream;
  },

  // Update stream
  updateStream: async (streamId: string, data: Partial<CreateStreamRequest>): Promise<Stream> => {
    const response = await axiosInstance.patch<{stream: Stream; message: string}>(`/streams/${streamId}`, data);
    return response.data.stream;
  },

  // Delete stream
  deleteStream: async (streamId: string): Promise<void> => {
    await axiosInstance.delete(`/streams/${streamId}`);
  },

  // Start stream
  startStream: async (streamId: string): Promise<Stream> => {
    const response = await axiosInstance.post<Stream>(`/streams/${streamId}/start`);
    return response.data;
  },

  // End stream
  endStream: async (streamId: string): Promise<Stream> => {
    const response = await axiosInstance.post<Stream>(`/streams/${streamId}/end`);
    return response.data;
  },

  // Get live streams
  getLiveStreams: async (limit = 20, offset = 0): Promise<LiveStreamsResponse> => {
    const response = await axiosInstance.get<LiveStreamsResponse>('/streams/live', {
      params: { limit, offset },
    });
    return response.data;
  },

  // Get user streams
  getUserStreams: async (userId: string): Promise<Stream[]> => {
    const response = await axiosInstance.get<Stream[]>(`/streams/user/${userId}`);
    return response.data;
  },

  // Get playback URL
  getPlaybackUrl: async (streamId: string): Promise<StreamPlayback> => {
    const response = await axiosInstance.get<{playback_url: string; protocol: string; note: string}>(`/streams/${streamId}/playback`);
    // Map backend response to frontend interface
    return {
      hls_url: response.data.playback_url,
      rtmp_url: undefined,
    };
  },

  // Join stream (increment viewer count)
  joinStream: async (streamId: string): Promise<void> => {
    await axiosInstance.post(`/streams/${streamId}/viewers/join`);
  },

  // Leave stream (decrement viewer count)
  leaveStream: async (streamId: string): Promise<void> => {
    await axiosInstance.post(`/streams/${streamId}/viewers/leave`);
  },
};
