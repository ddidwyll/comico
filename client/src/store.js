import { Store } from 'svelte/store.js'

const api = location.origin + '/api/'
const pub = location.origin + '/pub/'
const now = () => Date.now().toString().slice(0, 10)


/*======== store section ========*/

/** global store methods, not all **/
class ComicoStore extends Store {
  take() { this.set({ busy: true }) }
  release() { this.set({ busy: false }) }
  toggle(prop, state = {}) {
    state[prop] = !this.get()[prop]
    this.set(state)
  }
  message(text) {
    clearTimeout(this.get().timer)
    this.set({
      message: text,
      timer: setTimeout(() => this.set({ message: '' }), 5000)
    })
  }
  getOne(type, id) {
    return this.get()['_' + type].find(item => item.id === id) || {}
  }
  repairImage(img) {
    const empty = this.get().empty
    if (img.src !== empty) img.src = empty
  }
  showImage(target) {
    target.classList.remove('loading')
    this.release()
  }
  formatDate(unix) {
    const date = new Date(unix * 1000), mins = date.getMinutes()
    return `${date.getDate()}/${date.getMonth() + 1} ` +
      `${date.getHours()}:${mins < 10 ? '0' + mins : mins}`
  }
  reply(name) {
    this.set({ _comment: `@${name}, ` })
    document.querySelector('textarea').focus()
  }
  changeType(type) {
    if (this.get().hashType === type) return
    this.goto({ type, id: null, search: null, page: null })
    this.closeModal()
  }
  goBack() {
    const { isAlone, isModal, _lastHash } = this.get()
    if (isModal) this.closeModal()
    else if (isAlone && !!_lastHash) location.hash = _lastHash
    else this.search('')
  }
  nextPage() {
    const { hashPage, maxPages } = this.get()
    if (!!maxPages && hashPage < maxPages)
      this.goto({ page: hashPage + 1 })
  }
  prevPage() {
    const { hashPage } = this.get()
    if (hashPage > 0) this.goto({ page: hashPage - 1 || null})
  }
  nextItem() {
    const { curItem, isForm } = this.get()
    if (!!curItem && !!curItem.next) {
      this.goto({ id: curItem.next })
      if (isForm) this.closeModal()
    }
    if (!curItem) {
      const focus = document.querySelector('article:focus')
      const first = document.querySelector('article:first-child')
      if (!!focus && !!focus.nextElementSibling) focus.nextElementSibling.focus()
      else if (!!first) first.focus()
    }
  }
  prevItem() {
    const { curItem, isForm } = this.get()
    if (!!curItem && !!curItem.prev) {
      this.goto({ id: curItem.prev })
      if (isForm) this.closeModal()
    }
    if (!curItem) {
      const focus = document.querySelector('article:focus')
      const last = document.querySelector('article:last-child')
      if (!!focus && !!focus.previousElementSibling) focus.previousElementSibling.focus()
      else if (!!last) last.focus()
    }
  }
  gotoCmnt(type, id, cmntIndex) {
    const page = cmntIndex / 12 ^ 0, index = cmntIndex - (page * 12)
    this.goto({ type, id, page, search: null }); this.closeModal()
    setTimeout(() => {
      const cmnt = document.querySelectorAll('.comments article')[index]
      if (!!cmnt) scrollTo(0, cmnt.getBoundingClientRect().top - 70)
    }, 500)
  }
  gotoCmntsPage(page) {
    this.goto({ page })
    const pagerTop = document.querySelector('.comments > figure')
      .getBoundingClientRect().top
    if (pagerTop < 0) scrollBy(0, pagerTop)
  }
  search(value) {
    this.goto({ search: value || null })
  }
  setForm(value, prop, state = {}) {
    const { curItem, hashType, form } = this.get()
    value = typeof value === 'string' ? value.trim() : value
    form[prop] = value; state.form = Object.assign({}, form)
    state[`_${hashType}` + (!curItem ? 'Add' : 'Edit')] = form
    this.set(state)
  }
  setFormArr(prop, value, delim = ',', length = 25) {
    value = (value || '').trim()
    value = value.endsWith(delim) ? value.slice(0, -1) : value
    const result = value.split(delim).slice(0, 10)
      .filter(t => !!t).map(item => item.trim().slice(0, length))
    this.setForm(result, prop)
  }
} const proto = ComicoStore.prototype
/** store state properties, not all **/
const store = new ComicoStore({
  now: now(), online: true, busy: true, hash: {}, _goods: [], _posts: [], form: {},
  _users: [], _cmnts: [], _files: [],  _mtimes: {}, _comment: '', other: [],
  imgIndex: 0, _images: { goods: {}, users: {} }, empty: `data:image/svg+xml;utf8,
  <svg version="1.1" viewBox="0 0 50 50" xml:space="preserve" xmlns="http://www.w3.org/2000/svg">
  <path d="M1,43h48V7H1V43z M3,41v-7.586l11-11l10,10l17-17l6,6V41H3z M47,9v9.586l-6-6l-17,17l-10-10l-11,11V9H47z"/>
  <path d="m24 22c2.757 0 5-2.243 5-5s-2.243-5-5-5-5 2.243-5 5 2.243 5 5 5zm0-8c1.654 0 3 1.346 3 3s-1.346 3-3 3-3-1.346-3-3 1.346-3 3-3z"/>
  </svg>`
})


