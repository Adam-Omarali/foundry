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

  // If we have a session token, validate it against the API
  let isTokenValid = false
  if (session?.accessToken) {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/signin`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${session.accessToken}`,
        },
      })
      isTokenValid = response.ok
      console.log('API token validation:', isTokenValid ? 'Valid' : 'Invalid')
    } catch (error) {
      console.error('Error validating token with API:', error)
      isTokenValid = false
    }
  }

  console.log('isTokenValid:', isTokenValid)
  console.log('session:', session)

  // If no valid session, redirect to signin (except for public paths)
  if (!isPublicPath && (!session || !isTokenValid)) {
    console.log('Redirecting unauthenticated user to signin')
    return NextResponse.redirect(new URL('/signin', request.url))
  }

  // Redirect authenticated users away from public paths when they have valid tokens
  // This allows failed signin attempts to stay on the signin page
  if ((path === '/signin' || path === '/signup') && session && isTokenValid) {
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