import { type RouteConfig, index, route } from '@react-router/dev/routes';

const routes = [index('routes/home.tsx'), route('/leaderboard', 'routes/leaderboard.tsx')] satisfies RouteConfig;
if (import.meta.env.DEV) {
  routes.push(route('/playground', 'routes/playground.tsx'));
}

export default routes;
