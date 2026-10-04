import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { useAuth } from "~/providers/auth";
import Field from "~/components/Field";
import OAuthButton from "~/components/OAuthButton";
import { useOAuthLogin } from "~/hooks/useOAuthLogin";
import { AuthButton } from "~/components/AuthButton";
import AuthPlate from "~/components/AuthPlate";
import { useToast } from "~/providers/toast";
import { requireGuest } from "~/lib/auth-guard";
import type { Route } from "./+types/signup";
import { localRegisterUser } from "~/lib/api/auth";

export const clientLoader = () => {
  return requireGuest();
};

export function meta({}: Route.MetaArgs) {
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
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const { addToast } = useToast();
  const navigate = useNavigate();

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (username.length < 6) {
      addToast("Username must be at least 6 characters", "error");
      return;
    }

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
      const result = await localRegisterUser({ username, email, password });
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

  const { handleOAuth, isLoading: isOAuthLoading } = useOAuthLogin();

  return (
    <div className="flex min-h-[calc(100vh-9rem)] items-center justify-center px-4">
      <AuthPlate
        title="Create Account"
        eyebrow=""
        description="Begin your zero-trust vault."
      >
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
            <div className="w-full border-t border-brass-500/15" />
          </div>
          <div className="relative flex justify-center text-xs">
            <span className="bg-coal-900 px-3 text-sand-500">or</span>
          </div>
        </div>

        <form onSubmit={handleEmailSubmit} className="flex flex-col gap-4">
          <Field
            label="Username"
            type="text"
            value={username}
            onChange={setUsername}
            required
            placeholder="At least 6 characters"
            autoComplete="username"
          />
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

        <p className="mt-6 text-center text-sm text-sand-500">
          Already have access?{" "}
          <Link
            to="/login"
            className="font-medium text-brass-400 hover:text-brass-300"
          >
            Sign in
          </Link>
        </p>
      </AuthPlate>
    </div>
  );
}