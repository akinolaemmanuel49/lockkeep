type Config = {
    AUTH0_DOMAIN: string;
    AUTH0_CLIENT_ID: string;
    AUTH0_REDIRECT_URI: string;
    LOCKKEEP_API_URI: string;
};

export const config: Config = {
    AUTH0_DOMAIN: import.meta.env.VITE_AUTH0_DOMAIN,
    AUTH0_CLIENT_ID: import.meta.env.VITE_AUTH0_CLIENT_ID,
    AUTH0_REDIRECT_URI: import.meta.env.VITE_AUTH0_REDIRECT_URI,
    LOCKKEEP_API_URI: import.meta.env.VITE_LOCKKEEP_API_URI,
};