import { db } from '../core/database.ts'
import { Empty } from 'https://deno.land/x/sqlite@v2.3.2/src/rows.ts'
import { Monitor } from '../core/models.ts'

export class MonitorService {
  constructor() {
    db.query(`CREATE TABLE IF NOT EXISTS monitors 
    (
      name TEXT,
      plugin TEXT,
      enabled INTEGER,
      frequency INTEGER,
      runsOn TEXT,
      params TEXT,
      PRIMARY KEY (name, plugin)
    )`)
  }

  public list(): Monitor[] {
    const res = db.query(`SELECT * FROM monitors`)

    const monitors = new Array<Monitor>()
    for (const row of res.asObjects()) {
      monitors.push({
        name: row.name,
        plugin: row.plugin,
        enabled: row.enabled == 1,
        frequency: row.frequency,
        runsOn: row.runsOn.split(','),
        params: row.params
      })
    }

    return monitors
  }

  public create(mon: Monitor): boolean {
    db.query(
      `INSERT INTO monitors (name, plugin, enabled, frequency, runsOn, params) 
       VALUES (?, ?, ?, ?, ?, ?)`,
      [
        mon.name,
        mon.plugin,
        mon.enabled,
        mon.frequency,
        mon.runsOn.join(','),
        mon.params
      ]
    )

    return db.changes > 0
  }

  public remove(plugin: string, name: string): boolean {
    db.query(`DELETE FROM monitors WHERE plugin = ? AND name = ?`, [
      plugin,
      name
    ])

    return db.changes > 0
  }

  public read(plugin: string, name: string): Monitor | null {
    const res = db.query(
      `SELECT * FROM monitors WHERE plugin = ? AND name = ?`,
      [plugin, name]
    )

    if (res === Empty) {
      return null
    }

    return res.asObjects().next().value
  }

  public update(mon: Monitor): boolean {
    db.query(
      `UPDATE monitors SET enabled = ?, frequency = ?, runsOn = ?, params = ?`,
      [mon.enabled, mon.frequency, mon.runsOn.join(','), mon.params]
    )

    return db.changes > 0
  }
}
