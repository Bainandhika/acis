import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import axios from 'axios';
import apiClient, { getApiBaseUrl } from '../services/api';

const API_BASE_URL = getApiBaseUrl();

// --- Interfaces ---
export interface User {
    id: string;
    email: string;
    phone_number: string;
    name: string;
    role: 'admin' | 'member';
    avatar_url?: string | null;
    telegram_chat_id?: number | null;
}

export interface AuthResponse {
    token: string;
    user: User;
}

export interface OTPRequestPayload {
    email: string;
    telegram_identifier: string;
    phone_number?: string;
}

export interface OTPRequestResponse {
    message: string;
    auth_session: string;
    telegram_bot_username?: string;
    direct_sent: boolean;
    is_test_user?: boolean;
    test_otp?: string;
}

export interface VerifyOTPPayload {
    email: string;
    telegram_identifier: string;
    phone_number?: string;
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
    async function requestOTP(email: string, telegramIdentifier: string): Promise<OTPRequestResponse> {
        const { data } = await apiClient.post<OTPRequestResponse>(
            '/authentication/request-otp',
            { 
                email, 
                telegram_identifier: telegramIdentifier,
                phone_number: telegramIdentifier 
            } as OTPRequestPayload
        );
        return data;
    }

    async function verifyOTP(email: string, telegramIdentifier: string, code: string): Promise<void> {
        const { data } = await apiClient.post<AuthResponse>(
            '/authentication/verify-otp',
            { 
                email, 
                telegram_identifier: telegramIdentifier,
                phone_number: telegramIdentifier,
                otp: code 
            } as VerifyOTPPayload
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