import axios, {
    type AxiosInstance,
    type InternalAxiosRequestConfig,
    type AxiosResponse,
    type AxiosError
} from 'axios';
import { getAccessToken, useAuthStore, type AuthResponse } from '../stores/auth';

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Custom config type with retry flag
interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
    _retry?: boolean;
}

// Create axios instance with base config
const apiClient: AxiosInstance = axios.create({
    baseURL: API_BASE_URL,
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request Interceptor: Inject in-memory JWT token
apiClient.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = getAccessToken();
        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Shared promise for deduplicating concurrent refresh requests
let refreshPromise: Promise<string> | null = null;

// Response Interceptor: Handle 401 errors, silent refresh, single retry, and concurrent deduplication
apiClient.interceptors.response.use(
    (response: AxiosResponse) => response,
    async (error: AxiosError) => {
        const originalRequest = error.config as CustomAxiosRequestConfig | undefined;

        if (!error.response || error.response.status !== 401 || !originalRequest) {
            return Promise.reject(error);
        }

        const isAuthEndpoint =
            originalRequest.url?.includes('/authentication/refresh') ||
            originalRequest.url?.includes('/authentication/request-otp') ||
            originalRequest.url?.includes('/authentication/verify-otp') ||
            originalRequest.url?.includes('/authentication/logout');

        // If 401 happens on an auth endpoint or has already been retried, fail and redirect
        if (isAuthEndpoint || originalRequest._retry) {
            try {
                const authStore = useAuthStore();
                authStore.clearAuth();
            } catch {
                // Pinia may not be available yet in some contexts
            }

            if (!window.location.pathname.startsWith('/login')) {
                window.location.href = '/login';
            }
            return Promise.reject(error);
        }

        // Mark original request as retried
        originalRequest._retry = true;

        try {
            // Deduplicate concurrent refreshes using shared promise
            if (!refreshPromise) {
                refreshPromise = (async () => {
                    try {
                        const { data } = await axios.post<AuthResponse>(
                            `${API_BASE_URL}/authentication/refresh`,
                            {},
                            { withCredentials: true }
                        );

                        const authStore = useAuthStore();
                        authStore.setAuth(data.token, data.user);
                        return data.token;
                    } catch (refreshErr) {
                        const authStore = useAuthStore();
                        authStore.clearAuth();

                        if (!window.location.pathname.startsWith('/login')) {
                            window.location.href = '/login';
                        }
                        throw refreshErr;
                    } finally {
                        refreshPromise = null;
                    }
                })();
            }

            const newAccessToken = await refreshPromise;

            // Update headers and retry original request once
            if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
            }

            return apiClient(originalRequest);
        } catch (refreshErr) {
            return Promise.reject(refreshErr);
        }
    }
);

export default apiClient;