<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		if (!sessionStorage.getItem('user')) {
			goto('/login');
		} else {
			goto('/channels');
		}
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-100">
	<div class="text-center">
		<p class="text-lg text-gray-600">Redirecting...</p>
	</div>
</div>
