import { config } from "~/config";
import { authFetch } from "./core";
import type { CryptoPolicy, SetCurrentPolicyRequest } from "~/types";


export async function getCurrentPolicy(): Promise<CryptoPolicy> {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/crypto-policy/current`, {
        method: "GET",
        headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to fetch policy");
    }

    return res.json();
}

export async function setCurrentPolicy(setCurrentPolicy: SetCurrentPolicyRequest) {
    const res = await authFetch(`${config.LOCKKEEP_API_URI}/crypto-policy/current`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(setCurrentPolicy)
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || "Failed to update policy");
    }

    return res.json();
}