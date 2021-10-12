import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// CSS assets
import 'bulma/css/bulma.css'
import './assets/css/global.css'

createApp(App)
  .use(router)
  .mount('#app')
