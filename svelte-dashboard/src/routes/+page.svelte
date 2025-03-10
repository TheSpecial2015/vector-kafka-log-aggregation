<script lang="ts">
	import { onMount } from 'svelte';

	type Log = {
		host: string;
		ipaddr: string;
		date: string;
		time: string;
		pid: string;
		thread: string;
		level: string;
		message: string;
	};

	let logs = $state<Log[]>([]);

	// Function to get color based on log level
	function getLevelColor(level: string) {
		switch (level) {
			case 'ERROR':
				return 'bg-rose-500 text-white';
			case 'WARN':
				return 'bg-amber-500 text-white';
			case 'INFO':
				return 'bg-sky-500 text-white';
			case 'DEBUG':
				return 'bg-emerald-500 text-white';
			default:
				return 'bg-slate-500 text-white';
		}
	}

	// Search and filter state
	let searchQuery = $state<string>('');
	let selectedLevel = $state<string>('');

	// Time formatting
	function formatDateTime(date: string, time: string) {
		return `${date} · ${time.split('.')[0]}`;
	}

	// Filter logs based on search and level
	let filteredLogs = $state<Log[]>([]);
	$effect(() => {
		filteredLogs = logs.filter((log) => {
			const matchesSearch =
				searchQuery === '' ||
				log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
				log.thread.toLowerCase().includes(searchQuery.toLowerCase()) ||
				log.host.toLowerCase().includes(searchQuery.toLowerCase());

			const matchesLevel = selectedLevel === '' || log.level === selectedLevel;

			return matchesSearch && matchesLevel;
		});
	});

	// Get unique values for level filter
	let levels = $state<string[]>([]);
	$effect(() => {
		levels = [...new Set(logs.map((log) => log.level))];
	});

	// Dark mode toggle
	let darkMode = $state<boolean>(true);

	// Function to establish WebSocket connection and listen for logs
	function connectWebSocket() {
		const socket = new WebSocket('ws://localhost:8080/ws');
		socket.onmessage = (event) => {
			try {
				const newLog: Log = JSON.parse(event.data);
				logs.push(newLog);
				logs = logs.slice(0, 10); // Keep only the last 10 logs
			} catch (error) {
				console.error('Error parsing WebSocket message:', error);
			}
		};

		socket.onopen = () => console.log('Connected to WebSocket');
		socket.onerror = (error) => console.error('WebSocket Error:', error);
		socket.onclose = () => {
			console.warn('WebSocket closed. Reconnecting in 3 seconds...');
			setTimeout(connectWebSocket, 3000); // Auto-reconnect on close
		};
	}

	// Call function to establish WebSocket connection

	onMount(() => {
		connectWebSocket();
	});
</script>

<div
	class={`font-display min-h-screen ${darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'}`}
