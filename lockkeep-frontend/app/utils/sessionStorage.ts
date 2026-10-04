export const ACCESS_TOKEN_KEY = "access_token";
export const USER_KEY = "vault_user";

export function getSessionItem(key: string): string | null {
    return sessionStorage.getItem(key);
}

export function setSessionItem(key: string, value: string): void {
    sessionStorage.setItem(key, value);
}

export function removeSessionItem(key: string): void {
    sessionStorage.removeItem(key);
}

export function clearSession(): void {
    sessionStorage.clear();
}

export const getAccessToken = () =>
    getSessionItem(ACCESS_TOKEN_KEY);

export const setAccessToken = (token: string) =>
    setSessionItem(ACCESS_TOKEN_KEY, token);

export const removeAccessToken = () =>
    removeSessionItem(ACCESS_TOKEN_KEY);