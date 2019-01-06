const CACHE = 'comico'
const FILES = [
  '/',
  '/bundle.css',
  '/bundle.js'
]

self.addEventListener('install', (e) => {
  e.waitUntil(caches.open(CACHE).then(cache => cache.addAll(FILES)).then(() => self.skipWaiting()))
})

self.addEventListener('fetch', (e) => {
  if (e.request.method !== 'GET' || !!~e.request.url.indexOf('/pub/') || !!~e.request.url.indexOf('/api/')) return
  if (!!~e.request.url.lastIndexOf('_sm.jpg') && e.request.cache !== 'no-cache') return e.respondWith(fromCache(e.request))
  e.respondWith(toCache(e.request))
})

async function fromCache(request) {
  return await (await caches.open(CACHE)).match(request) || new Response(null, { status: 404 })
}

async function toCache(request) {
  const response = await fetch(request).catch(() => fromCache(request))
  if (!!response && response.ok) (await caches.open(CACHE)).put(request, response.clone())
  return response
}
