import 'https://deno.land/x/dotenv@v1.0.1/load.ts'
import { Application, Router } from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import * as log from 'https://deno.land/std@0.78.0/log/mod.ts'

import { CollectorRoutes } from './routes/collector.ts'
import * as database from './core/database.ts'
import { StatusRoutes } from './routes/status.ts'
import { MonitorRoutes } from './routes/monitors.ts'

log.info(`🚀 Mntr server is starting`)

// Handle config defaults
const MNTR_SERVER_PORT = parseInt(Deno.env.get('MNTR_SERVER_PORT') || '8000')
const MNTR_SERVER_DB = Deno.env.get('MNTR_SERVER_DB') || 'mntr.db'
log.info(`🧩 Port: ${MNTR_SERVER_PORT}, Database: ${MNTR_SERVER_DB}`)

// Open database
database.open(MNTR_SERVER_DB)
log.info('💠 Database opened')

// Create app and load all routes & controllers
const app = new Application()
new CollectorRoutes(app)
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
