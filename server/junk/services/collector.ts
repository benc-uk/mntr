import { Collector } from '../core/models.ts'
import { db } from '../core/database.ts'
import { Empty } from 'https://deno.land/x/sqlite@v2.3.2/src/rows.ts'

export class CollectorService {
  constructor() {
    db.query(`CREATE TABLE IF NOT EXISTS collectors 
      (
        hostname TEXT PRIMARY KEY,
        version,
        lastSeen INTEGER 
      )`)
  }

  public createorUpdate(collector: Collector): boolean {
    // We overwrite the lastSeen value even if it's provided
    const lastSeen = Date.now()
    // Note we don't care if it already exists, effectively an upsert
    db.query(`INSERT OR REPLACE INTO collectors VALUES (?, ?, ?)`, [
      collector.hostname,
      collector.version,
      lastSeen
    ])
    return db.changes > 0
  }

  public read(hostname: string): Collector | null {
    const res = db.query(`SELECT * FROM collectors WHERE hostname = ?`, [
      hostname
    ])

    if (res === Empty) {
      return null
    }

    return res.asObjects().next().value
  }

  // deno-lint-ignore no-explicit-any
  public list(): Record<string, any>[] {
    const res = db.query(`SELECT * FROM collectors`)

    return [...res.asObjects()]
  }

  public remove(hostname: string): boolean {
    db.query(`DELETE FROM collectors WHERE hostname = ?`, [hostname])

    return db.changes > 0
  }
}
