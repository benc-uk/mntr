import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'

import { sendData, sendError } from '../core/api-helper.ts'
import Ajv from '../vendor/ajv.js'
import { MonitorService } from '../services/monitor.ts'
import { Monitor } from '../core/models.ts'

const RES_PLURAL = 'monitors'
const RES = 'monitor'

export class MonitorRoutes {
  private service: MonitorService
  private validate

  constructor(app: Application) {
    const router = new Router()
    this.service = new MonitorService()

    this.dumpConfig = this.dumpConfig.bind(this)
    this.getAll = this.getAll.bind(this)
    this.post = this.post.bind(this)
    this.put = this.put.bind(this)
    this.delete = this.delete.bind(this)

    router.get(`/api/${RES_PLURAL}/config`, this.dumpConfig)
    router.get(`/api/${RES_PLURAL}`, this.getAll)
    router.post(`/api/${RES_PLURAL}`, this.post)
    router.put(`/api/${RES_PLURAL}/:plugin/:name`, this.put)
    router.delete(`/api/${RES_PLURAL}/:plugin/:name`, this.delete)
    app.use(router.routes())

    const schema = JSON.parse(Deno.readTextFileSync(`../schema/${RES}.json`))
    const ajv = new Ajv({ allErrors: true })
    this.validate = ajv.compile(schema)
  }

  async post(ctx: RouterContext) {
    try {
      const mon = await ctx.request.body().value
      const valid = this.validate(mon)

      if (!valid) {
        throw this.validate.errors
      }

      this.service.create(mon)
      sendData(ctx, mon)
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  delete(ctx: RouterContext) {
    try {
      const plugin = ctx.params.plugin || ''
      const name = ctx.params.name || ''
      const ok = this.service.remove(plugin, name)
      if (!ok) {
        sendError(ctx, new Error(`${RES} '${plugin}/${name}' not found`), 404)
        return
      }

      sendData(ctx, {
        msg: `${RES} '${plugin}/${name}' was deleted successfully`
      })
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  dumpConfig(ctx: RouterContext) {
    ctx.response.headers.append('content-type', 'application/x-yaml')
    const allYamls = new Array<string>()

    const monitors = this.service.list()

    // Weird formatting required due to YAML reasons
    for (const mon of monitors) {
      const yaml = `name: ${mon.name}
plugin: ${mon.plugin}
frequency: ${mon.frequency}
enabled: ${mon.enabled ? 'true' : 'false'}
runsOn: [${mon.runsOn.join(',')}]
params: 
  ${mon.params}`

      allYamls.push(yaml)
    }

    ctx.response.body = allYamls.join('\n---\n')
  }

  getAll(ctx: RouterContext) {
    try {
      const monitors = this.service.list()
      sendData(ctx, monitors)
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }

  async put(ctx: RouterContext) {
    try {
      const plugin = ctx.params.plugin || ''
      const name = ctx.params.name || ''
      const existingMon = this.service.read(plugin, name)
      if (!existingMon) {
        sendError(ctx, new Error(`${RES} '${plugin}/${name}' not found`), 404)
        return
      }

      const mon: Monitor = await ctx.request.body().value
      mon.name = name
      mon.plugin = plugin
      const valid = this.validate(mon)
      if (!valid) {
        throw this.validate.errors
      }

      this.service.update(mon)
      sendData(ctx, this.service.read(plugin, name))
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }
}
