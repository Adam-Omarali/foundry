import { useSession } from "next-auth/react";
import { useEffect, useState } from "react";

export function useAuth() {
  const { data: session, status } = useSession();
  const [jwt, setJwt] = useState<string | null>(null);
  const [isLoadingJWT, setIsLoadingJWT] = useState(false);

  useEffect(() => {
    const fetchJWT = async () => {
      if (session?.accessToken && !jwt && !isLoadingJWT) {
        setIsLoadingJWT(true);
        try {
          const response = await fetch(process.env.NEXT_PUBLIC_API_URL + "/api/signin", {
            method: "GET",
            headers: {
              "Authorization": `Bearer ${session.accessToken}`,
            },
          });

          if (response.ok) {
            const data = await response.json();
            setJwt(data.token);
          } else {
            console.error('Failed to get JWT token');
          }
        } catch (error) {
          console.error('Error fetching JWT:', error);
        } finally {
          setIsLoadingJWT(false);
        }
      }
    };

    fetchJWT();
  }, [session?.accessToken, jwt, isLoadingJWT]);

  return {
    accessToken: session?.accessToken,
    jwt,
    isAuthenticated: status === "authenticated" && !!jwt,
    isLoading: status === "loading" || isLoadingJWT,
    user: session?.user,
  };
} 