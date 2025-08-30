export const API_BASE = (import.meta.env.VITE_API_BASE as string | undefined) || '';

export const API_USERS_BASE = `${API_BASE}/api/users`;

export function wsUrl(userId: number): string {
	const proto =
		typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';
	return `${proto}//${host}/api/ws/${userId}`;
}
