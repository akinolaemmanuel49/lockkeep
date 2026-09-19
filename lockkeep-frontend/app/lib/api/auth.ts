import { config } from "~/config";
import type { KDFParams, User, LocalLoginRequest, LocalRegisterRequest } from "~/types";
import { authFetch } from "./core";

export async function oauthCallback(
    accessToken: string,
): Promise<{ access_token: string; refresh_token: string; user: User }> {
    const res = await fetch(`${config.LOCKKEEP_API_URI}/auth/oauth`, {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${accessToken}`,
        },
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "OAuth failed");
    }

    return res.json();
}

export async function localRegisterUser(
    newUser: LocalRegisterRequest,
): Promise<{ access_token: string; refresh_token: string; user: User }> {
    const res = await fetch(`${config.LOCKKEEP_API_URI}/auth/register`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: newUser.email, password: newUser.password }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to signup");
    }

    return res.json();
}

export async function localLogin(
    authCredentials: LocalLoginRequest,
): Promise<{ access_token: string; refresh_token: string; user: User }> {
    const res = await fetch(`${config.LOCKKEEP_API_URI}/auth/login`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            email: authCredentials.email,
            password: authCredentials.password,
        }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to login");
    }

    return res.json();
}

export async function setVerificationHash(
    verificationHash: string,
    kdfParams: KDFParams,
): Promise<{ user: User }> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/auth/vault/create`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            verification_hash: verificationHash,
            kdf_params: kdfParams,
        }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to set vault password");
    }

    return res.json();
}

export async function fetchKDFParams(): Promise<KDFParams> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/auth/vault/kdfparams`, {
        method: "GET",
        headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to retrieve KDFParams");
    }

    return res.json();
}

export async function updateEmail(
    newEmail: string,
): Promise<{ access_token: string; refresh_token: string; user: User }> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/auth/update/email`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: newEmail }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to change email address");
    }

    return res.json();
}

export async function updateAccountPassword(
    currentPassword: string,
    newPassword: string,
): Promise<void> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/auth/update/password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ currentPassword, newPassword }),
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to change password");
    }
}