import { Leaderboard } from '~/components/Leaderboard';

export function meta() {
  return [{ title: 'Leaderboard | atoyr' }, { name: 'description', content: 'Welcome to React Router!' }];
}

export default function LeaderboardPage() {
  return <Leaderboard HeadingComponent="h1" />;
}
