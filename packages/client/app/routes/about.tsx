import { marked, Renderer } from 'marked';

export function meta() {
  return [{ title: 'About | Atoyr' }, { name: 'description', content: 'A page containing info about the Atoyr game.' }];
}

const renderer = new Renderer();
renderer.heading = function ({ tokens, depth }) {
  const text = renderer.parser.parseInline(tokens);
  return `<h${depth} class="font-bold text-lg my-4 first:mt-0">${text}</h${depth}>`;
};
renderer.paragraph = function ({ tokens }) {
  const text = renderer.parser.parseInline(tokens);
  return `<p class="text-sm mb-3">${text}</p>`;
};
renderer.link = function ({ tokens, href, title }) {
  const text = renderer.parser.parseInline(tokens);
  const attrs: Record<string, string | undefined | null> = {
    title,
    href,
  };
  const attrStringified = Object.entries(attrs)
    .filter(([, val]) => !!val)
    .map(([key, value]) => `${key}="${value}"`)
    .join(' ');

  return `<a class="underline decoration-dotted" ${attrStringified}>${text}</a>`;
};

export default function LeaderboardPage() {
  const html = marked(getAboutPageMarkdown(), {
    renderer,
  });
  return <div dangerouslySetInnerHTML={{ __html: html }} className="text-dark-text-secondary w-full h-full" />;
}

function getAboutPageMarkdown() {
  return `
# About

This game was inspired by [Wordle](https://www.nytimes.com/games/wordle/index.html). However, instead of once-per-day, this has a time trial theme.

The more words you answered correctly, the better your accuracy is, and the earlier you get that score, you will place higher than others in the leaderboard.

Additionally, this game also tells you the "short meaning" of the scrambled word. Hopefully, it can be useful for you to learn new vocabularies.

## Dictionary

The collection of words in this game was sourced from https://github.com/david47k/top-english-wordlists, particularly the [top 10k words](https://github.com/david47k/top-english-wordlists/blob/master/top_english_words_lower_10000.txt).

It was then filtered to only include words with exactly 5 characters. After that, I used LLM to generate the definitions (because I'm not going to lie, manually working on 1000+ definitions is painful).

## Author and source

The repository is closed-source. I will appreciate it if you want to provide feedback, which you can drop to [my Twitter](https://twitter.com/imballinst).
`.trim();
}
