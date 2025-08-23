<script lang="ts">
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/button/button.svelte';
	import CardContent from '$lib/components/ui/card/card-content.svelte';
	import CardDescription from '$lib/components/ui/card/card-description.svelte';
	import CardHeader from '$lib/components/ui/card/card-header.svelte';
	import CardTitle from '$lib/components/ui/card/card-title.svelte';
	import Card from '$lib/components/ui/card/card.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { UserService } from '$lib/services/user.service';
	import type { AuthCredentials, UserDto } from '$lib/types/user.types';

	const loginForm: AuthCredentials = $state({
		username: '',
		password: ''
	});

	const handleSubmit = async (event: Event) => {
		event.preventDefault();

		try {
			const userService = new UserService();
			const user: UserDto = await userService.login(loginForm.username, loginForm.password);
			console.log('Login successful', user);
			goto('/');
		} catch (err: unknown) {
			if (err instanceof Error) {
				console.error('Error during login', err.message);
			} else {
				console.error('Unexpected error during login', err);
			}
		}
	};
	const handleRegisterRedirect = () => goto('/register');
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-100">
	<Card class="w-full max-w-md">
		<CardHeader class="text-center">
			<CardTitle>Login</CardTitle>
			<CardDescription>Log your credentials to access the chat.</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit}>
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input
						id="username"
						type="text"
						bind:value={loginForm.username}
						placeholder="Enter your username"
						required
					/>
				</div>
				<div class="mt-4 space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={loginForm.password}
						placeholder="Enter your password"
						required
					/>
				</div>
				<Button type="submit" class="mt-4 w-full cursor-pointer">Login</Button>
				<Button type="button" class="mt-2 w-full cursor-pointer" onclick={handleRegisterRedirect}
					>Register</Button
				>
			</form>
		</CardContent>
	</Card>
</div>
