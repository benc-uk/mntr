import 'https://deno.land/x/dotenv@v1.0.1/load.ts'
import { Application, Router } from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import * as log from 'https://deno.land/std@0.78.0/log/mod.ts'

import { StatusRoutes } from './routes/status.ts'
import { MonitorRoutes } from './routes/monitors.ts'
import * as database from './core/database.ts'

log.info(`🚀 Mntr server is starting`)

// Handle config defaults
const MNTR_SERVER_PORT = parseInt(Deno.env.get('MNTR_SERVER_PORT') || '8000')
const MNTR_DATABASE = 'mntr.db'
log.info(`🧩 Port: ${MNTR_SERVER_PORT} DB: ${MNTR_DATABASE}`)

// Shared database instance
database.open(MNTR_DATABASE)

// Create app and load all routes & controllers
const app = new Application()
new StatusRoutes(app)
new MonitorRoutes(app)

// Redirect root and /api to status API
const defaultRoutes = new Router()
defaultRoutes.get(`/`, (ctx) => {
  ctx.response.redirect('/api/status')
})
defaultRoutes.get(`/`, (ctx) => {
  ctx.response.redirect('/api/status')
})
app.use(defaultRoutes.routes())

// Boom
await app.listen({ port: MNTR_SERVER_PORT, hostname: '0.0.0.0' })
