import { Leaderboard } from '~/components/Leaderboard';
import type { Route } from './+types/leaderboard';

export function meta({}: Route.MetaArgs) {
  return [{ title: 'Leaderboard | atoyr' }, { name: 'description', content: 'Welcome to React Router!' }];
}

export default function LeaderboardPage() {
  return <Leaderboard />;
}
