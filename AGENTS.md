## A Test of Your Reflexes

A Test of Your Reflexes, or Atoyr, is a game where the users will see 5 characters in a random order and they will need to re-order those into a correct word, given a definition of the word. For the specs, see the `.opencode/specs` folder for more details. For the skills, see the `.opencode/skills` folder.

## Conventions

- Use TypeScript (except for the `packages/server` which is using Go).
- Use TailwindCSS (v4)
- Use Prettier for formatting
- Use Yarn Modern with nodeLinker node_modules
- Use latest React, no need for useCallback and useMemo
- DO NOT PUT UNNECESSARY COMMENTS between lines unless absolutely necessary. Also don't put unnecessary JSDoc as well for the emitted functions unless the intentions are not clear.
- DO NOT SPLIT INTO MULTIPLE COMPONENTS unless absolutely necessary. If it's possible to colocate the components, co-locate.
- When writing specs, put AS LITTLE DETAIL AS POSSIBLE to the implementation details. Just have the higher level; only show code snippets when necessary.
- DO NOT put overly-detailed file structure (apart from top-level ones) because it has potential to change over time.
- DO NOT create additional Markdown files in `.opencode` folder unless otherwise stated.

## Structure

- `packages/client`: The UI, this is the one that the user sees. It communicates with `packages/server` with HTTP and SSE.
- `packages/server`: The server, this is the one that orchestrates the events, words, and stuff.
- `packages/shared`: Shared libraries that is used by both client and server.

## Tests

Run top level `yarn test` to run all tests in all packages. Otherwise, use `yarn workspaces <folder_name>` to run individual tests. If possible, ALWAYS add unit tests with `vitest` for any logic-related functionalities. For UI related functionalities (such as CSS), it is not necessary unless otherwise stated.
