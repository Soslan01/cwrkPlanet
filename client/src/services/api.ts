import axios, { AxiosInstance, AxiosError } from 'axios';

// Use /api for same-origin requests (nginx/Vite proxy). Override with VITE_API_BASE_URL if needed.
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api';

export interface User {
  id: number;
  email: string;
  username: string;
  displayName: string;
  avatarUrl?: string | null;
  createdAt: number;
  updatedAt: number;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  displayName: string;
  avatarUrl?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface RefreshResponse {
  accessToken: string;
  refreshToken: string;
  expires_in: number;
}

export interface MeResponse {
  user: User;
}

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('accessToken');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Add response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as any;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            const refreshToken = localStorage.getItem('refreshToken');
            if (refreshToken) {
              const response = await this.refreshToken(refreshToken);
              localStorage.setItem('accessToken', response.accessToken);
              localStorage.setItem('refreshToken', response.refreshToken);
              originalRequest.headers.Authorization = `Bearer ${response.accessToken}`;
              return this.client(originalRequest);
            }
          } catch (refreshError) {
            // Refresh failed, clear tokens and redirect to login
            localStorage.removeItem('accessToken');
            localStorage.removeItem('refreshToken');
            window.location.href = '/';
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await this.client.post<RegisterResponse>('/v1/auth/register', data);
    return response.data;
  }

  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<LoginResponse>('/v1/auth/login', data);
    return response.data;
  }

  async refreshToken(refreshToken: string): Promise<RefreshResponse> {
    const response = await this.client.post<RefreshResponse>('/v1/auth/refresh', {
      refreshToken,
    });
    return response.data;
  }

  async logout(refreshToken: string): Promise<void> {
    await this.client.post('/v1/auth/logout', { refreshToken });
  }

  async getMe(): Promise<MeResponse> {
    const response = await this.client.get<MeResponse>('/v1/auth/me');
    return response.data;
  }
}

export const apiService = new ApiService();
