import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import { sendData } from '../core/api-helper.ts'

const RES_PLURAL = 'status'
const RES = 'status'

export class StatusRoutes {
  constructor(app: Application) {
    const router = new Router()
    this.get = this.get.bind(this)

    router.get(`/api/${RES_PLURAL}`, this.get)
    app.use(router.routes())
  }

  get(ctx: RouterContext) {
    sendData(ctx, {
      alive: true
    })
  }
}
