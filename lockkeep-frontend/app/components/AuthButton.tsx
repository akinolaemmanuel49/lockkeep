export function AuthButton({
  ctx,
  isLoading,
}: {
  ctx: "signup" | "signin";
  isLoading: boolean;
}) {
  return (
    <button
      type="submit"
      disabled={isLoading}
      className="mt-2 rounded-lg bg-sky-400 py-3 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50 cursor-pointer"
    >
      {ctx === "signup"
        ? isLoading
          ? "Creating account..."
          : "Create Account"
        : isLoading
          ? "Signing in..."
          : "Sign In"}
    </button>
  );
}
