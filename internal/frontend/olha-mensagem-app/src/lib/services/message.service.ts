import { API_BASE } from '$lib/config';
import type { MessageDto } from '$lib/types/message';

export class MessageService {
	private readonly _fullUrl: string = `${API_BASE}/messages`;
	private readonly _headers: HeadersInit = {
		'Content-Type': 'application/json'
	};

	public async getHistoryMessagesByChannel(channelId: number): Promise<MessageDto[]> {
		try {
			const response = await fetch(`${this._fullUrl}/history/${channelId}`, {
				method: 'GET',
				headers: this._headers
			});

			if (!response.ok) {
				throw new Error(await response.text());
			}

			const messages: MessageDto[] = await response.json();
			return messages;
		} catch (error: unknown) {
			throw new Error(
				`An error occurred while fetching history messages: ${error instanceof Error ? error.message : String(error)}`
			);
		}
	}
}
