// src/lib/redis.ts
import { createClient, type RedisClientType } from 'redis';

import { REDIS_HOST, REDIS_PORT, REDIS_PASSWORD } from '$env/static/private';

let redisClient: RedisClientType | null = null;

export async function getRedisClient(): Promise<RedisClientType> {
	if (!redisClient) {
		try {
			redisClient = createClient({
				socket: {
					host: REDIS_HOST || 'localhost',
					port: parseInt(REDIS_PORT || '6379', 10)
				},
				password: REDIS_PASSWORD || ''
			});

			await redisClient.connect().catch((err: Error) => {
				console.error('Failed to connect to Redis:', err);
				redisClient = null;
				throw err;
			});

			console.log('Connected to Redis successfully');
		} catch (error) {
			console.error('Error connecting to Redis:', error);
			throw error;
		}
	}

	return redisClient;
}
