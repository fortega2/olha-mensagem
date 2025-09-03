export type MessageType = 'Chat' | 'Notification';

export type ChatMessage = {
	type: MessageType;
	userId: number;
	username: string;
	content: string;
	timestamp: string;
	color: string;
};
