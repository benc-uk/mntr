import {
  Application,
  Router,
  RouterContext
} from 'https://deno.land/x/oak@v6.3.2/mod.ts'
import { sendData, sendError } from '../core/api-helper.ts'
import { YamlLoader } from 'https://deno.land/x/yaml_loader/mod.ts'

const RES_PLURAL = 'plugins'
const RES = 'plugin'

export class PluginRoutes {
  constructor(app: Application) {
    const router = new Router()
    this.get = this.get.bind(this)
    this.getAll = this.getAll.bind(this)

    router.get(`/api/${RES_PLURAL}`, this.getAll)
    router.get(`/api/${RES_PLURAL}/:name`, this.get)
    app.use(router.routes())
  }

  getAll(ctx: RouterContext) {
    const results = new Array<string>()
    for (const dirEntry of Deno.readDirSync('./plugins/')) {
      const fn = dirEntry.name
      if (fn.endsWith('.yaml')) {
        results.push(fn.substring(0, fn.length - 5))
      }
    }
    sendData(ctx, results)
  }

  async get(ctx: RouterContext) {
    const name = ctx.params.name || ''
    const yamlLoader = new YamlLoader()

    try {
      const plugin = await yamlLoader.parseFile(`./plugins/${name}.yaml`)
      sendData(ctx, plugin)
    } catch (err) {
      sendError(ctx, err)
      return
    }
  }
}
