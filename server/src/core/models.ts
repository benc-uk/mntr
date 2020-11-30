export interface Monitor {
  name: string
  plugin: string
  enabled: boolean
  runsOn: string[]
  frequency: number
  params: string
}
