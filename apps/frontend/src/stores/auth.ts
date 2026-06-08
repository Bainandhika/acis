import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import apiClient from '../services/api';

// --- Interfaces (Mirip Go Structs) ---
export interface User {
    id: string;
    email: string;
    name: string;
    role: 'admin' | 'member'; // Literal types biar lebih strict
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
    code: string;
}

// --- Pinia Store ---
export const useAuthStore = defineStore('auth', () => {
    // State
    const token = ref<string | null>(localStorage.getItem('acis_token') || null);
    const user = ref<User | null>(
        JSON.parse(localStorage.getItem('acis_user') || 'null')
    );

    // Getters (Computed)
    const isAuthenticated = computed(() => !!token.value);

    // Actions (Methods)
    async function requestOTP(email: string): Promise<void> {
        await apiClient.post('/auth/request-otp', { email } as OTPRequestPayload);
    }

    async function verifyOTP(email: string, code: string): Promise<void> {
        const { data } = await apiClient.post<AuthResponse>(
            '/auth/verify-otp',
            { email, code } as VerifyOTPPayload
        );

        token.value = data.token;
        user.value = data.user;

        localStorage.setItem('acis_token', token.value);
        localStorage.setItem('acis_user', JSON.stringify(user.value));
    }

    function logout(): void {
        token.value = null;
        user.value = null;
        localStorage.removeItem('acis_token');
        localStorage.removeItem('acis_user');
    }

    return { token, user, isAuthenticated, requestOTP, verifyOTP, logout };
});