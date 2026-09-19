import { useEffect } from "react";
import { useNavigate } from "react-router";
import { oauthCallback } from "~/lib/api/auth";
import { useAuth } from "~/providers/auth";
import { useToast } from "~/providers/toast";

export default function Callback() {
    const navigate = useNavigate();
    const { login } = useAuth();
    const { addToast } = useToast();

    useEffect(() => {
        const hash = window.location.hash;
        const params = new URLSearchParams(hash.slice(1));
        const accessToken = params.get("access_token");

        if (!accessToken) {
            addToast("Authentication failed: no token received", "error");
            navigate("/login");
            return;
        }

        oauthCallback(accessToken)
            .then((data) => {
                login(data.user, data.access_token);
                addToast("Signed in successfully", "success");
                navigate(data.user.hasMasterPassword ? "/unlock" : "/setup");
            })
            .catch((err) => {
                addToast(err instanceof Error ? err.message : "OAuth failed", "error");
                navigate("/login");
            });
    }, [navigate, login, addToast]);

    return (
        <div className="flex min-h-[calc(100vh-8rem)] items-center justify-center">
            <div className="flex items-center gap-3 text-slate-500">
                <div className="h-5 w-5 animate-spin rounded-full border-2 border-slate-700 border-t-sky-400" />
                <span>Completing sign in...</span>
            </div>
        </div>
    );
}