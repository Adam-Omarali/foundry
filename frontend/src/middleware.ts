import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { getToken } from "next-auth/jwt";

// Optimistic check only: decrypt the session cookie locally and look at the
// Foundry token's expiry. No network calls. The Go API is what actually
// enforces auth on every request.
export async function middleware(request: NextRequest) {
  const path = request.nextUrl.pathname;
  const isAuthPage = path === "/signin" || path === "/signup";
  const isPublicPath = isAuthPage || path === "/policy";

  const token = await getToken({ req: request, secret: process.env.NEXTAUTH_SECRET });
  const isSignedIn =
    !!token?.foundryToken && (token.foundryExpiresAt ?? 0) * 1000 > Date.now();

  if (!isPublicPath && !isSignedIn) {
    return NextResponse.redirect(new URL("/signin", request.url));
  }
  if (isAuthPage && isSignedIn) {
    return NextResponse.redirect(new URL("/", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|public).*)"],
};
