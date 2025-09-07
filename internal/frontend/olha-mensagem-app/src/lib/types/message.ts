import type { ChatMessage } from './websocket.types';

export type MessageDto = {
	id: number;
	channelId: number;
	userId: number;
	userUsername: string;
	userColor: string;
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
		color: message.userColor
	};
}