/*======== fetch api section ========*/

store.compute('headers', ['_token'], (token) => ({
  'Content-Type': 'application/json',
  'Authorization': 'Bearer ' + token
}))
store.compute('method', ['curItem'], (item) => !item ? 'POST' : 'PUT')
/** send GET request to public server api, return data **/
proto.GET = async function(type) {
  this.take()
  const response = await fetch(pub + type).catch(() => ({ ok: false }))
  if (type === 'mtimes') this.set({ online: response.ok })
  this.release()
  return response.ok ? await response.json().catch(() => []) : []
}
/** set subscribe or ignore tag for current user **/
proto.tag = async function(action, tag) {
  const { isSigned, headers, busy } = this.get()
  if (!isSigned || busy) return
  this.take()
  const response = await fetch(api + action + '/' + tag, { headers })
  if (response.ok) await this.checkUpdate('users')
  setTimeout(() => this.release(), 1000)
  this.release()
}
/** compare local and server last modification times and fetch new data if its not equal **/
proto.checkUpdate = async function(type, mtimes, updates = {}) {
  const local = this.get()._mtimes, net = mtimes || await this.GET('mtimes')
  if (!net[type] || local[type] === net[type]) return
  const value = updates['_' + type] = await this.GET(type)
  local[type] = net[type]; updates._mtimes = local
  if (!!value && !!value.sort) this.set(updates)
}
/** check updates for all data types, every ~0.5 min and just now **/
async function checkUpdates() {
  setTimeout(() => checkUpdates(), 30000)
  const mtimes = await store.GET('mtimes');
  ['users', 'goods', 'posts', 'cmnts', 'files']
    .forEach(type => store.checkUpdate(type, mtimes))
}
/** PUT or POST form data, JWT for authorization  **/
proto.PUST = async function({ form = {}, formErrs }, type, state = {}) {
  const { isSigned, busy, curItem, hashType, _id, headers, method } = this.get()
  if (!!formErrs || !isSigned || busy) return
  this.take()
  type = type || hashType
  if (!curItem && type !== 'users') form = Object.assign(form, { id: now(), auth: _id })
  const body = JSON.stringify(form)
  const response = await fetch(api + type, { method, headers, body })
    .catch(() => null)
  const result = !response ? 'Connection error' : await response.json()
  if (!!response && response.ok && type !== 'pass') {
    await this.checkUpdate(type)
    state[`_${hashType}` + (!curItem ? 'Add' : 'Edit')] = {}
    this.set(state); this.closeModal()
    this.goto({ type: type, id: form.id, search: null, page: null })
  } this.message(result.message || result); this.release()
}
/** delete item **/
proto.DELETE = async function(type, id) {
  const { headers, isSigned, hashType, busy } = this.get(), method = 'DELETE'
  if (!isSigned || busy || !confirm('Are you sure?')) return
  this.take()
  const response = await fetch(api + type + '/' + id, { headers, method })
    .catch(() => null)
  const result = !response ? 'Connection error' : await response.json()
  if (!!response && response.ok) {
    await this.checkUpdate(type); this.closeModal()
    if (type !== 'cmnts') setTimeout(() => this.checkUpdate('cmnts'), 1500)
    if (type === hashType) this.goto({ id: null, search: null, page: null })
  } this.message(result.message || result); this.release()
}
/** upload image **/
proto.uploadImage = async function(event) {
  const { curItem, hashType, isSigned, busy} = this.get(), body = new FormData()
  if (!curItem || !isSigned || busy) return
  this.take()
  const headers = { Authorization: this.get().headers.Authorization }, method = 'POST'
  body.set('file', event.target.files[0]); body.set('name', curItem.id)
  const response = await fetch(api + 'upload/' + hashType, { method, headers, body })
    .catch(() => null)
  const result = !response ? 'Connection error' : await response.json()
  if (!!response && response.ok) await this.checkUpdate('files')
  this.message(result.message || result); this.release()
}


