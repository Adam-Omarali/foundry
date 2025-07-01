let DEV_MODE = true;
let API_URL = DEV_MODE ? "http://localhost:8080" : "https://foundry-production-7176.up.railway.app";

// Create context menu on installation
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: "foundryScrape",
    title: "Save to Foundry",
    contexts: ["page"]
  });
});

// Handle context menu clicks
chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === "foundryScrape") {
    // Check if user is authenticated
    chrome.storage.local.get(['user'], function(result) {
      if (!result.user || !result.user.token) {
        console.log("User not authenticated");
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon.png',
          title: 'Foundry',
          message: 'Please sign in to save pages'
        });
        // Launch Google auth flow directly
        const clientId = "634531581191-h4u6sth79fvsobc80r6ube6rjv5t9j50.apps.googleusercontent.com";
        const redirectUri = chrome.identity.getRedirectURL();
        const scopes = ["https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"];
        const authUrl = `https://accounts.google.com/o/oauth2/auth?client_id=${clientId}` +
                      `&response_type=token` +
                      `&redirect_uri=${encodeURIComponent(redirectUri)}` +
                      `&scope=${encodeURIComponent(scopes.join(" "))}` +
                      `&prompt=consent`;

        chrome.identity.launchWebAuthFlow({
          url: authUrl,
          interactive: true
        }, function(redirectUrl) {
          if (chrome.runtime.lastError || !redirectUrl) {
            console.error("Auth Error:", chrome.runtime.lastError?.message);
            return;
          }

          const params = new URLSearchParams(new URL(redirectUrl).hash.substring(1));
          const accessToken = params.get("access_token");

          // Sign in with backend
          fetch(`${API_URL}/api/signin`, {
            method: "GET", 
            headers: {
              "Authorization": `Bearer ${accessToken}`,
            },
          })
          .then(res => res.json())
          .then(user => {
            chrome.storage.local.set({ user: user }, () => {
              // After storing user, proceed with saving the page
              fetch(`${API_URL}/api/remember`, {
                method: "POST",
                headers: {
                  "Authorization": `Bearer ${user.token}`,
                  "Content-Type": "application/json"
                },
                body: JSON.stringify({ url: tab.url })
              })
              .then(res => {
                if (res.ok) {
                  chrome.notifications.create({
                    type: 'basic',
                    iconUrl: 'icon.png',
                    title: 'Foundry',
                    message: 'Successfully signed in and saved page!'
                  });
                } else {
                  // Try to parse the error response
                  res.json().then(errorData => {
                    if (res.status === 403 && errorData.error === "Document limit reached") {
                      chrome.notifications.create({
                        type: 'basic',
                        iconUrl: 'icon.png',
                        title: 'Foundry',
                        message: 'You\'ve reached your free plan limit. Please upgrade to add more documents.'
                      });
                    } else {
                      chrome.notifications.create({
                        type: 'basic',
                        iconUrl: 'icon.png',
                        title: 'Foundry',
                        message: errorData.message || 'Signed in but failed to save page. Please try again.'
                      });
                    }
                  }).catch(() => {
                    chrome.notifications.create({
                      type: 'basic', 
                      iconUrl: 'icon.png',
                      title: 'Foundry',
                      message: 'Signed in but failed to save page. Please try again.'
                    });
                  });
                }
              });
            });
          });
        });
        return;
      }

      // Verify token is still valid
      fetch(`${API_URL}/api/verifytoken`, {
        method: "GET",
        headers: {
          "Authorization": `Bearer ${result.user.token}`,
        },
      }).then((res) => {
        if (!res.ok) {
          // Token invalid or expired
          chrome.storage.local.remove(['user'], function() {
            console.log("Session expired");
            chrome.notifications.create({
              type: 'basic',
              iconUrl: 'icon.png',
              title: 'Foundry',
              message: 'Please sign in again'
            });
          });
          return;
        }

        console.log("Token valid");

        // Token valid, proceed with scraping
        chrome.scripting.executeScript({
          target: { tabId: tab.id },
          func: () => {
            function generalScraper() {
              console.log("Scraping general page content");
              return {
                title: document.title,
                url: window.location.href,
                raw_text: document.body.innerText,
                type: "general",
              };
            }

            function youtubeScraper() {
              console.log("Scraping YouTube page content");
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
        }, (results) => {
          if (chrome.runtime.lastError) {
            console.log(chrome.runtime.lastError);
            console.error(chrome.runtime.lastError);
            chrome.notifications.create({
              type: 'basic',
              iconUrl: 'icon.png',
              title: 'Foundry',
              message: 'Failed to scrape page'
            });
            return;
          }
          
          const scrapedContent = results[0].result;
          console.log("Scraped content:", scrapedContent);

          // Send scraped content to backend with user's JWT token
          fetch(`${API_URL}/api/remember`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Authorization": `Bearer ${result.user.token}`
            },
            body: JSON.stringify({
              ...scrapedContent,
              user_id: result.user.email
            })
          }).then(response => {
            if (response.ok) {
              console.log("Page saved successfully");
              chrome.notifications.create({
                type: 'basic',
                iconUrl: 'icon.png',
                title: 'Foundry',
                message: 'Page saved successfully'
              });
            } else {
              // Try to parse the error response
              response.json().then(errorData => {
                console.log("Error data:", errorData);
                if (response.status === 403 && errorData.error === "Document limit reached") {
                  chrome.notifications.create({
                    type: 'basic',
                    iconUrl: 'icon.png',
                    title: 'Foundry',
                    message: 'You\'ve reached your free plan limit. Please upgrade to add more documents.'
                  });
                } else {
                  chrome.notifications.create({
                    type: 'basic',
                    iconUrl: 'icon.png',
                    title: 'Foundry',
                    message: errorData.message || 'Failed to save page'
                  });
                }
              }).catch(() => {
                // If we can't parse the error response, show a generic message
                chrome.notifications.create({
                  type: 'basic',
                  iconUrl: 'icon.png',
                  title: 'Foundry',
                  message: 'Failed to save page'
                });
              });
            }
          }).catch(error => {
            console.error('Error saving page:', error);
            chrome.notifications.create({
              type: 'basic',
              iconUrl: 'icon.png',
              title: 'Foundry',
              message: 'Error saving page'
            });
          });
        });
      }).catch(error => {
        console.error('Error verifying token:', error);
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon.png',
          title: 'Foundry',
          message: 'Error verifying authentication'
        });
      });
    });
  }
});