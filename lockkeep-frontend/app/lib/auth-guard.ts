import { redirect } from "react-router";
import type { User } from "~/types";

const USER_KEY = "vault_user";

export function getCurrentUser(): User | null {
    try {
        const stored = sessionStorage.getItem(USER_KEY);
        return stored ? JSON.parse(stored) : null;
    } catch {
        return null;
    }
}

export function requireAuth() {
    const user = getCurrentUser();
    if (!user) {
        return redirect("/login");
    }
    return null;
}

export function requireGuest() {
    const user = getCurrentUser();
    if (user) {
        return redirect("/dashboard");
    }
    return null;
}