importScripts("auth.js"); // API_URL, signInWithGoogle

// Authentication function that can be reused
function authenticateUser(callback) {
  signInWithGoogle()
    .then((user) => {
      if (callback) callback(user);
    })
    .catch((error) => {
      console.error("Auth Error:", error);
      const msg = error.message || "";
      let errorMessage = "Authentication failed. Please try again.";
      if (msg.includes("not approve") || msg.includes("cancel")) {
        errorMessage = "Sign-in cancelled. Please try again and allow access to continue.";
      } else if (msg.includes("network")) {
        errorMessage = "Network error. Please check your connection and try again.";
      } else if (msg.includes("Server sign-in failed")) {
        errorMessage = "Failed to sign in with server. Please try again.";
      }
      chrome.notifications.create({
        type: "basic",
        iconUrl: "icon.png",
        title: "Foundry - Authentication Error",
        message: errorMessage,
      });
    });
}

// Reusable function to scrape page content
function scrapePageContent(tabId, callback) {
  chrome.scripting.executeScript({
    target: { tabId: tabId },
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

      function scrapePageContentInternal() {
        const url = window.location.href;
        return url.includes('youtube.com/watch') ? youtubeScraper() : generalScraper();
      }

      return scrapePageContentInternal();
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
    callback(scrapedContent);
  });
}

// Reusable function to save content to backend
function savePageContent(scrapedContent, userToken, userEmail, successMessage = 'Page saved successfully') {
  fetch(`${API_URL}/api/documents`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${userToken}`
    },
    body: JSON.stringify({
      ...scrapedContent,
      user_id: userEmail
    })
  }).then(response => {
    if (response.ok) {
      console.log("Page saved successfully");
      chrome.notifications.create({
        type: 'basic',
        iconUrl: 'icon.png',
        title: 'Foundry',
        message: successMessage
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
}

// Combined function to scrape and save page content
function processPageSave(tab, user, successMessage = 'Page saved successfully') {
  scrapePageContent(tab.id, (scrapedContent) => {
    savePageContent(scrapedContent, user.token, user.email, successMessage);
  });
}

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
        // Use the reusable authentication function
        authenticateUser((user) => {
          // After successful authentication, use the reusable process function
          processPageSave(tab, user, 'Successfully signed in and saved page!');
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

        // Token valid, use the reusable process function
        processPageSave(tab, result.user);
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