import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import axios from 'axios';
import apiClient from '../services/api';

const API_BASE_URL = 'http://localhost:8080/api/v1';

// --- Interfaces ---
export interface User {
    id: string;
    email: string;
    name: string;
    role: 'admin' | 'member';
    avatar_url?: string | null;
}

export interface AuthResponse {
    token: string;
    user: User;
}

export interface OTPRequestPayload {
    email: string;
}

export interface VerifyOTPPayload {
    email: string;
    otp: string;
}

// In-memory token reference accessible synchronously across modules
let inMemoryAccessToken: string | null = null;

export function getAccessToken(): string | null {
    return inMemoryAccessToken;
}

export function setAccessToken(token: string | null): void {
    inMemoryAccessToken = token;
}

// --- Pinia Store ---
export const useAuthStore = defineStore('auth', () => {
    // State: strictly in-memory (no tokens in localStorage)
    const token = ref<string | null>(inMemoryAccessToken);
    const user = ref<User | null>(null);
    const isInitialized = ref(false);

    // Getters (Computed)
    const isAuthenticated = computed(() => !!token.value);

    function setAuth(newToken: string, newUser: User): void {
        token.value = newToken;
        user.value = newUser;
        setAccessToken(newToken);
    }

    function clearAuth(): void {
        token.value = null;
        user.value = null;
        setAccessToken(null);
    }

    // Actions (Methods)
    async function requestOTP(email: string): Promise<void> {
        await apiClient.post('/authentication/request-otp', { email } as OTPRequestPayload);
    }

    async function verifyOTP(email: string, code: string): Promise<void> {
        const { data } = await apiClient.post<AuthResponse>(
            '/authentication/verify-otp',
            { email, otp: code } as VerifyOTPPayload
        );

        setAuth(data.token, data.user);
    }

    async function refreshToken(): Promise<string> {
        const { data } = await axios.post<AuthResponse>(
            `${API_BASE_URL}/authentication/refresh`,
            {},
            { withCredentials: true }
        );

        setAuth(data.token, data.user);
        return data.token;
    }

    async function logout(): Promise<void> {
        try {
            await apiClient.post('/authentication/logout');
        } catch {
            // Ignore failure on logout
        } finally {
            clearAuth();
        }
    }

    return {
        token,
        user,
        isInitialized,
        isAuthenticated,
        setAuth,
        clearAuth,
        requestOTP,
        verifyOTP,
        refreshToken,
        logout,
    };
});