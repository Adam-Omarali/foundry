/*
Foundry extension auth, shared by background.js (via importScripts) and
popup.html (via <script>).

Flow: Google OpenID Connect via launchWebAuthFlow → Google ID token →
POST /api/auth/google → Foundry session JWT stored in chrome.storage.local.
*/
const DEV_MODE = true;
const API_URL = DEV_MODE ? "http://localhost:8080" : "https://foundry-production-7176.up.railway.app";

const GOOGLE_CLIENT_ID = "634531581191-h4u6sth79fvsobc80r6ube6rjv5t9j50.apps.googleusercontent.com";

// Decode a JWT's payload (no verification; the backend verifies the signature).
function decodeJwtPayload(jwt) {
  const b64 = jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
  return JSON.parse(atob(b64.padEnd(b64.length + ((4 - (b64.length % 4)) % 4), "=")));
}

// Resolves with { token, email, expires_at } and saves it as `user` in storage.
function signInWithGoogle() {
  const nonce = crypto.randomUUID();
  const authUrl = "https://accounts.google.com/o/oauth2/v2/auth?" + new URLSearchParams({
    client_id: GOOGLE_CLIENT_ID,
    response_type: "id_token",
    redirect_uri: chrome.identity.getRedirectURL(),
    scope: "openid email",
    nonce,
    prompt: "select_account",
  });

  return new Promise((resolve, reject) => {
    chrome.identity.launchWebAuthFlow({ url: authUrl, interactive: true }, (redirectUrl) => {
      if (chrome.runtime.lastError || !redirectUrl) {
        reject(new Error(chrome.runtime.lastError?.message || "Sign-in cancelled"));
        return;
      }
      const params = new URLSearchParams(new URL(redirectUrl).hash.substring(1));
      const idToken = params.get("id_token");
      if (!idToken) {
        reject(new Error(params.get("error") || "No id_token returned"));
        return;
      }
      // The nonce ties this ID token to the request we just made.
      if (decodeJwtPayload(idToken).nonce !== nonce) {
        reject(new Error("Nonce mismatch"));
        return;
      }
      resolve(idToken);
    });
  })
    .then((idToken) =>
      fetch(`${API_URL}/api/auth/google`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id_token: idToken }),
      })
    )
    .then((res) => {
      if (!res.ok) throw new Error(`Server sign-in failed (${res.status})`);
      return res.json();
    })
    .then((user) => new Promise((resolve) => chrome.storage.local.set({ user }, () => resolve(user))));
}
