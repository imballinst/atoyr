/**
 * @filename: lint-staged.config.js
 * @type {import('lint-staged').Configuration}
 */
export default {
  "*.{ts,tsx}": ["oxlint --fix --fix-suggestions", "oxfmt"],
  "packages/server/topics/indonesian-politician-quotes.json": [
    "node scripts/validate-quotes.mjs",
  ],
};
