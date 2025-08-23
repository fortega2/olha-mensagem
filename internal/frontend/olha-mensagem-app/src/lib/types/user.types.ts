export type UserDto = {
	id: string;
	username: string;
};

export type LoginForm = {
	username: string;
	password: string;
};

export type LoginRequest = {
	username: string;
	password: string;
};

export type RegisterForm = {
	username: string;
	password: string;
	confirmPassword: string;
};

export type RegisterRequest = {
	username: string;
	password: string;
};
