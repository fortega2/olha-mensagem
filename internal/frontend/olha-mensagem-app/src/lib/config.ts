export const API_BASE = ((import.meta.env.VITE_API_BASE as string | undefined) || '') + '/api';

export function wsUrl(channelId: number, userId: number): string {
	const proto =
		typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';
	return `${proto}//${host}/api/ws/${channelId}/${userId}`;
}
