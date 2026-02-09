import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface User {
    username: string;
    role: string;
    name?: string;
    department?: string;
}

interface AuthState {
    isLoggedIn: boolean;
    token: string | null;
    user: User | null;
}

const initialState: AuthState = {
    isLoggedIn: false,
    token: null,
    user: null
};

function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>(initialState);

    return {
        subscribe,
        // UBAH NAMA dari setLogin -> login
        login: (token: string, user: User) => {
            if (browser) {
                localStorage.setItem('auth_token', token);
                localStorage.setItem('auth_user', JSON.stringify(user));
            }
            set({ isLoggedIn: true, token, user });
        },
        // UBAH NAMA dari setLogout -> logout
        logout: () => {
            if (browser) {
                localStorage.removeItem('auth_token');
                localStorage.removeItem('auth_user');
            }
            set(initialState);
        },
        initialize: () => {
            if (browser) {
                const token = localStorage.getItem('auth_token');
                const userStr = localStorage.getItem('auth_user');
                if (token && userStr) {
                    try {
                        const user = JSON.parse(userStr);
                        set({ isLoggedIn: true, token, user });
                    } catch (e) {
                        localStorage.removeItem('auth_token');
                        localStorage.removeItem('auth_user');
                        set(initialState);
                    }
                }
            }
        }
    };
}

export const auth = createAuthStore();

// Export helper untuk kompatibilitas dengan kode lain yang mungkin import { login }
export const login = (token: string, user: User) => auth.login(token, user);
export const logout = () => auth.logout();