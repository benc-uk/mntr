import { DB } from 'https://deno.land/x/sqlite@v2.3.2/mod.ts'

const DATA_PATH = 'data'

export let db: DB

export function open(dbFile: string) {
  try {
    Deno.mkdirSync(DATA_PATH)
  } catch (err) {
    // Do nothing!
  }

  // Open main database
  db = new DB(`${DATA_PATH}/${dbFile}`)
}
