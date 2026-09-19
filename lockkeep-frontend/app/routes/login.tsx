import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { useAuth } from "~/providers/auth";
import Field from "~/components/Field";
import OAuthButton from "~/components/OAuthButton";
import { useOAuthLogin } from "~/hooks/useOAuthLogin";
import { AuthButton } from "~/components/AuthButton";
import { useToast } from "~/providers/toast";
import { requireGuest } from "~/lib/auth-guard";
import type { Route } from "./+types/login";
import { localLogin } from "~/lib/api/auth";


export const clientLoader = () => {
  return requireGuest();
};

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Sign In - LockKeep" },
    {
      name: "description",
      content:
        "Sign in to your LockKeep vault. Access your encrypted credentials securely.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const { addToast } = useToast();
  const navigate = useNavigate();

  const {
    handleOAuth,
    isLoading: isOAuthLoading,
  } = useOAuthLogin();

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const result = await localLogin({ email, password });
      login(result.user, result.access_token);
      navigate(result.user.hasMasterPassword ? "/dashboard" : "/setup");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to sign in",
        "error",
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900 p-8">
        <div className="mb-6 text-center">
          <h1 className="text-2xl font-bold text-slate-100">Sign In</h1>
          <p className="mt-1 text-sm text-slate-500">
            Welcome back to your vault
          </p>
        </div>

        <div className="mb-6 flex flex-col gap-3">
          <OAuthButton
            provider="google"
            onClick={() => handleOAuth("google")}
            isLoading={isOAuthLoading}
          />
          <OAuthButton
            provider="github"
            onClick={() => handleOAuth("github")}
            isLoading={isOAuthLoading}
          />
        </div>

        <div className="relative mb-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-slate-800" />
          </div>
          <div className="relative flex justify-center text-xs">
            <span className="bg-slate-900 px-3 text-slate-600">or</span>
          </div>
        </div>

        <form onSubmit={handleEmailSubmit} className="flex flex-col gap-4">
          <Field
            label="Email"
            type="email"
            value={email}
            onChange={setEmail}
            required
          />
          <Field
            label="Password"
            type="password"
            value={password}
            onChange={setPassword}
            required
          />

          <AuthButton ctx="signin" isLoading={isLoading} />
        </form>

        <p className="mt-6 text-center text-sm text-slate-500">
          Don't have an account?{" "}
          <Link
            to="/signup"
            className="font-medium text-sky-400 hover:text-sky-300"
          >
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
