import type { LoginRequest, UserDto } from "$lib/types/user.types";

export class UserService {
    private readonly _baseUrl: string = 'http://localhost:8080';

    async login(username: string, password: string): Promise<UserDto> {
        const loginData: LoginRequest = { username, password };

        const response = await fetch(`${this._baseUrl}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginData)
        });

        if (!response.ok) {
            throw new Error(`Login failed: ${await response.text()}`);
        }

        const user: UserDto = await response.json();
        return user;
    }
}
