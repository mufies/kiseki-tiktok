import axiosInstance from './axios';
import type { User } from '../types';

export interface UserProfile extends User {
  followerCount?: number;
  followingCount?: number;
  createdAt?: string;
  avatarUrl?: string;
}

export interface UpdateProfileRequest {
  username?: string;
  avatarUrl?: string;
}

export const userAPI = {
  // Get current user profile
  getCurrentProfile: async (): Promise<UserProfile> => {
    const response = await axiosInstance.get<UserProfile>('api/users/me');
    return response.data;
  },

  // Get user profile by ID
  getUserProfile: async (userId: string): Promise<UserProfile> => {
    const response = await axiosInstance.get<UserProfile>(`api/users/${userId}`);
    return response.data;
  },

  // Get user profile by username
  getUserByUsername: async (username: string): Promise<UserProfile> => {
    const response = await axiosInstance.get<UserProfile>(`api/users/username/${username}`);
    return response.data;
  },

  // Update user profile
  updateProfile: async (data: UpdateProfileRequest): Promise<UserProfile> => {
    const response = await axiosInstance.put<UserProfile>('api/users/me', data);
    return response.data;
  },

  // Change email
  changeEmail: async (newEmail: string, currentPassword: string): Promise<UserProfile> => {
    const response = await axiosInstance.put<UserProfile>('api/users/me/email', {
      newEmail,
      currentPassword,
    });
    return response.data;
  },

  // Change password
  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    await axiosInstance.put('api/users/me/password', {
      currentPassword,
      newPassword,
    });
  },

  // Get followers
  getFollowers: async (userId: string): Promise<UserProfile[]> => {
    const response = await axiosInstance.get<UserProfile[]>(`api/users/${userId}/followers`);
    return response.data;
  },

  // Get following
  getFollowing: async (userId: string): Promise<UserProfile[]> => {
    const response = await axiosInstance.get<UserProfile[]>(`api/users/${userId}/following`);
    return response.data;
  },

  // Follow user
  followUser: async (userId: string): Promise<void> => {
    await axiosInstance.post(`api/users/${userId}/follow`);
  },

  // Unfollow user
  unfollowUser: async (userId: string): Promise<void> => {
    await axiosInstance.delete(`api/users/${userId}/follow`);
  },
};
