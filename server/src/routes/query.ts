import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'

import { sendData, sendError } from '../core/api-helper.ts'

export class MonitorRoutes {
  constructor(app: Application) {
    const router = new Router()
    this.service = new MonitorService()
  }

  async post(ctx: RouterContext) {
    try {
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }
}
