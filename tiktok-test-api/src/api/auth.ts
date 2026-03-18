import axios from 'axios';
import axiosInstance, { setAccessToken, getAccessToken } from './axios';
import type { AuthResponse, User } from '../types';
import { jwtDecode } from 'jwt-decode';

const API_BASE_URL = 'http://localhost:8080';

export const authAPI = {
  // Register new user
  register: async (username: string, email: string, password: string): Promise<AuthResponse> => {
    const response = await axios.post<AuthResponse>(`${API_BASE_URL}/auth/register`, {
      username,
      email,
      password,
    });
    return response.data;
  },

  // Login
  login: async (emailOrUsername: string, password: string): Promise<AuthResponse> => {
    const response = await axios.post<AuthResponse>(
      `${API_BASE_URL}/auth/login`,
      {
        email: emailOrUsername,
        username: emailOrUsername,
        password,
      },
      { withCredentials: true }
    );

    // Store access token
    setAccessToken(response.data.accessToken);
    return response.data;
  },

  // Logout
  logout: async (): Promise<void> => {
    await axiosInstance.post('/auth/logout');
    setAccessToken(null);
  },

  // Get current user profile
  getCurrentUser: async (): Promise<User> => {
    const token = getAccessToken();
    if (!token) {
      throw new Error('No access token available');
    }

    // Decode JWT token to get userId from sub claim
    const decoded = jwtDecode<{ sub: string }>(token);
    const userId = decoded.sub;

    // Add userId to request as query parameter
    const response = await axiosInstance.get<User>('/api/users/me', {
      params: { userId }
    });
    return response.data;
  },
};
