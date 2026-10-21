import { Response } from './error';

// Function to add a song to the favorites
/**
 * @param {string} token
 * @param {string} host
 * @returns {Promise<any>}
 */
export async function AddFavorite(token, host) {
	let name =
		/** @type {HTMLInputElement | null} */ (document.getElementById('name'))?.value.trim() ?? '';
	let link =
		/** @type {HTMLInputElement | null} */ (document.getElementById('link'))?.value.trim() ?? '';
	let folder =
		/** @type {HTMLInputElement | null} */ (document.getElementById('folder'))?.value.trim() ?? '';

	// Request
	let route = `${host}/favorites`;
	let response = await fetch(route, {
		method: 'POST',
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/x-www-form-urlencoded'
		},
		body: new URLSearchParams({
			name: name,
			link: link,
			folder: folder,
			token: token
		}).toString()
	});

	//Error Handling
	switch (response.status) {
		case 200:
			return [{ name: name, link: link, folder: folder }];
		case 401:
			return Response.FAVORITE_TOKEN_ERROR;
		case 500:
			return Response.DUPLICATE_ERROR;
		default:
			return Response.FAVORITE_ERROR;
	}
}

// Function to remove a song from the favorites
/**
 * @param {string} token
 * @param {string} name
 * @param {string} host
 * @returns {Promise<number>}
 */
export async function RemoveFavorite(token, name, host) {
	// Request
	let route = `${host}/favorites?` + new URLSearchParams({ name: name, token: token }).toString();
	let response = await fetch(route, {
		method: 'DELETE',
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/x-www-form-urlencoded'
		},
		body: new URLSearchParams({
			name: name,
			token: token
		}).toString()
	});

	//Error Handling
	switch (response.status) {
		case 200:
			return Response.SUCCESS;
		case 401:
			return Response.FAVORITE_TOKEN_ERROR;
		case 500:
			return Response.FAVORITE_ERROR;
		default:
			return Response.FAVORITE_ERROR;
	}
}

// Function to get the favorite songs
/**
 * @param {string} token
 * @param {string} host
 * @returns {Promise<any>}
 */
export async function GetFavorites(token, host) {
	// Request
	let route = `${host}/favorites?` + new URLSearchParams({ token: token }).toString();
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
			return await response.json();
		case 401:
			return Response.FAVORITE_TOKEN_ERROR;
		default:
			return Response.FAVORITE_ERROR;
	}
}
