export const UI_ROUTES = [
  '/', '/login', '/dashboard', '/listeners', '/users', '/traffic', '/logs',
  '/routing', '/cluster', '/config', '/settings', '/change-password',
] as const;

export type UiRoute = typeof UI_ROUTES[number];
