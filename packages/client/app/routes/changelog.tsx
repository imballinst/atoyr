import { marked } from 'marked';

import { changelogMarkdown } from '~/lib/changelog';
import { renderer } from '~/lib/markdown';

export function meta() {
  return [{ title: 'Changelog | Atoyr' }, { name: 'description', content: 'Changelog for the Atoyr game.' }];
}

export default function ChangelogPage() {
  const html = marked(changelogMarkdown, {
    renderer,
  });
  return <div dangerouslySetInnerHTML={{ __html: html }} className="text-dark-text-secondary w-full h-full" />;
}
