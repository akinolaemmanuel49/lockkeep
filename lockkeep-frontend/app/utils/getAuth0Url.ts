import { config } from "~/config";

export type OAuthProvider = "google" | "github";

export function getAuth0Url(provider: OAuthProvider): string {
    const params = new URLSearchParams({
        response_type: "token",
        client_id: config.AUTH0_CLIENT_ID,
        redirect_uri: config.AUTH0_REDIRECT_URI,
        scope: 'openid profile email',
        connection: provider === "google" ? "google-oauth2" : "github",

    });

    return `https://${config.AUTH0_DOMAIN}/authorize?${params.toString()}`;
}

// export function loginWithGoogle() {
//     window.location.href = getAuth0Url("google");
// }

// export function loginWithGitHub() {
//     window.location.href = getAuth0Url("github");
// }