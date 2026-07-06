import { Leaderboard } from '~/components/Leaderboard';

export function meta() {
  return [{ title: 'Leaderboard | atoyr' }, { name: 'description', content: 'Leaderboard of the atoyr game.' }];
}

export default function LeaderboardPage() {
  return <Leaderboard HeadingComponent="h1" />;
}
