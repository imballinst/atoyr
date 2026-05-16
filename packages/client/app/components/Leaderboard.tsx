import { useLeaderboard } from '../api/hooks';

export function Leaderboard() {
  const leaderboardQuery = useLeaderboard();

  return (
    <div className="text-dark-text-tertiary">
      <h2>Leaderboard</h2>

      {leaderboardQuery.isLoading ? (
        <div>Loading...</div>
      ) : leaderboardQuery.data ? (
        <ol>
          {leaderboardQuery.data.entries.map((entry, index) => (
            <li key={index}>
              {entry.user?.username ?? 'Unknown'}: {entry.score}
            </li>
          ))}
        </ol>
      ) : leaderboardQuery.isError ? (
        <div>Error loading leaderboard</div>
      ) : null}
    </div>
  );
}
