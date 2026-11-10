import GoogleProvider from "next-auth/providers/google";
import type { NextAuthOptions } from "next-auth";

// What the browser can see via useSession(). Only the Foundry session JWT is
// exposed; Google tokens never leave the server.
declare module "next-auth" {
  interface Session {
    foundryToken?: string;
    error?: "FoundrySignInFailed";
    user: {
      id: string;
      name?: string | null;
      email?: string | null;
      image?: string | null;
    };
  }
}

// What NextAuth stores in its encrypted session cookie.
declare module "next-auth/jwt" {
  interface JWT {
    foundryToken?: string;
    foundryExpiresAt?: number; // unix seconds
    error?: "FoundrySignInFailed";
  }
}

// This code runs on the Next.js server, not in the browser. Inside Docker the
// server reaches the API over the internal network (e.g. http://api:8080);
// the browser uses the public NEXT_PUBLIC_API_URL.
const SERVER_API_URL = process.env.API_INTERNAL_URL ?? process.env.NEXT_PUBLIC_API_URL;

type FoundrySession = { token: string; email: string; expires_at: number };

async function exchangeGoogleIdToken(idToken: string): Promise<FoundrySession> {
  const res = await fetch(`${SERVER_API_URL}/api/auth/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ id_token: idToken }),
  });
  if (!res.ok) {
    throw new Error(`Foundry sign-in failed: ${res.status} ${await res.text()}`);
  }
  return res.json();
}

const SESSION_MAX_AGE = 30 * 24 * 60 * 60; // matches SessionTTL in backend/utils/jwt.go

export const authOptions: NextAuthOptions = {
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
      authorization: { params: { scope: "openid email profile" } },
    }),
  ],
  session: { strategy: "jwt", maxAge: SESSION_MAX_AGE },
  pages: { signIn: "/signin" },
  callbacks: {
    // `account` is only present on the request right after Google redirects
    // back. That's the single moment we trade Google's ID token for a Foundry
    // session; every later request just reads the cookie.
    async jwt({ token, account }) {
      if (account?.id_token) {
        try {
          const s = await exchangeGoogleIdToken(account.id_token);
          token.foundryToken = s.token;
          token.foundryExpiresAt = s.expires_at;
          delete token.error;
        } catch (err) {
          console.error(err);
          delete token.foundryToken;
          delete token.foundryExpiresAt;
          token.error = "FoundrySignInFailed";
        }
      }
      return token;
    },
    async session({ session, token }) {
      session.foundryToken = token.foundryToken;
      session.error = token.error;
      return session;
    },
  },
  secret: process.env.NEXTAUTH_SECRET,
};
