import { Response } from './error';

// This file contains a function used in the queue.svelte component

// Function to pause or resume the current song
/**
 * @param {string} GuildID
 * @param {string} token
 * @param {string} action
 * @param {string} host
 * @returns {Promise<number>}
 */
export async function ToggleSong(GuildID, token, action = '', host) {
	// Request
	if (action === '') {
		return -8;
	} else {
		let route =
			`${host}/song/${action}/${GuildID}?` + new URLSearchParams({ token: token }).toString();
		let response = await fetch(route, {
			method: 'GET',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/x-www-form-urlencoded'
			}
		});

		// Error Handling
		switch (response.status) {
			case 200:
				return Response.SUCCESS;
			case 401:
				return Response.QUEUE_TOKEN_ERR;
			case 406:
				return Response.SONG_PAUSED_ERR;
			case 500:
				return Response.SONG_TOGGLE_ERR;
			default:
				return Response.SONG_TOGGLE_ERR;
		}
	}
}
