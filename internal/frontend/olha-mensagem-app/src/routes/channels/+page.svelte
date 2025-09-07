<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Card from '$lib/components/ui/card/card.svelte';
	import CardHeader from '$lib/components/ui/card/card-header.svelte';
	import CardTitle from '$lib/components/ui/card/card-title.svelte';
	import CardContent from '$lib/components/ui/card/card-content.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { ChannelService } from '$lib/services/channel.service';
	import type { UserDto } from '$lib/types/user.types';
	import type { Channel } from '$lib/types/channel';
	import { Plus, Users, Calendar, User, RefreshCwIcon } from '@lucide/svelte';

	let user = $state<UserDto | null>(null);
	let channels = $state<Channel[]>([]);
	let loading = $state(true);
	let showCreateForm = $state(false);
	let creating = $state(false);

	let newChannelName = $state('');
	let newChannelDescription = $state('');

	const channelSrv = new ChannelService();

	onMount(async () => {
		if (typeof window === 'undefined') return;

		try {
			const stored = sessionStorage.getItem('user');
			if (!stored) {
				goto('/login');
				return;
			}
			user = JSON.parse(stored) as UserDto;
			await loadChannels();
		} catch {
			goto('/login');
			return;
		}
	});

	const loadChannels = async () => {
		try {
			loading = true;
			channels = await channelSrv.getAllChannels();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Error loading channels');
		} finally {
			loading = false;
		}
	};

	const createChannel = async (e: Event) => {
		e.preventDefault();

		if (!user || !newChannelName.trim()) return;

		try {
			creating = true;
			const newChannel = await channelSrv.createChannel({
				name: newChannelName.trim(),
				description: newChannelDescription.trim() ?? undefined,
				userId: Number(user.id)
			});

			channels = [newChannel, ...channels];
			newChannelName = '';
			newChannelDescription = '';
			showCreateForm = false;
			toast.success(`Channel "${newChannel.name}" created`);
		} catch (error: unknown) {
			toast.error(error instanceof Error ? error.message : 'Error creating channel');
		} finally {
			creating = false;
		}
	};

	const joinChannel = (channel: Channel) => {
		sessionStorage.setItem('selectedChannel', JSON.stringify(channel));
		goto('/chat');
	};

	const deleteChannel = async (channel: Channel) => {
		if (!user) return;

		if (Number(user.id) !== channel.createdBy) {
			toast.error('Only the creator can delete this channel');
			return;
		}

		if (!confirm(`Are you sure you want to delete the channel "${channel.name}"?`)) {
			return;
		}

		try {
			await channelSrv.deleteChannel(channel.id, Number(user.id));
			channels = channels.filter((c) => c.id !== channel.id);
			toast.success(`Channel "${channel.name}" deleted`);
		} catch (error: unknown) {
			toast.error(error instanceof Error ? error.message : 'Error deleting channel');
		}
	};

	const formatDate = (dateString: string) => {
		try {
			return new Date(dateString).toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			});
		} catch {
			return 'Invalid date';
		}
	};

	const logout = () => {
		sessionStorage.removeItem('user');
		sessionStorage.removeItem('selectedChannel');
		goto('/login');
	};
</script>

<svelte:head>
	<title>Channels - Chat</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 p-4">
	<div class="mx-auto w-full">
		<div class="mb-8 flex items-center justify-between">
			<div>
				<h1 class="text-3xl font-bold text-gray-900">Chat Channels</h1>
				{#if user}
					<p class="text-gray-600">
						Welcome, <span class="font-semibold">{user.username}</span>
					</p>
				{/if}
			</div>
			<div class="flex gap-3">
				<Button
					onclick={() => loadChannels()}
					class="flex cursor-pointer items-center gap-2"
					disabled={loading}
				>
					<RefreshCwIcon size={16} class={loading ? 'animate-spin' : ''} />
					{loading ? 'Loading...' : 'Refresh'}
				</Button>
				<Button
					onclick={() => (showCreateForm = !showCreateForm)}
					class="flex cursor-pointer items-center gap-2"
				>
					<Plus size={16} />
					Create Channel
				</Button>
				<Button variant="outline" class="cursor-pointer" onclick={logout}>Logout</Button>
			</div>
		</div>

		{#if showCreateForm}
			<Card class="mb-6">
				<CardHeader>
					<CardTitle>Create New Channel</CardTitle>
				</CardHeader>
				<CardContent>
					<form class="space-y-4" onsubmit={createChannel}>
						<div>
							<Label for="channelName">Channel Name*</Label>
							<Input
								id="channelName"
								bind:value={newChannelName}
								placeholder="Ej: General"
								required
								disabled={creating}
							/>
						</div>
						<div>
							<Label for="channelDescription">Description</Label>
							<Input
								id="channelDescription"
								bind:value={newChannelDescription}
								placeholder="General channel for all topics"
								disabled={creating}
							/>
						</div>
						<div class="flex gap-2">
							<Button
								type="submit"
								class="cursor-pointer"
								disabled={creating || !newChannelName.trim()}
							>
								{creating ? 'Creating...' : 'Create Channel'}
							</Button>
							<Button
								variant="outline"
								class="cursor-pointer"
								onclick={() => (showCreateForm = false)}
								disabled={creating}
							>
								Cancel
							</Button>
						</div>
					</form>
				</CardContent>
			</Card>
		{/if}

		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="text-gray-500">Loading channels...</div>
			</div>
		{:else if channels.length === 0}
			<Card>
				<CardContent class="py-12 text-center">
					<Users size={48} class="mx-auto mb-4 text-gray-400" />
					<h3 class="mb-2 text-lg font-semibold text-gray-900">No channels available</h3>
					<p class="mb-4 text-gray-600">Be the first to create a channel!</p>
					<Button class="cursor-pointer" onclick={() => (showCreateForm = true)}>
						<Plus size={16} class="mr-2" />
						Create Channel
					</Button>
				</CardContent>
			</Card>
		{:else}
			<div class="grid grid-cols-1 gap-4">
				{#each channels as channel (channel.id)}
					<Card class="group transition-shadow hover:shadow-lg">
						<CardContent class="p-6">
							<div class="mb-4">
								<h3 class="text-lg font-semibold text-gray-900">{channel.name}</h3>
								{#if channel.description}
									<p class="mt-1 text-sm text-gray-600">{channel.description}</p>
								{/if}
							</div>

							<div class="mb-4 space-y-2 text-sm text-gray-500">
								<div class="flex items-center gap-2">
									<Calendar size={14} />
									<span>Created {formatDate(channel.createdAt)}</span>
								</div>
								<div class="flex items-center gap-2">
									<User size={14} />
									<span>Creator: {channel.createdByUsername}</span>
								</div>
							</div>

							<div class="flex justify-between">
								<Button
									class="w-[15%] cursor-pointer sm:w-[12%] lg:w-[10%]"
									onclick={() => joinChannel(channel)}>Connect</Button
								>
								{#if user && Number(user.id) === channel.createdBy}
									<Button
										variant="destructive"
										class="w-[12%] cursor-pointer sm:w-[10%] lg:w-[8%]"
										onclick={() => deleteChannel(channel)}
									>
										Delete
									</Button>
								{/if}
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{/if}
	</div>
</div>
