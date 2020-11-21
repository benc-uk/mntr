# Mntr

A polyglot experiment in agentless web & system monitoring

## 💥 Highly unstable and in very active development

Consider this code base extremely volatile and in constant state of flux until an alpha version is ready

Components:

- Collector: Runs monitors on a schedule and sends results to server. Written in Go
- Plugins: Monitor 'task runners' loaded by the collector, such as `plugins/web` for HTTP and web content monitoring. Also written in Go
- Server: Backend API server and data store. Written in Deno with TypeScript
- Frontend: Web UI for reporting and viewing data. Written in Vue.js
