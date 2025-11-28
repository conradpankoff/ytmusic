const input = document.getElementById("apiEndpoint");
const saveBtn = document.getElementById("saveBtn");

// Load saved endpoint
chrome.storage.sync.get(["apiEndpoint"], data => {
  if (data.apiEndpoint) input.value = data.apiEndpoint;
});

saveBtn.onclick = () => {
  chrome.storage.sync.set({ apiEndpoint: input.value }, () => {
    alert("API endpoint saved!");
  });
};
