import axiosInstance from './axios';

export type NotificationType = 0 | 1 | 2 | 3; // 0: Like, 1: Comment, 2: Follow, 3: Bookmark

export const NotificationTypes = {
  Like: 0 as NotificationType,
  Comment: 1 as NotificationType,
  Follow: 2 as NotificationType,
  Bookmark: 3 as NotificationType,
};

export interface NotificationDetail {
  id: string;
  userId: string;
  type: NotificationType;
  isRead: boolean;
  createdAt: string;
  fromUserId: string;
  fromUsername: string;
  fromAvatarUrl?: string;
  videoId?: string;
  videoTitle?: string;
  videoThumbnail?: string;
  commentId?: string;
  commentContent?: string;
  message: string;
}

export interface NotificationPagedResult {
  items: NotificationDetail[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface UnreadCountResponse {
  unreadCount: number;
}

export const notificationAPI = {
  // Get detailed notifications
  getNotifications: async (
    userId: string,
    page: number = 1,
    pageSize: number = 20
  ): Promise<NotificationPagedResult> => {
    const response = await axiosInstance.get<NotificationPagedResult>(
      `/notifications/${userId}`,
      {
        params: { page, pageSize, detailed: true },
      }
    );
    return response.data;
  },

  // Get unread count
  getUnreadCount: async (userId: string): Promise<number> => {
    const response = await axiosInstance.get<UnreadCountResponse>(
      `/notifications/${userId}/unread-count`
    );
    return response.data.unreadCount;
  },

  // Mark notifications as read
  markAsRead: async (userId: string, notificationIds: string[]): Promise<void> => {
    await axiosInstance.post(`/notifications/${userId}/mark-read`, {
      notificationIds,
    });
  },

  // Mark all as read
  markAllAsRead: async (userId: string): Promise<void> => {
    await axiosInstance.post(`/notifications/${userId}/mark-all-read`);
  },
};
