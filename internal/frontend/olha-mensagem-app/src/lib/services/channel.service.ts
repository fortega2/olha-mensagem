import { API_CHANNELS_BASE } from '$lib/config';
import type { Channel, CreateChannelRequest } from '$lib/types/channel';

export class ChannelService {
	private readonly _fullUrl: string = API_CHANNELS_BASE;
	private readonly _headers: HeadersInit = {
		'Content-Type': 'application/json'
	};

	public async getAllChannels(): Promise<Channel[]> {
		try {
			const response = await fetch(this._fullUrl, {
				method: 'GET',
				headers: this._headers
			});

			if (!response.ok) {
				throw new Error(await response.text());
			}

			const channels: Channel[] = await response.json();
			return channels;
		} catch (error: unknown) {
			throw new Error(
				`An error occurred while fetching channels: ${error instanceof Error ? error.message : String(error)}`
			);
		}
	}

	public async createChannel(req: CreateChannelRequest): Promise<Channel> {
		try {
			const response = await fetch(this._fullUrl, {
				method: 'POST',
				headers: this._headers,
				body: JSON.stringify(req)
			});

			if (!response.ok) {
				throw new Error(await response.text());
			}

			const channel: Channel = await response.json();
			return channel;
		} catch (error: unknown) {
			throw new Error(
				`An error occurred while creating the channel: ${error instanceof Error ? error.message : String(error)}`
			);
		}
	}

	public async deleteChannel(channelId: number, userId: number): Promise<void> {
		try {
			const response = await fetch(`${this._fullUrl}/${channelId}/users/${userId}`, {
				method: 'DELETE',
				headers: this._headers
			});

			if (!response.ok) {
				throw new Error(await response.text());
			}

			return;
		} catch (error: unknown) {
			throw new Error(
				`An error occurred while deleting the channel: ${error instanceof Error ? error.message : String(error)}`
			);
		}
	}
}
