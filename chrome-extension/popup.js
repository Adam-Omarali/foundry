/*
Foundry Extension - Google Auth via launchWebAuthFlow
*/

document.getElementById("loginBtn").addEventListener("click", () => {
  const clientId = "634531581191-h4u6sth79fvsobc80r6ube6rjv5t9j50.apps.googleusercontent.com";
  const redirectUri = chrome.identity.getRedirectURL();
  const scopes = ["https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"];
  const authUrl = `https://accounts.google.com/o/oauth2/auth?client_id=${clientId}` +
                  `&response_type=token` +
                  `&redirect_uri=${encodeURIComponent(redirectUri)}` +
                  `&scope=${encodeURIComponent(scopes.join(" "))}` +
                  `&prompt=consent`;

  console.log("Auth URL:", authUrl);

  chrome.identity.launchWebAuthFlow({
    url: authUrl,
    interactive: true
  }, function (redirectUrl) {
    if (chrome.runtime.lastError || !redirectUrl) {
      console.error("Auth Error:", chrome.runtime.lastError?.message);
      return;
    }

    const params = new URLSearchParams(new URL(redirectUrl).hash.substring(1));
    const accessToken = params.get("access_token");
    console.log("Access token:", accessToken);

    fetch("http://localhost:8080/api/signin", {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${accessToken}`,
      },
    }).then((res) => res.json()).then((user) => {
      console.log("User:", user);
      // Store the user data with JWT token
      chrome.storage.local.set({ user: user }, function() {
        console.log('User data saved');
      });
      document.getElementById("status").innerText = `Signed in as ${user.email}`;
    });

    document.getElementById("loginBtn").style.display = 'none';
    document.getElementById("rememberSiteBtn").disabled = false;

    // fetch("https://www.googleapis.com/oauth2/v2/userinfo", {
    //   headers: {
    //     Authorization: `Bearer ${accessToken}`,
    //   },
    // })
    //   .then((res) => res.json())
    //   .then((user) => {
    //     console.log("User info:", user);
    //     document.getElementById("status").innerText = `Signed in as ${user.email}`;
    //   });
  });
});

document.addEventListener('DOMContentLoaded', function() {
  const rememberSiteBtn = document.getElementById('rememberSiteBtn');
  const loginBtn = document.getElementById('loginBtn');
  const statusElement = document.getElementById('status');
  
  // Initially disable the remember site button
  if (rememberSiteBtn) {
    rememberSiteBtn.disabled = true;
  }
  
  statusElement.textContent = 'Please sign in to use the extension';

  // Check for stored user data
  chrome.storage.local.get(['user'], function(result) {
    console.log('Stored user data:', result.user);
  });

  // Check for stored JWT token
  chrome.storage.local.get(['user'], function(result) {
    if (result.user && result.user.token) {
      // Verify the token is still valid
      fetch("http://localhost:8080/api/verifytoken", {
        method: "GET",
        headers: {
          "Authorization": `Bearer ${result.user.token}`,
        },
      }).then((res) => {
        if (res.ok) {
          // Token is valid
          if (loginBtn) {
            loginBtn.style.display = 'none';
          }
          if (rememberSiteBtn) {
            rememberSiteBtn.disabled = false;
          }
          statusElement.innerText = `Signed in as ${result.user.email}`;
        } else {
          // Token is invalid or expired
          chrome.storage.local.remove(['user'], function() {
            console.log('Invalid token, cleared user data');
            if (loginBtn) {
              loginBtn.style.display = 'block';
            }
            if (rememberSiteBtn) {
              rememberSiteBtn.disabled = true;
            }
            statusElement.textContent = 'Session expired. Please sign in again.';
          });
        }
      }).catch((error) => {
        console.error('Error verifying token:', error);
        // On error, clear storage and show login
        chrome.storage.local.remove(['user'], function() {
          if (loginBtn) {
            loginBtn.style.display = 'block';
          }
          if (rememberSiteBtn) {
            rememberSiteBtn.disabled = true;
          }
          statusElement.textContent = 'Error verifying session. Please sign in again.';
        });
      });
    } else {
      // No JWT token, check Google auth
      chrome.identity.getAuthToken({ interactive: false }, function (token) {
        if (chrome.runtime.lastError || !token) {
          // No token exists, show login button
          console.log("No token exists, showing login button");
          if (loginBtn) {
            loginBtn.style.display = 'block';
          }
          if (rememberSiteBtn) {
            rememberSiteBtn.disabled = true;
          }
          statusElement.textContent = 'Please sign in to use the extension';
        } else {
          // Token exists, hide login button and update status
          if (loginBtn) {
            loginBtn.style.display = 'none';
          }
          if (rememberSiteBtn) {
            rememberSiteBtn.disabled = false;
          }
          // Use the existing token to get user info
          fetch("http://localhost:8080/api/signin", {
            method: "GET",
            headers: {
              "Authorization": `Bearer ${token}`,
            },
          }).then((res) => res.json()).then((user) => {
            console.log("User:", user);
            // Store the user data with JWT token
            chrome.storage.local.set({ user: user }, function() {
              console.log('User data saved');
            });
            statusElement.innerText = `Signed in as ${user.email}`;
          });
        }
      });
    }
  });

  if (!rememberSiteBtn) {
    statusElement.textContent = 'Error: Button not found!';
    return;
  }

  rememberSiteBtn.addEventListener('click', async () => {
    statusElement.textContent = 'Scraping in progress...';
    
    try {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      
      if (!tab) {
        throw new Error("No active tab found");
      }

      const results = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: () => {
          function generalScraper() {
            return {
              title: document.title,
              url: window.location.href,
              raw_text: document.body.innerText,
              type: "general",
            };
          }

          function youtubeScraper() {
            return {
              title: document.title,
              url: window.location.href,
              raw_text: document.body.innerText.split("Related"),
              type: "youtube",
            };
          }

          function scrapePageContent() {
            const url = window.location.href;
            return url.includes('youtube.com/watch') ? youtubeScraper() : generalScraper();
          }

          return scrapePageContent();
        }
      });

      const scrapedContent = results[0].result;
      console.log("Scraped content:", scrapedContent);

      // Get the JWT token from storage
      const token = await chrome.storage.local.get(['user']);
      if (!token.user.token) {
        throw new Error("Not authenticated - please sign in first");
      }

      // Send scraped content to backend
      const response = await fetch('http://localhost:8080/api/remember', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.user.token}`
        },
        body: JSON.stringify({
          url: scrapedContent.url,
          title: scrapedContent.title,
          raw_text: scrapedContent.raw_text,
          type: scrapedContent.type,
        })
      });

      if (!response.ok) {
        throw new Error(`Server error: ${response.status}`);
      }
      
      statusElement.textContent = 'Site content scraped successfully!';
    } catch (error) {
      console.error("Error during scraping:", error);
      statusElement.textContent = `Error: ${error.message}`;
    }
  });
});