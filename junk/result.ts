import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import { sendData, sendError } from '../core/api-helper.ts'
import { ResultService } from '../services/result.ts'
import Ajv from '../vendor/ajv.js'

const RES_PLURAL = 'results'
const RES = 'result'

export class ResultRoutes {
  private service: ResultService
  private validate

  constructor(app: Application) {
    const router = new Router()
    this.service = new ResultService()
    this.post = this.post.bind(this)

    router.post(`/api/${RES_PLURAL}`, this.post)
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
      const result = await ctx.request.body().value
      const valid = this.validate(result)

      if (!valid) {
        throw this.validate.errors
      }

      const id = this.service.create(result)
      sendData(ctx, { id })
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }
}
