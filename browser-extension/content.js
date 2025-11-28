(function () {
  let currentUrl = location.href;
  let btnInserted = false;
  let btn, spinner, textNode;

  // Observe DOM for video container
  const observer = new MutationObserver(checkAndInsertButton);
  observer.observe(document.body, { childList: true, subtree: true });

  // Watch for SPA URL changes
  const urlObserver = new MutationObserver(() => {
    if (location.href !== currentUrl) {
      currentUrl = location.href;
      resetButtonForNewVideo();
    }
  });
  urlObserver.observe(document.body, { childList: true, subtree: true });

  function checkAndInsertButton() {
    if (btnInserted) return;
    const target = document.querySelector("#below");
    if (!target) return;

    // Create button
    btn = document.createElement("button");
    btn.id = "ytUrlPosterBtn";
    btn.style = `
      margin: 8px;
      padding: 6px 12px;
      background: #ff0000;
      color: white;
      border: none;
      border-radius: 4px;
      cursor: pointer;
      font-weight: bold;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
    `;

    // Text node for status
    textNode = document.createTextNode("Send to API");
    btn.appendChild(textNode);

    // Spinner
    spinner = document.createElement("span");
    spinner.style.cssText = `
      width: 16px;
      height: 16px;
      border: 2px solid rgba(255,255,255,0.3);
      border-top-color: #fff;
      border-radius: 50%;
      animation: spin 1s linear infinite;
      display: none;
    `;
    btn.appendChild(spinner);

    // Add keyframes for spinner
    const style = document.createElement("style");
    style.textContent = `
      @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
    `;
    document.head.appendChild(style);

    target.prepend(btn);
    btnInserted = true;

    initializeButtonState();
    attachButtonListeners();
    observer.disconnect();
  }

  function initializeButtonState() {
    chrome.storage.sync.get(["sentUrls"], data => {
      const isAlreadySent = data.sentUrls?.includes(currentUrl) || false;
      btn.dataset.isAlreadySent = isAlreadySent ? "true" : "false";
      btn.dataset.isSending = "false";
      textNode.nodeValue = isAlreadySent ? "✅ Already sent" : "Send to API";
    });
  }

  function attachButtonListeners() {
    btn.onclick = () => triggerSend(currentUrl);

    chrome.runtime.onMessage.addListener(msg => {
      if (msg.action === "updateButton" && msg.url === currentUrl) {
        showTemporaryFeedback(msg.message);
      }
    });
  }

  function triggerSend(url) {
    if (btn.dataset.isSending === "true") return;

    btn.dataset.isSending = "true";
    showSpinner(true);
    chrome.runtime.sendMessage({ action: "submitUrl", url });
  }

  function showSpinner(show) {
    spinner.style.display = show ? "inline-block" : "none";
    textNode.nodeValue = show ? "Sending..." :
      (btn.dataset.isAlreadySent === "true" ? "✅ Already sent" : "Send to API");
  }

  function showTemporaryFeedback(message) {
    showSpinner(false);
    textNode.nodeValue = message;

    setTimeout(() => {
      chrome.storage.sync.get(["sentUrls"], data => {
        const isAlreadySent = data.sentUrls?.includes(currentUrl) || false;
        btn.dataset.isAlreadySent = isAlreadySent ? "true" : "false";
        textNode.nodeValue = isAlreadySent ? "✅ Already sent" : "Send to API";
        btn.dataset.isSending = "false";
      });
    }, 3000);
  }

  function resetButtonForNewVideo() {
    btnInserted = false;
    if (btn) {
      btn.remove();
      btn = null;
      spinner = null;
      textNode = null;
    }
    checkAndInsertButton();
  }
})();
