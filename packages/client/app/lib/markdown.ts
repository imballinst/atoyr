import { Renderer } from 'marked';

const renderer = new Renderer();
renderer.heading = function ({ tokens, depth }) {
  const text = renderer.parser.parseInline(tokens);
  let className = 'text-lg';
  if (depth === 2) {
    className = 'text-base';
  }

  return `<h${depth} class="font-bold ${className} my-4 first:mt-0">${text}</h${depth}>`;
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
renderer.list = function ({ ordered, items }) {
  const itemsHtml = items
    .map((item) => {
      const text = item.tokens.map((t) => renderer.parser.parseInline([t])).join('');
      return `<li class="text-sm mb-3">${text}</li>`;
    })
    .join('');
  const tag = ordered ? 'ol' : 'ul';
  return `<${tag} class="list-disc list-inside mb-3">${itemsHtml}</${tag}>`;
};

export { renderer };
