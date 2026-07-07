import { Leaderboard } from '~/components/Leaderboard';

export function meta() {
  return [{ title: 'Leaderboard | Atoyr' }, { name: 'description', content: 'Leaderboard of the Atoyr game.' }];
}

export default function LeaderboardPage() {
  return <Leaderboard HeadingComponent="h1" />;
}
