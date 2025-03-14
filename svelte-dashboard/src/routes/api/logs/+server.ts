// src/routes/api/logs/+server.ts
import { getRedisClient } from '$lib/redis';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';

// Define the log entry type
interface LogEntry {
	host: string;
	ipaddr: string;
	timestamp: string;
	pid: string;
	thread: string;
	level: string;
	message: string;
}

export const GET: RequestHandler = async () => {
	try {
		const redisClient = await getRedisClient();

		// Fetch logs from Redis list
		// Get the 10 most recent log entries (if using LPUSH -10, -1 for RPUSH)
		const logs = await redisClient.lRange('log_entries', 0, 9);
		const parsedLogs: LogEntry[] = logs.map((log: string) => JSON.parse(log));

		return json(parsedLogs);
	} catch (error) {
		console.error('Error fetching logs from Redis:', error);
		return json({ error: 'Failed to fetch logs' }, { status: 500 });
	}
};