>
	<div class="mx-auto max-w-screen-xl p-4">
		<!-- Header area -->
		<div class="mb-6 flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
			<div>
				<h1 class="text-2xl font-bold tracking-tight">System Logs</h1>
				<p class={`mt-1 ${darkMode ? 'text-slate-400' : 'text-slate-500'}`}>
					Real-time monitoring dashboard
				</p>
			</div>

			<!-- Controls -->
			<div class="flex items-center space-x-4">
				<button
					onclick={() => (darkMode = !darkMode)}
					class={`rounded-full p-2 ${darkMode ? 'bg-slate-800 text-slate-200' : 'border border-slate-200 bg-white text-slate-700 shadow-sm'}`}
				>
					{#if darkMode}
						<!-- Sun icon -->
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							viewBox="0 0 20 20"
							fill="currentColor"
						>
							<path
								fill-rule="evenodd"
								d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z"
								clip-rule="evenodd"
							/>
						</svg>
					{:else}
						<!-- Moon icon -->
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							viewBox="0 0 20 20"
							fill="currentColor"
						>
							<path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
						</svg>
					{/if}
				</button>

				<a
					href={null}
					class={`hidden rounded-md px-4 py-2 text-sm font-medium sm:inline-flex ${darkMode ? 'bg-slate-800 text-slate-200 hover:bg-slate-700' : 'border border-slate-200 bg-white text-slate-700 shadow-sm hover:bg-slate-50'}`}
				>
					Export Logs
				</a>
			</div>
		</div>

		<!-- Search and filters -->
		<div
			class={`mb-6 rounded-lg p-4 ${darkMode ? 'bg-slate-800' : 'border border-slate-200 bg-white shadow-sm'}`}
		>
			<div class="flex flex-col gap-4 sm:flex-row">
				<div class="relative flex-grow">
					<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
						<svg
							class={`h-5 w-5 ${darkMode ? 'text-slate-500' : 'text-slate-400'}`}
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							aria-hidden="true"
						>
							<path
								fill-rule="evenodd"
								d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
								clip-rule="evenodd"
							/>
						</svg>
					</div>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search logs..."
						class={`block w-full rounded-md border-0 py-2 pl-10 ${
							darkMode
								? 'bg-slate-700 text-white placeholder-slate-400 focus:ring-slate-500'
								: 'border-slate-200 bg-white text-slate-900 placeholder-slate-400 focus:ring-slate-500'
						}`}
					/>
				</div>

				<div class="sm:w-48">
					<select
						bind:value={selectedLevel}
						class={`block w-full rounded-md border-0 py-2 ${
							darkMode
								? 'bg-slate-700 text-white focus:ring-slate-500'
								: 'border-slate-200 bg-white text-slate-900 focus:ring-slate-500'
						}`}
					>
						<option value="">All levels</option>
						{#each levels as level}
							<option value={level}>{level}</option>
						{/each}
					</select>
				</div>
			</div>
		</div>

		<!-- Log list -->
		<div
			class={`overflow-hidden rounded-lg ${darkMode ? 'bg-slate-800' : 'border border-slate-200 bg-white shadow-sm'}`}
		>
			{#if filteredLogs.length > 0}
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class={darkMode ? 'bg-slate-700' : 'bg-slate-50'}>
							<tr>
								<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
									>Level</th
								>
								<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
									>Timestamp</th
								>
								<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
									>Host / IP</th
								>
								<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
									>Process</th
								>
								<th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider"
									>Message</th
								>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-200 dark:divide-slate-700">
							{#each filteredLogs as log}
								<tr class={`hover:${darkMode ? 'bg-slate-700/50' : 'bg-slate-50'}`}>
									<td class="whitespace-nowrap px-4 py-3">
										<span
											class={`rounded-md px-2 py-1 text-xs font-medium ${getLevelColor(log.level)}`}
										>
											{log.level}
										</span>
									</td>
									<td
										class={`whitespace-nowrap px-4 py-3 text-sm ${darkMode ? 'text-slate-300' : 'text-slate-600'}`}
									>
										{formatDateTime(log.date, log.time)}
									</td>
									<td class="px-4 py-3 text-sm">
										<div class="font-medium">{log.host}</div>
										<div class={darkMode ? 'text-slate-400' : 'text-slate-500'}>{log.ipaddr}</div>
									</td>
									<td class="px-4 py-3 text-sm">
										<div class="font-medium">PID: {log.pid}</div>
										<div
											class={`max-w-[150px] truncate ${darkMode ? 'text-slate-400' : 'text-slate-500'}`}
											title={log.thread}
										>
											{log.thread}
										</div>
									</td>
									<td class={`px-4 py-3 text-sm ${darkMode ? 'text-slate-300' : 'text-slate-700'}`}>
										<div class="max-w-md truncate" title={log.message}>{log.message}</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else}
				<div class="py-16 text-center">
					<svg
						class={`mx-auto h-12 w-12 ${darkMode ? 'text-slate-600' : 'text-slate-400'}`}
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h3 class="mt-2 text-sm font-medium">No logs found</h3>
					<p class={`mt-1 text-sm ${darkMode ? 'text-slate-400' : 'text-slate-500'}`}>
						Try adjusting your search or filters.
					</p>
				</div>
			{/if}
		</div>

		<!-- Status bar -->
		<div class="mt-4 flex items-center justify-between text-sm">
			<div class={darkMode ? 'text-slate-400' : 'text-slate-500'}>
				{filteredLogs.length} logs found
			</div>
			<button
				class={`text-sm ${darkMode ? 'text-slate-300 hover:text-white' : 'text-slate-600 hover:text-slate-900'}`}
			>
				Refresh logs
			</button>
		</div>
	</div>
</div>
