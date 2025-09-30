# A Toy R — Development README

This repository is a prototype bootstrap for the game described in `DESIGN.md`.

Workspaces:

- packages/client — Vite + React prototype UI
- packages/server — Express server with matchmaking and QTE resolve endpoints (mocked)
- packages/shared — Shared game logic (QTE, utility functions)

How to run locally (requires Yarn modern):

```bash
# from repository root
yarn install
yarn dev
```

This will start the client on :5173 and the server on :4000. The AccelByte integration is not implemented — server endpoints are stubs to show where you'd wire in AccelByte Gaming Services and Multiplayer Servers.
