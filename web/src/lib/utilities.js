// Function to add an object to the array
export function AddObjectToArray(promise, objectToAdd) {
	return promise.then((array) => {
		if (!Array.isArray(array)) {
			return [...objectToAdd];
		}

		return [...array, ...objectToAdd];
	});
}

// Function to add an object to the array as the second element
export function AddObjectToArrayAsSecond(promise, objectToAdd) {
	return promise.then((array) => {
		if (!array || array.length === 0) {
			return [...objectToAdd];
		}

		return [array[0], ...objectToAdd, ...array.slice(1)];
	});
}

// Function to remove the first object from the array
export function RemoveFirstObjectFromArray(promise) {
	return promise.then((array) => {
		if (!Array.isArray(array)) {
			return [];
		}

		return array.slice(1);
	});
}

// Function to remove the given object from the array
export function RemoveObjectFromArray(promise, objectToRemove) {
	return promise.then((array) => {
		if (!Array.isArray(array)) {
			return [];
		}

		return array.filter((object) => {
			return object !== objectToRemove;
		});
	});
}

// Function to remove favorites with the given name from the array
export function RemoveObjectFromArrayByName(promise, name) {
	return promise.then((array) => {
		if (!Array.isArray(array)) {
			return [];
		}

		return array.filter((object) => object.name !== name);
	});
}

// Function to negate isPaused of the first element in the queue
export function TogglePause(promise) {
	return promise.then((array) => {
		if (Array.isArray(array) && array.length > 0) {
			array[0].isPaused = !array[0].isPaused;
		}
		return array;
	});
}

// Function to set isPaused of the first element in the queue
export function SetPause(promise, isPaused) {
	return promise.then((array) => {
		if (Array.isArray(array) && array.length > 0) {
			array[0].isPaused = isPaused;
		}

		return array;
	});
}

// Function used when clearing the queue
export function ClearArray(promise) {
	return promise.then(() => {
		return [];
	});
}

// Function to get the frames
export function GetFrames(promise) {
	return promise.then((array) => {
		if (array.length > 0) {
			return array[0].frames;
		}
		return 0;
	});
}

// Function to get the isPaused status from the promise
export function GetPauseStatus(promise) {
	return promise.then((array) => {
		if (array.length > 0) {
			return !array[0].isPaused;
		}
		return false;
	});
}

// Function to get the GuildID from URL query
export function GetGuildID() {
	let query = new URLSearchParams(window.location.search);
	return query.get('GuildId') ?? '';
}

// Function to get the Token from URL query
export function GetToken() {
	let query = new URLSearchParams(window.location.search);
	return query.get('token') ?? '';
}

// Function to get the hostname and protocol from URL
export function GetHost() {
	return window.location.protocol + '//' + window.location.host;
}

// Function to get the number of folders
export function GetFolders(favorites) {
	let folders = [''];
	if (!Array.isArray(favorites)) {
		return folders;
	}

	for (let i = 0; i < favorites.length; i++) {
		if (!folders.includes(favorites[i].folder)) {
			folders.push(favorites[i].folder);
		}
	}

	return folders;
}

// Function to get timestamp
export function GetTime(seconds) {
	seconds = Math.max(0, Math.floor(Number(seconds) || 0));

	const hours = Math.floor(seconds / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	const secs = seconds % 60;
	const pad = (value) => String(value).padStart(2, '0');

	if (hours !== 0) {
		return `${pad(hours)}:${pad(minutes)}:${pad(secs)}`;
	}
	if (minutes !== 0) {
		return `${pad(minutes)}:${pad(secs)}`;
	}

	return `00:${pad(secs)}`;
}

// Only allow http(s) links in href attributes.
export function SafeHref(link) {
	if (!link) {
		return '#';
	}

	try {
		const base = typeof window !== 'undefined' ? window.location.origin : 'http://localhost';
		const url = new URL(link, base);
		if (url.protocol === 'http:' || url.protocol === 'https:') {
			return link;
		}
	} catch {
		// Ignore invalid URLs.
	}

	return '#';
}
