document.getElementById("sendBtn").addEventListener("click", async () => {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab.url.includes("youtube.com/watch")) {
    alert("Not a YouTube video page.");
    return;
  }

  chrome.runtime.sendMessage({ action: "submitUrl", url: tab.url });
});

document.getElementById("openOptions").addEventListener("click", () => {
  chrome.runtime.openOptionsPage();
});
