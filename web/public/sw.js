self.addEventListener("install", (event) => {
  self.skipWaiting();
  event.waitUntil(caches.open("casca-v1").then((cache) => cache.addAll(["/"])));
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("fetch", () => {
  // Online-first: a rede é a fonte da verdade; o SW só habilita instalação.
});
