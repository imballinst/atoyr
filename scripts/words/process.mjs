import fs from "fs/promises";
import path from "path";

const PATH_TO_ORIGINAL_TXT = path.join(
  process.cwd(),
  "scripts/words/original.txt",
);
const PATH_TO_OUTPUT_TXT = path.join(process.cwd(), "scripts/words/output.txt");

const content = await fs.readFile(PATH_TO_ORIGINAL_TXT, "utf-8");
const filtered = content
  .trim()
  .split("\n")
  .filter((word) => word.length === 5 && /^[a-z]{5}$/.test(word))
  .sort();

await fs.writeFile(PATH_TO_OUTPUT_TXT, filtered.join("\n"), "utf-8");
