chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.action === "submitUrl" && msg.url) {
    chrome.storage.sync.get(["apiEndpoint", "sentUrls"], data => {
      const apiEndpoint = data.apiEndpoint;
      const sentUrls = data.sentUrls || [];

      if (!apiEndpoint) {
        notify(sender.tab?.id, "❌ API endpoint not configured.", msg.url);
        sendResponse({ success: false, error: "API endpoint not configured." });
        return;
      }

      fetch(apiEndpoint, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ urls_or_ids: msg.url })
      })
        .then(async res => {
          const text = await res.text();
          if (res.ok) {
            if (!sentUrls.includes(msg.url)) {
              sentUrls.push(msg.url);
              chrome.storage.sync.set({ sentUrls });
            }
            notify(sender.tab?.id, "✅ URL sent successfully!", msg.url);
            sendResponse({ success: true, response: text });
          } else {
            notify(sender.tab?.id, "❌ Failed: " + text, msg.url);
            sendResponse({ success: false, response: text });
          }
        })
        .catch(err => {
          notify(sender.tab?.id, "❌ Error: " + err.message, msg.url);
          sendResponse({ success: false, error: err.message });
        });
    });

    return true; // async response
  }
});

chrome.commands.onCommand.addListener(async (command) => {
  if (command !== "send_youtube_url") return;

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab || !tab.url.includes("youtube.com/watch")) return;

  chrome.runtime.sendMessage({ action: "submitUrl", url: tab.url });
});

function notify(tabId, message, url) {
  chrome.notifications.create({
    type: "basic",
    iconUrl: "icon.png",
    title: "YouTube URL Poster",
    message
  });

  if (tabId) {
    chrome.tabs.sendMessage(tabId, { action: "updateButton", message, url });
  }
}
