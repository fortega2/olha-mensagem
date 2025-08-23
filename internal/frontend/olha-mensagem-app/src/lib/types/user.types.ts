export type UserDto = {
	id: string;
	username: string;
};

export type AuthCredentials = {
	username: string;
	password: string;
};

export type RegisterForm = AuthCredentials & {
	confirmPassword: string;
};
