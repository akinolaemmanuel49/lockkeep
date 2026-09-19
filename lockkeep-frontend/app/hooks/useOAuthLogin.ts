import { useState } from "react";
import { getAuth0Url, type OAuthProvider } from "~/utils/getAuth0Url";

export function useOAuthLogin() {
  const [isLoading, setIsLoading] = useState(false);

  const handleOAuth = (provider: OAuthProvider) => {
    setIsLoading(true);
    const url = getAuth0Url(provider);
    window.location.href = url;
  };

  return {
    handleOAuth,
    isLoading,
  };
}