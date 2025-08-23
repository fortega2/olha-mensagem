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
	import type { RegisterForm, UserDto } from '$lib/types/user.types';
	import { toast } from 'svelte-sonner';

	const registerForm: RegisterForm = $state({
		username: '',
		password: '',
		confirmPassword: ''
	});

	const handleSubmit = async (event: Event) => {
		event.preventDefault();

		if (!validateForm()) {
			toast.error('Passwords do not match');
			return;
		}

		try {
			const userService = new UserService();
			const user: UserDto = await userService.register(
				registerForm.username,
				registerForm.password
			);
			sessionStorage.setItem('user', JSON.stringify(user));
			goto('/login');
		} catch (err: unknown) {
			err instanceof Error
				? toast.error(`Registration failed: ${err.message}`)
				: toast.error('Registration failed: Unexpected error');
		}
	};
	const validateForm = () => {
		if (registerForm.password !== registerForm.confirmPassword) {
			return false;
		} else {
			return true;
		}
	};
	const handleLoginRedirect = () => goto('/login');
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-100">
	<Card class="w-full max-w-md">
		<CardHeader class="text-center">
			<CardTitle>Register</CardTitle>
			<CardDescription>Register your credentials to access the chat.</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit}>
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input
						id="username"
						type="text"
						bind:value={registerForm.username}
						placeholder="Enter your username"
						required
					/>
				</div>
				<div class="mt-4 space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={registerForm.password}
						placeholder="Enter your password"
						required
					/>
				</div>
				<div class="mt-4 space-y-2">
					<Label for="confirmPassword">Confirm Password</Label>
					<Input
						id="confirmPassword"
						type="password"
						bind:value={registerForm.confirmPassword}
						placeholder="Confirm your password"
						required
					/>
				</div>
				<Button type="submit" class="mt-4 w-full cursor-pointer">Register</Button>
				<Button type="button" class="mt-2 w-full cursor-pointer" onclick={handleLoginRedirect}
					>Login</Button
				>
			</form>
		</CardContent>
	</Card>
</div>
