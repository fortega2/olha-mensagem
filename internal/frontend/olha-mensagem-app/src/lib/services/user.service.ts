import type { LoginRequest, RegisterRequest, UserDto } from '$lib/types/user.types';

export class UserService {
	private readonly _baseUrl: string = 'http://localhost:8080/api';
	private readonly _resourceName: string = 'users';
	private readonly _fullUrl: string = `${this._baseUrl}/${this._resourceName}`;
	private readonly _headers: HeadersInit = {
		'Content-Type': 'application/json'
	};

	public async login(username: string, password: string): Promise<UserDto> {
		const loginData: LoginRequest = { username, password };

		const response = await fetch(`${this._fullUrl}/login`, {
			method: 'POST',
			headers: this._headers,
			body: JSON.stringify(loginData)
		});

		if (!response.ok) {
			throw new Error(`Login failed: ${await response.text()}`);
		}

		const user: UserDto = await response.json();
		return user;
	}

	public async register(username: string, password: string): Promise<UserDto> {
		const registerData: RegisterRequest = { username, password };

		const response = await fetch(this._fullUrl, {
			method: 'POST',
			headers: this._headers,
			body: JSON.stringify(registerData)
		});

		if (!response.ok) {
			throw new Error(`Registration failed: ${await response.text()}`);
		}

		const user: UserDto = await response.json();
		return user;
	}
}
