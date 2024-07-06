// Injecting the script when the website specified in the manifest is loaded
const script: HTMLScriptElement = document.createElement('script');
script.src = chrome.runtime.getURL('script.js');
(document.head || document.documentElement).appendChild(script);
script.onload = function () {
  script.remove();
};

fetch('https://www.dotabuff.com/heroes/arc-warden', {
  method: 'GET',
})
  .then((response) => {
    console.log('response', response);
    return response.text();
  })
  .then((data) => {
    console.log('data', data);
  });


// chrome.storage.local.get('enabled',  async function (result) {  
//   if (result == undefined) {
//     // Here you define the logic of your script for the website
//     console.log('Owl esports does not exist, creating it and enabling it');
//     const res = chrome.storage.local.set({ enabled: true })
//     const awaitRes = await res;
//     console.log('Owl esports enbling result', awaitRes);
//   } else if (result.enabled) {
//     console.log('Owl esports enabled, disabling it');
//     const res = await chrome.storage.local.set({ enabled: false })
//     console.log('Owl esports disabling result', res);
//   } else if (!result.enabled) {
//     console.log('Owl esports not enabled, enabling it');
//     const res = await chrome.storage.local.set({ enabled: true })
//     console.log('Owl esports enabling result', res);
//   }
// });