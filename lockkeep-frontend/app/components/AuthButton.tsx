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
      className="lk-btn lk-btn--primary mt-2 w-full py-3 text-sm"
    >
      {ctx === "signup"
        ? isLoading
          ? "Forging account..."
          : "Create Account"
        : isLoading
          ? "Authenticating..."
          : "Sign In"}
    </button>
  );
}