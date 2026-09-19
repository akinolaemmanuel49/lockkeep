import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { useAuth } from "~/providers/auth";
import Field from "~/components/Field";
import OAuthButton from "~/components/OAuthButton";
import { useOAuthLogin } from "~/hooks/useOAuthLogin";
import { AuthButton } from "~/components/AuthButton";
import { useToast } from "~/providers/toast";
import { requireGuest } from "~/lib/auth-guard";
import type { Route } from "./+types/signup";
import { localRegisterUser } from "~/lib/api/auth";

export const clientLoader = () => {
  return requireGuest();
};

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Create Account - LockKeep" },
    {
      name: "description",
      content:
        "Create your LockKeep account. Start your zero-trust password vault today.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function Signup() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const { addToast } = useToast();
  const navigate = useNavigate();

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      addToast("Passwords do not match", "error");
      return;
    }

    if (password.length < 8) {
      addToast("Password must be at least 8 characters", "error");
      return;
    }

    setIsLoading(true);
    try {
      const result = await localRegisterUser({ email, password });
      login(result.user, result.access_token);
      addToast("Account created successfully", "success");
      navigate("/setup");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to create account",
        "error",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const {
    handleOAuth,
    isLoading: isOAuthLoading,
  } = useOAuthLogin();


  return (
    <div className="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900 p-8">
        <div className="mb-6 text-center">
          <h1 className="text-2xl font-bold text-slate-100">Create Account</h1>
          <p className="mt-1 text-sm text-slate-500">
            Start your zero-trust vault
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
            placeholder="At least 8 characters"
          />
          <Field
            label="Confirm Password"
            type="password"
            value={confirmPassword}
            onChange={setConfirmPassword}
            required
          />

          <AuthButton ctx="signup" isLoading={isLoading} />
        </form>

        <p className="mt-6 text-center text-sm text-slate-500">
          Already have an account?{" "}
          <Link
            to="/login"
            className="font-medium text-sky-400 hover:text-sky-300"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
