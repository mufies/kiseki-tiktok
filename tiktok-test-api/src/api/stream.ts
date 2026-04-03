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

  // Get playback URL - Returns HLS URL for browser playback
  // API Endpoint: GET /streams/{stream_id}/playback
  // Returns: { playback_url: "http://localhost:8083/hls/{stream_id}/master.m3u8", protocol: "HLS" }
  getPlaybackUrl: async (streamId: string): Promise<StreamPlayback> => {
    console.log(`[Stream API] Requesting playback URL for stream ID: ${streamId}`);

    try {
      const response = await axiosInstance.get<{playback_url: string; protocol: string; note: string}>(`/streams/${streamId}/playback`);
      console.log('[Stream API] Backend playback response:', {
        playback_url: response.data.playback_url,
        protocol: response.data.protocol,
        note: response.data.note
      });

      const playbackUrl = response.data.playback_url;

      // Validate that we received an HLS URL (HTTP/HTTPS with .m3u8)
      // IMPORTANT: This should be HLS URL, NOT RTMP URL
      // ✓ Correct: http://localhost:8083/hls/{stream_id}/master.m3u8
      // ✗ Wrong: rtmp://localhost:1935/live/{stream_key}
      if (!playbackUrl) {
        throw new Error('Backend returned empty playback URL');
      }

      if (!playbackUrl.startsWith('http://') && !playbackUrl.startsWith('https://')) {
        console.error('[Stream API] ERROR: Backend returned non-HTTP URL:', playbackUrl);
        console.error('[Stream API] Expected HLS URL format: http://localhost:8083/hls/{stream_id}/master.m3u8');
        console.error('[Stream API] Got protocol:', playbackUrl.split(':')[0]);
        throw new Error(
          `Invalid playback URL: Expected HTTP/HTTPS HLS URL but got "${playbackUrl}". ` +
          `Make sure the backend /streams/${streamId}/playback endpoint returns the HLS URL, not RTMP URL.`
        );
      }

      if (!playbackUrl.includes('.m3u8')) {
        console.warn('[Stream API] Warning: Playback URL does not contain .m3u8:', playbackUrl);
      }

      console.log('[Stream API] ✓ Valid HLS URL received');

      // Map backend response to frontend interface
      return {
        hls_url: playbackUrl,
        rtmp_url: undefined,
      };
    } catch (error: any) {
      console.error('[Stream API] Failed to get playback URL:', error);

      // Provide helpful error messages
      if (error.response?.status === 404) {
        throw new Error(`Stream ${streamId} not found`);
      } else if (error.response?.status === 400) {
        throw new Error(error.response.data?.error || 'Stream is not live or not ready for playback');
      } else if (error.message) {
        throw error; // Re-throw our custom errors
      } else {
        throw new Error(`Failed to get playback URL: ${error.response?.data?.error || 'Unknown error'}`);
      }
    }
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
