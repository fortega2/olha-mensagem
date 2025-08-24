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

	let ws: WebSocket | null = null;

	let user = $state<UserDto | null>(null);
	let messages = $state<ChatMessage[]>([]);
	let pendingMessage = $state('');
	let connecting = $state(true);

	let messagesContainer: HTMLDivElement | null = null;

	onMount(() => {
		if (typeof window === 'undefined') return;

		try {
			const stored = sessionStorage.getItem('user');
			if (!stored) {
				goto('/login');
				return;
			}
			user = JSON.parse(stored) as UserDto;
		} catch {
			goto('/login');
			return;
		}

		connectWebSocket();

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

	const connectWebSocket = () => {
		if (!user) return;
		connecting = true;
		ws = new WebSocket(`ws://localhost:8080/api/ws/${user.id}`);

		ws.onopen = () => {
			connecting = false;
			toast.info('Conectado');
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

	const formatTime = (ts: string) => {
		try {
			return new Date(ts).toLocaleTimeString();
		} catch {
			return '';
		}
	};
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-100 p-4">
	<Card class="flex h-[80vh] w-full max-w-2xl flex-col">
		<CardHeader class="pb-2">
			<CardTitle class="flex items-center justify-between text-lg">
				<span>Chat</span>
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
					<p class="text-gray-400 italic">Sin mensajes todavía.</p>
				{:else}
					{#each messages as m (m.timestamp + m.userId)}
						<div class="group">
							<div class="flex items-baseline gap-2">
								<span class="font-semibold" style={`color:${m.color}`}
									>#{m.userId} {m.username}</span
								>
								<span class="text-[10px] text-gray-400">{formatTime(m.timestamp)}</span>
							</div>
							<div class="mt-0.5 pl-1 break-words whitespace-pre-wrap">
								{m.content}
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<form class="mt-3 flex gap-2" onsubmit={handleSubmit}>
				<Input
					class="flex-1"
					placeholder="Escribe un mensaje..."
					bind:value={pendingMessage}
					disabled={connecting}
				/>
				<Button type="submit" class="shrink-0" disabled={connecting || !pendingMessage.trim()}
					>Enviar</Button
				>
			</form>
		</CardContent>
	</Card>
</div>