/*======== simple router section ========*/

/** change window location hash, last value if empty, skip if null **/
proto.goto = function({ type, id, page, search, event = { keyCode: 13 } }) {
  if (event.keyCode !== 13 && event.keyCode !== 32) return
  const { hash } = this.get(), args = arguments[0], query = [];
  ['id', 'page', 'search'].forEach(key => {
    const value = args[key] !== undefined ? args[key] : hash[key] || null
    if (value !== null) query.push(key + '=' + value)
  })
  location.hash = (type || hash.type || 'goods') +
    (!!query.length ? '?' + query.join('&') : '')
}
/** window location parse to state, store previous hash without 'id' **/
function parseHash(e = { oldURL: '#id=' }, result = {}) {
  const hash = location.hash.slice(1).split('?')
  if (hash[1]) hash[1].split('&').forEach(str => {
    str = str.split('='); if (!!str[0] && !!str[1])
      result[decodeURI(str[0]).toLowerCase()] = decodeURI(str[1]).toLowerCase()
  }); result.type = hash[0] || 'goods'
  const lastHash = !e.oldURL.split('#')[1] || !!~e.oldURL.split('#')[1].indexOf('id=') ?
    store.get()._lastHash : e.oldURL.split('#')[1]
  store.set({ hash: result, _lastHash: lastHash || '' })
  if (!result.id || !~e.oldURL.indexOf('id=')) scrollTo(0, 0)
} window.addEventListener('hashchange', parseHash)
/** current location computed statuses **/
store.compute('curHash', ['hash'], () => location.hash || '#goods')
store.compute('hashType', ['hash'], (hash) => hash.type || 'goods')
store.compute('hashId', ['hash'], (hash) => hash.id || null)
store.compute('hashPage', ['hash'], (hash) => +hash.page || 0)
store.compute('isAlone', ['hashId'], (id) => !!id)
store.compute('isPost', ['hashType'], (type) => type === 'posts')
store.compute('isGood', ['hashType'], (type) => type === 'goods')
store.compute('isUser', ['hashType'], (type) => type === 'users')
store.compute('hashSearch', ['hash'], (hash, result = {}) => {
  const search = (hash.search || '').trim()
  const start = search.indexOf('{'), end = search.lastIndexOf('}')
  if (!search || !~start) return { search }
  if (!~end) return { search: search.slice(0, start) }
  search.slice(start + 1, end).split(',').forEach(str => {
    const [ key, value ] = str.split(':')
    if (!!key && !!value) result[key.trim()] = value.trim()
  })
  result.search = search.slice(0, start) + search.slice(end + 1, search.length)
  return result
})
/** compute current document title **/
store.compute('title', ['hash', 'curItem', 'pagedItems'], (hash, item, items) => {
  const titles = { goods: 'Goods list', posts: 'Posts list', users: 'Users list' }
  const title = hash.id ? item ? item.title || item.id : '' :
    !!hash.search && !items.length ? '' : titles[hash.type]
  return title || 'Not Found'
})


/*======== modal window section ========*/

