import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import { sendData, sendError } from '../core/api-helper.ts'
import { CollectorService } from '../services/collector.ts'
import Ajv from '../vendor/ajv.js'

const RES_PLURAL = 'collectors'
const RES = 'collector'

export class CollectorRoutes {
  private service: CollectorService
  private validate

  constructor(app: Application) {
    const router = new Router()
    this.service = new CollectorService()
    this.get = this.get.bind(this)
    this.getAll = this.getAll.bind(this)
    this.post = this.post.bind(this)
    this.put = this.put.bind(this)
    this.delete = this.delete.bind(this)

    router.get(`/api/${RES_PLURAL}/:hostname`, this.get)
    router.get(`/api/${RES_PLURAL}`, this.getAll)
    router.post(`/api/${RES_PLURAL}`, this.post)
    router.put(`/api/${RES_PLURAL}/:hostname`, this.put)
    router.delete(`/api/${RES_PLURAL}/:hostname`, this.delete)
    app.use(router.routes())

    const schema = JSON.parse(Deno.readTextFileSync(`../schema/${RES}.json`))
    const ajv = new Ajv({ allErrors: true })
    this.validate = ajv.compile(schema)
  }

  get(ctx: RouterContext) {
    try {
      const hostname = ctx.params.hostname || ''
      const collector = this.service.read(hostname)
      if (!collector) {
        sendError(ctx, new Error(`${RES} '${hostname}' not found`), 404)
        return
      }
      sendData(ctx, collector)
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  async post(ctx: RouterContext) {
    try {
      const collector = await ctx.request.body().value
      const valid = this.validate(collector)

      if (!valid) {
        throw this.validate.errors
      }

      this.service.createorUpdate(collector)
      sendData(ctx, this.service.read(collector.hostname))
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  async put(ctx: RouterContext) {
    try {
      const hostname = ctx.params.hostname || ''
      const existingCollector = this.service.read(hostname)
      if (!existingCollector) {
        sendError(ctx, new Error(`${RES} '${hostname}' not found`), 404)
        return
      }
      const collector = await ctx.request.body().value
      // Mutate collector and overwrite hostname if they set it
      collector.hostname = hostname
      const valid = this.validate(collector)

      if (!valid) {
        throw this.validate.errors
      }

      this.service.createorUpdate(collector)
      sendData(ctx, this.service.read(hostname))
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  getAll(ctx: RouterContext) {
    try {
      const collectors = this.service.list()
      sendData(ctx, collectors)
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  delete(ctx: RouterContext) {
    try {
      const hostname = ctx.params.hostname || ''
      const ok = this.service.remove(hostname)
      if (!ok) {
        sendError(ctx, new Error(`${RES} '${hostname}' not found`), 404)
        return
      }

      sendData(ctx, { msg: `${RES} '${hostname}' was deleted successfully` })
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }
}
