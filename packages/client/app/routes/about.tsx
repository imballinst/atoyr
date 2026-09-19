import { marked } from 'marked';

import { renderer } from '~/lib/markdown';

export function meta() {
  return [{ title: 'About | Atoyr' }, { name: 'description', content: 'A page containing info about the Atoyr game.' }];
}

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

The more words you answer correctly, the better your accuracy is, and the earlier you reach that score, the higher you will place on the leaderboard.

This game also tells you the "short meaning" of the scrambled word; hopefully it can be useful for learning new vocabulary.

## Words per topic

### English words

The collection of words in this game was sourced from [WordNet 3.1](https://wordnet.princeton.edu/download), a large lexical database of English from Princeton University.

From there, the words were filtered to 5-letter lowercase lemmas across noun, verb, adjective, and adverb. Definitions come straight from WordNet's glosses; per word, the most general sense is preferred (noun > adjective > verb > adverb), and proper-name senses (the noun.person lexical file) are excluded so no real people end up as answers.

Reference: "About WordNet." WordNet, Princeton University, 2010, [wordnet.princeton.edu](https://wordnet.princeton.edu/)

### Indonesian Politician Quotes

The quotes were extracted from online publications (mostly news, with credible sources being preferred). When the game ends, you will be able to see the source of the quote and validate it yourself.

## Source

You can view the source code of this game at https://github.com/imballinst/atoyr. If you have feedback or ideas, please open an issue first.
`.trim();
}
