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
	import type { UserDto } from '$lib/types/user.types';
	import type { ChatMessage } from '$lib/types/websocket.types';
	import { wsUrl } from '$lib/config';
	import type { Channel } from '$lib/types/channel';
	import { ArrowLeft } from '@lucide/svelte';
	import { MessageService } from '$lib/services/message.service';
	import { messageDtoToChatMessage, type MessageDto } from '$lib/types/message';

	let ws: WebSocket | null = null;

	let user = $state<UserDto | null>(null);
	let selectedChannel = $state<Channel | null>(null);
	let messages = $state<ChatMessage[]>([]);
	let pendingMessage = $state('');
	let connecting = $state(true);

	let messagesContainer: HTMLDivElement | null = null;

	const messageSrv: MessageService = new MessageService();

	onMount(() => {
		if (typeof window === 'undefined') return;

		try {
			const storedUser = sessionStorage.getItem('user');
			if (!storedUser) {
				goto('/login');
				return;
			}
			user = JSON.parse(storedUser) as UserDto;

			const storedChannel = sessionStorage.getItem('selectedChannel');
			if (!storedChannel) {
				goto('/channels');
				return;
			}
			selectedChannel = JSON.parse(storedChannel) as Channel;

			loadHistoryMessages(selectedChannel.id);
			connectWebSocket();
		} catch {
			goto('/login');
			return;
		}

		return () => {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.close(1000, 'Component unmounted');
			}
		};
	});

	$effect(() => {
		if (messagesContainer) {
			messagesContainer.scrollTop = messagesContainer.scrollHeight;
		}
	});

	const loadHistoryMessages = async (channelId: number) => {
		try {
			const historyMsg = await messageSrv.getHistoryMessagesByChannel(channelId);
			messages = historyMsg.map((m: MessageDto) => messageDtoToChatMessage(m));
		} catch (err: unknown) {
			toast.error(
				`${err instanceof Error ? err.message : 'Error while fetching history messages'}`
			);
		}
	};

	const connectWebSocket = () => {
		if (!user || !selectedChannel) return;

		connecting = true;
		ws = new WebSocket(wsUrl(selectedChannel.id, Number(user.id)));

		ws.onopen = () => {
			connecting = false;
			toast.success(`Conectado al canal "${selectedChannel?.name}"`);
		};

		ws.onerror = (e) => {
			toast.error('Error WebSocket');
			console.error(e);
		};

		ws.onclose = (ev) => {
			toast.info('Conexión cerrada');
			if (ev.code !== 1000) {
				setTimeout(() => {
					toast.message('Reintentando...');
					connectWebSocket();
				}, 2000);
			}
		};

		ws.onmessage = (evt: MessageEvent<string>) => {
			try {
				const msg: ChatMessage = JSON.parse(evt.data);
				messages.push(msg);
			} catch (err) {
				toast.error(`Mensaje inválido: ${err instanceof Error ? err.message : ''}`);
			}
		};
	};

	const sendMessage = () => {
		if (!pendingMessage.trim()) return;
		if (!ws || ws.readyState !== WebSocket.OPEN) {
			toast.error('No conectado');
			return;
		}

		ws.send(pendingMessage.trim());
		pendingMessage = '';
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		sendMessage();
	};

	const goBackToChannels = () => {
		goto('/channels');
	};

	const formatTime = (ts: string) => {
		try {
			return new Date(ts).toLocaleTimeString();
		} catch {
			return ts;
		}
	};
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-100 p-4">
	<Card class="flex h-[80vh] w-full max-w-2xl flex-col">
		<CardHeader class="pb-2">
			<CardTitle class="flex items-center justify-between text-lg">
				<div class="flex items-center gap-3">
					<Button variant="ghost" size="sm" class="cursor-pointer" onclick={goBackToChannels}>
						<ArrowLeft size={16} />
					</Button>
					<div>
						<span class="text-lg font-semibold"># {selectedChannel?.name || 'Canal'}</span>
						{#if selectedChannel?.description}
							<p class="text-sm font-normal text-gray-600">{selectedChannel.description}</p>
						{/if}
					</div>
				</div>
			</CardTitle>
		</CardHeader>
		<CardContent class="flex flex-1 flex-col overflow-hidden pt-0">
			<div
				bind:this={messagesContainer}
				class="flex-1 space-y-3 overflow-y-auto rounded border bg-white p-3 text-sm"
			>
				{#if connecting}
					<p class="text-gray-500 italic">Conectando...</p>
				{:else if messages.length === 0}
					<p class="text-gray-400 italic">
						No hay mensajes en #{selectedChannel?.name || 'este canal'} todavía.
					</p>
				{:else}
					{#each messages as m, index (m.userId + '-' + m.timestamp + '-' + index)}
						{#if m.type === 'Chat'}
							<div class="group">
								<div class="flex items-baseline gap-2">
									<span class="font-semibold" style={`color:${m.color}`}>{m.username}</span>
									<span class="text-[10px] text-gray-400">{formatTime(m.timestamp)}</span>
								</div>
								<div class="mt-0.5 pl-1 break-words whitespace-pre-wrap">
									{m.content}
								</div>
							</div>
						{:else}
							<div class="text-center text-xs text-gray-500 italic">
								{m.content} <span class="text-[10px]">({formatTime(m.timestamp)})</span>
							</div>
						{/if}
					{/each}
				{/if}
			</div>

			<form class="mt-3 flex gap-2" onsubmit={handleSubmit}>
				<Input
					class="flex-1"
					placeholder={`Mensaje para #${selectedChannel?.name || 'canal'}...`}
					bind:value={pendingMessage}
					disabled={connecting}
				/>
				<Button
					type="submit"
					class="shrink-0 cursor-pointer"
					disabled={connecting || !pendingMessage.trim()}
				>
					Enviar
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
