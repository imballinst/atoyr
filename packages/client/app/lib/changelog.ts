export const changelogMarkdown = import.meta.env.VITE_CHANGELOG ?? '';

export function deriveReleaseDate(markdown: string): Date | null {
  const h1Match = markdown.match(/^# ([A-Z][a-z]+ \d{4})/m);
  if (!h1Match) return null;

  const restAfterH1 = markdown.slice(h1Match.index! + h1Match[0].length);
  const h2Match = restAfterH1.match(/^## Week ([1-5])/m);
  if (!h2Match) return null;

  const [monthName, yearStr] = h1Match[1].split(' ');
  const weekNumber = parseInt(h2Match[1], 10);
  const year = parseInt(yearStr, 10);
  const monthIndex = monthNames.indexOf(monthName);

  if (monthIndex === -1) return null;

  const releaseDay = (weekNumber - 1) * 7 + 1;
  return new Date(year, monthIndex, Math.min(releaseDay, daysInMonth(monthIndex, year)));
}

const monthNames = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

function daysInMonth(monthIndex: number, year: number): number {
  return new Date(year, monthIndex + 1, 0).getDate();
}
