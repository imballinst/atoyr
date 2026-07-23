# AGS integration — handoff

Completed:

- IAM clients created: server-to-server (confidential `e50254a28f0740369981d194137abd14`) and web/public (`cecf5228881640ef88fc9a3fbc079cc3`)
- Server-to-server client permissions added: Cloud Save (Player Records CRU), Statistics (User Values CU), Leaderboard (Config/Data R, User Visibility R)
- AccelByte Go SDK installed and wired into `packages/server/cmd/service/main.go`
- `.env.local`/`.env.local.example` updated with AGS config

## Next steps

- **Create Cloud Save / Statistics / Leaderboard configurations in AGS** (`/ags manage-resource`) — set up stat codes (`atoyr_score_vanilla`, `atoyr_score_blind`, etc.) and leaderboard configs per the spec.
- **Implement the actual migration of `game.service.go`** endpoints to use AGS Cloud Save (round persistence), Statistics (post-game score/attempts/accuracy), and Leaderboards (top-N/percentile queries).
- **Decide on the leaderboard tie-breaker strategy** — composite score formula vs. AGS Extend override vs. keep custom post-processing in the Go server.
