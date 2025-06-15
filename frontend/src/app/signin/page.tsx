"use client";
import { useSession, signIn, signOut } from "next-auth/react";

export default function Page() {
  const { data: session } = useSession();

  if (session) {
    return (
      <div>
        <p>Signed in as {session.user?.email}</p>
        <button onClick={() => signOut()}>Sign out</button>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center h-screen w-[90%] mx-auto">
      <button
        className="px-6 py-2 bg-black text-white rounded-lg font-bold transform hover:-translate-y-1 transition duration-400"
        onClick={() => signIn("google")}
      >
        Sign in with Google
      </button>
    </div>
  );
}
