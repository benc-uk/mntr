import { Context } from 'https://deno.land/x/oak@v6.3.2/mod.ts'

// deno-lint-ignore no-explicit-any
export function sendData(ctx: Context, payload: any) {
  ctx.response.headers.set('Content-Type', 'application/json')
  ctx.response.status = 200
  ctx.response.body = JSON.stringify(payload)
}

export function sendError(ctx: Context, error: Error, code = 500) {
  let errorMessage
  // Sqlite errors are weird unless you call toString() you get no details
  if (!error.message) {
    errorMessage = error
  } else {
    errorMessage = error.message
  }

  ctx.response.headers.set('Content-Type', 'application/json')
  ctx.response.status = code
  ctx.response.body = JSON.stringify({ error: errorMessage })
}
