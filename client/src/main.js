import store from './store.js'
import Main from './components/main.html'
import Header from './components/header.html'
import Aside from './components/aside.html'
import './styles/style.scss'

document.addEventListener('DOMContentLoaded', main)

function main() {
  window.store = store
  new Header({ target: document.querySelector('header'), store })
  new Main({ target: document.querySelector('main'), store })
  new Aside({ target: document.querySelector('aside'), store })
}
