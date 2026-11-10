import { useSession } from "next-auth/react";

// The Foundry JWT lives in the NextAuth session, obtained once at sign-in.
// No fetch here: refreshing the page doesn't re-sign-in.
export function useAuth() {
  const { data: session, status } = useSession();
  const jwt = session?.foundryToken ?? null;

  return {
    jwt,
    isAuthenticated: status === "authenticated" && !!jwt,
    isLoading: status === "loading",
    error: session?.error,
    user: session?.user,
  };
}
