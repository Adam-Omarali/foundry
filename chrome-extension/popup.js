/*
Foundry Extension popup. API_URL and signInWithGoogle come from auth.js,
loaded before this script in popup.html.
*/
document.getElementById("loginBtn").addEventListener("click", () => {
  const statusElement = document.getElementById("status");
  signInWithGoogle()
    .then((user) => {
      statusElement.innerText = `Signed in as ${user.email}`;
      document.getElementById("loginBtn").style.display = 'none';
      document.getElementById("rememberSiteBtn").disabled = false;
    })
    .catch((error) => {
      console.error("Auth Error:", error);
      statusElement.textContent = 'Sign-in failed. Please try again.';
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
      fetch(`${API_URL}/api/verifytoken`, {
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
      // No Foundry session stored: show the login button
      if (loginBtn) {
        loginBtn.style.display = 'block';
      }
      if (rememberSiteBtn) {
        rememberSiteBtn.disabled = true;
      }
      statusElement.textContent = 'Please sign in to use the extension';
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
              raw_text: document.body.innerText,
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
      const response = await fetch(`${API_URL}/api/documents`, {
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
        // Try to parse the error response
        try {
          const errorData = await response.json();
          if (response.status === 403 && errorData.error === "Document limit reached") {
            statusElement.textContent = 'You\'ve reached your free plan limit. Please upgrade to add more documents.';
          } else {
            statusElement.textContent = `Error: ${errorData.message || 'Failed to save site'}`;
          }
        } catch (parseError) {
          // If we can't parse the error response, show a generic message
          statusElement.textContent = `Error: Failed to save site (${response.status})`;
        }
        return;
      }
      
      statusElement.textContent = 'Site content scraped successfully!';
    } catch (error) {
      console.error("Error during scraping:", error);
      statusElement.textContent = `Error: ${error.message}`;
    }
  });
});