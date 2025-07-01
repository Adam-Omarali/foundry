import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { getToken } from 'next-auth/jwt'

// This function can be marked `async` if using `await` inside
export async function middleware(request: NextRequest) {
  // Get the pathname of the request (e.g. /, /protected, /sign-in)
  const path = request.nextUrl.pathname
  console.log('Middleware running for path:', path)

  // Define public paths that don't require authentication
  const isPublicPath = path === '/signin' || path === '/signup'

  // Get the session token using NextAuth
  const session = await getToken({ 
    req: request,
    secret: process.env.NEXTAUTH_SECRET
  })
  console.log('Session status:', session ? 'Authenticated' : 'Not authenticated')

  // If no session, redirect to signin (except for public paths)
  if (!isPublicPath && !session) {
    console.log('Redirecting unauthenticated user to signin')
    return NextResponse.redirect(new URL('/signin', request.url))
  }

  // Redirect authenticated users away from public paths
  if (isPublicPath && session) {
    console.log('Redirecting authenticated user from public path to home')
    return NextResponse.redirect(new URL('/', request.url))
  }

  return NextResponse.next()
}

// Configure which paths the middleware should run on
export const config = {
  matcher: [
    '/',
    '/signin',
    '/signup',
    '/((?!api|_next/static|_next/image|favicon.ico|public).*)',
  ],
} 