/** is modal showing? toggle body modal **/
store.compute('isModal', ['_modal'], (modal) => !!modal)
/** compute current modal type **/
store.compute('isSignIn', ['_modal'], (modal) => modal === 'signin')
store.compute('isSignUp', ['_modal'], (modal) => modal === 'signup')
store.compute('isForm', ['_modal'], (modal) => modal === 'form')
store.compute('isImage', ['_modal'], (modal) => modal === 'image')
store.compute('isActivity', ['_modal'], (modal) => modal === 'activity')
/** open modal 'type' **/
proto.openModal = (type) => store.set({ _modal: type })
/** click to close modal **/
proto.closeModal = (event = { target: { nodeName: 'ASIDE' } }) =>
  event.target.nodeName === 'ASIDE' && store.set({ _modal: '' })


/*======== user helpers section ========*/

store.compute('me', ['_id', '_users'], (id, users) =>
  users.find(user => user.id === id) || null)
store.compute('tagged', ['me'], (me) => ({
  ignores: me ? me.ignores || [] : [], scribes: me ? me.scribes || [] : [] }))
store.compute('isSigned', ['_expire', 'now', 'me', 'online'],
  (expire, now, user, online) => expire > now && !!user && !!online)
store.compute('isMy', ['isSigned', 'curItem', 'me'], (isSigned, item, me) =>
  !!isSigned && !!item && (item.auth === me.id || item.id === me.id || !!me.status))
proto.logout = () => store.set({ _id: '', _expire: '', _token: '' })
setInterval(() => store.set({ now: now() }), 3600000)


/*======== filter and pagination section ========*/

/** compute current items by type **/
store.compute('items', ['_goods', '_posts', '_users'],
  (goods, posts, users) => ({ goods, posts, users }))
/** compute users map id=>name **/
store.compute('users', ['_users'], (users, result = {}) => {
  users.forEach(user => { result[user.id] = user.title || user.id })
  return result
})
/** compute comments and replies, group by owner type and id **/
store.compute('comments', ['_cmnts', 'me', 'tagged'], (items, me, tags) => {
  const activity = { goods: {}, posts: {}, users: {},
    count: 0,usersCount: 0, goodsCount: 0, postsCount: 0
  }, result = { goods: {}, posts: {}, users: {}, activity }
  if (!!me && !!tags.ignores.length) items = items.filter(item =>
    !~tags.ignores.indexOf(item.auth.toLowerCase()))
  items.sort((a, b) => a.id - b.id).forEach(cmnt => {
    const cmntType = result[cmnt.type]
    if (!cmntType[cmnt.owner]) cmntType[cmnt.owner] = []
    cmnt.index = cmntType[cmnt.owner].length
    cmntType[cmnt.owner].push(cmnt)
    if (!!me && cmnt.auth !== me.id && (cmnt.to === me.id ||
      (cmnt.to === 'support' && !!me.status) || cmnt.to === 'all')) {
      const cmntType = result.activity[cmnt.type]
      if (!cmntType[cmnt.owner]) cmntType[cmnt.owner] = []
      cmntType[cmnt.owner].push(cmnt)
      if (+cmnt.id > me.activity * 1000) {
        result.activity.count++
        result.activity[`${cmnt.type}Count`]++
      }
    }
  }); return result
})
/** compute items ignored by user **/
store.compute('ignoredItems', ['items', 'tagged'], (items, tags) => {
  if (!!tags.ignores.length) ['goods', 'posts'].forEach(type =>
    items[type] = items[type].filter(item =>
      !~tags.ignores.indexOf(item.type.toLowerCase()) &&
      !~tags.ignores.indexOf(item.auth.toLowerCase())))
  return items
})
/** compute items subscribed by user **/
store.compute('scribedItems', ['items', 'tagged'], (items, tags) => {
  if (!tags.scribes.length) return { goods: [], posts: [], users: [] }
  const scribed = Object.assign({}, items);
  ['goods', 'posts'].forEach(type =>
    scribed[type] = scribed[type].filter(item =>
      !!~tags.scribes.indexOf(item.type.toLowerCase()) ||
      !!~tags.scribes.indexOf(item.auth.toLowerCase())))
  return scribed
})
/** compute current items by location **/
store.compute('curItems', ['ignoredItems', 'scribedItems', '_isScribes', 'hashType'],
  (ignored, scribed, iS, type) => (iS ? scribed[type] : ignored[type]) || [])
