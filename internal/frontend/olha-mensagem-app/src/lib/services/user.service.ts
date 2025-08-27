import { API_USERS_BASE } from '$lib/config';
import type { AuthCredentials, UserDto } from '$lib/types/user.types';

export class UserService {
	private readonly _fullUrl: string = API_USERS_BASE;
	private readonly _headers: HeadersInit = {
		'Content-Type': 'application/json'
	};

	public async login(username: string, password: string): Promise<UserDto> {
		const loginData: AuthCredentials = { username, password };

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
		const registerData: AuthCredentials = { username, password };

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
