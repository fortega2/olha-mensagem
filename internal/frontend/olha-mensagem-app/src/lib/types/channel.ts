export type Channel = {
	id: number;
	name: string;
	description: string | null;
	createdBy: number;
	createdByUsername: string;
	createdAt: string;
};

export type CreateChannelRequest = {
	name: string;
	description?: string;
	userId: number;
};