/** compute current, next and previous items, if 'id' in hash **/
store.compute('curItem', ['hashId', 'curItems'], (id, items) => {
  if (!id || !items || !items.length) return null
  const index = items.findIndex(item => item.id.toLowerCase() === id)
  if (!~index) return null
  const item = items[index]
  item.prev = items[index - 1] ? items[index - 1].id : ''
  item.next = items[index + 1] ? items[index + 1].id : ''
  return item
})
/** compute current items by search filter **/
store.compute('filteredItems', ['hashSearch', 'curItems', 'hashId', 'hashType', 'comments'],
  (search, items, id, type, cmnts) => {
    if (!!id) items = cmnts[type] ? cmnts[type][id] || [] : []
    const str = (search.search || '').trim(), auth = search.auth || ''
    delete search.search; delete search.auth
    const standart = ['id', 'price', 'type', 'text', 'title'], keys = Object.keys(search)
    return !str && !keys.length && !auth ? items : items.filter(item =>
      standart.some(prop => !!item[prop] && !!~item[prop].toLowerCase().indexOf(str)) &&
        (!keys.length || !!item.Table && keys.every(prop =>
          !!item.Table[prop] && !!~item.Table[prop].toLowerCase().indexOf(search[prop]))) &&
            (!auth || !item.auth || auth === item.auth.toLowerCase()))
  })
/** compute max pages count, 12 items per page **/
store.compute('maxPages', ['hashPage', 'filteredItems'], (page, items) => {
  const max = (items.length - 1) / 12 ^ 0
  if (page > max && page !== 999)
    setTimeout(store.goto({ page: max || null }), 500)
  return max
})
/** compute current paginated items **/
store.compute('pagedItems', ['hashPage', 'filteredItems'], (page, items) => {
  if (page === 999) return items
  return items.slice(12 * page, 12 * page + 12)
})


/*======== state and files cache section ========*/

/** sync state with localStorage, only with '_' prefix **/
store.on('state', ({ changed, current }) => {
  if (changed._files) cacheImages(current._files)
  Object.keys(changed).forEach(prop => {
    if (!prop.indexOf('_'))
      localStorage.setItem(prop, JSON.stringify(current[prop]))
  })
})
/** load state from localStorage **/
function loadState(state = {}) {
  for (let i = 0; i < localStorage.length; i++) {
    const prop = localStorage.key(i)
    const value = JSON.parse(localStorage.getItem(prop) || 'null')
    if (!!value && !prop.indexOf('_')) state[prop] = value
  } store.set(state)
}
/** cache new files, delete old files **/
async function cacheImages(newFiles = []) {
  const oldFiles = JSON.parse(localStorage.getItem('_files') || '[]')
  let iN = newFiles.length - 1, iO = oldFiles.length - 1
  if (!~iN) return
  store.take()
  const cache = caches ? await caches.open('comico') : null
  const images = { goods: {}, users: {} }
  if (!!cache && !!~iO) do {
    const file = oldFiles[iO], [ id, type ] = file.split(':')
    if (!~newFiles.indexOf(file))
      await cache.delete(`/img/${type}_${id}_sm.jpg`) &&
      await cache.delete(`/img/${type}_${id}.jpg`); iO--
  } while (~iO); do {
    const file = newFiles[iN], [ id, type, hash ] = file.split(':')
    const src = `/img/${type}_${id}_sm.jpg`
    images[type][id] = `${src}#${hash}`
    if (!!cache && !~oldFiles.indexOf(file))
      await cache.add(new Request(src, { cache: 'no-cache' })); iN--
  } while (~iN)
  store.set({ _images: images })
  store.release()
}


/*======== end section ========*/

if (!!navigator && !!navigator.serviceWorker) {
  navigator.serviceWorker.register('./worker.js')
}

loadState()
parseHash()
checkUpdates()

export default store
