import { clearSession, getAccessToken, setAccessToken } from "~/utils/sessionStorage";
import { config } from "~/config";

export async function refreshToken(): Promise<boolean> {
    try {
        const res = await fetch(`${config.LOCKKEEP_API_URI}/auth/refresh`, {
            method: "POST",
            credentials: "include",
        });

        if (!res.ok) return false;

        const data = await res.json();
        setAccessToken(data.access_token);
        return true;
    } catch {
        return false;
    }
}

export async function authFetch(
    url: string,
    options: RequestInit = {},
): Promise<Response> {
    let token = getAccessToken();

    let response = await fetch(url, {
        ...options,
        credentials: "include",
        headers: {
            ...options.headers,
            Authorization: `Bearer ${token}`,
        },
    });

    if (response.status !== 401) {
        return response;
    }

    const refreshed = await refreshToken();
    if (!refreshed) {
        clearSession();
        window.location.href = "/login";
        throw new Error("Session expired");
    }

    token = getAccessToken();
    response = await fetch(url, {
        ...options,
        credentials: "include",
        headers: {
            ...options.headers,
            Authorization: `Bearer ${token}`,
        },
    });

    return response;
}