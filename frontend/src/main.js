import './style.css'
import './lib/theme.js' // applies the saved light/dark theme on startup
import App from './App.svelte'

const app = new App({ target: document.getElementById('app') })

export default app
