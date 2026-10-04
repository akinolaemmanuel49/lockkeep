import { useEffect } from "react";
import { useNavigate } from "react-router";
import { oauthCallback } from "~/lib/api/auth";
import { useAuth } from "~/providers/auth";
import { useToast } from "~/providers/toast";
import VaultDial from "~/components/VaultDial";

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
        navigate(data.user.hasMasterPassword ? "/dashboard" : "/setup");
      })
      .catch((err) => {
        addToast(
          err instanceof Error ? err.message : "OAuth failed",
          "error",
        );
        navigate("/login");
      });
  }, [navigate, login, addToast]);

  return (
    <div className="flex min-h-[calc(100vh-9rem)] items-center justify-center">
      <div className="flex flex-col items-center gap-6 text-sand-500">
        <VaultDial size={120} spinning />
        <div className="flex items-center gap-3">
          <span className="text-sm uppercase tracking-[0.22em]">
            Verifying clearance...
          </span>
        </div>
      </div>
    </div>
  );
}