import { useSession } from "next-auth/react";
import { useEffect, useState } from "react";

export function useAuth() {
  const { data: session, status } = useSession();
  const [jwt, setJwt] = useState<string | null>(null);

  useEffect(() => {
    if (session?.accessToken) {
      fetch(process.env.NEXT_PUBLIC_API_URL + "/api/signin", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${session.accessToken}`,
        },
      })
        .then((res) => res.json())
        .then((data) => setJwt(data.token))
        .catch((error) => {
          console.error("Error fetching JWT:", error);
          window.location.href = '/signin';
          setJwt(null);
        });
    } else {
      setJwt(null);
    }
  }, [session?.accessToken]);
  
  return {
    accessToken: session?.accessToken,
    jwt,
    isAuthenticated: status === "authenticated",
    isLoading: status === "loading",
    user: session?.user,
  };
} 