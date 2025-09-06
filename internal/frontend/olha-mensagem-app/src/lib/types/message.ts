import type { ChatMessage } from './websocket.types';

export type MessageDto = {
	id: number;
	channelId: number;
	userId: number;
	userUsername: string;
	content: string;
	timestamp: string;
};

export function messageDtoToChatMessage(message: MessageDto): ChatMessage {
	return {
		type: 'Chat',
		userId: message.userId,
		username: message.userUsername,
		content: message.content,
		timestamp: message.timestamp,
		color: '#000000'
	};
}
