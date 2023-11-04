// Function to add an object to the array
export function AddObjectToArray(promise, objectToAdd) {
    return promise.then((array) => {
        return [...array, ...objectToAdd];
    });
}

// Function to remove the first object from the array 
export function RemoveFirstObjectFromArray(promise) {
    return promise.then((array) => {
        array.shift();
        return array;
    });
}

// Function to remove the given object from the array
export function RemoveObjectFromArray(promise, objectToRemove) {
    return promise.then((array) => {
        return array.filter((object) => {
            return object !== objectToRemove;
        });
    });
}

// Function to negate isPaused of the first element in the queue
export function TogglePause(promise) {
    return promise.then((array) => {
        if (array.length > 0) {
            array[0].isPaused = !array[0].isPaused;
        }
        return array;
    });
}

// Function to set isPaused of the first element in the queue
export function SetPause(promise, isPaused) {
    return promise.then((array) => {
        if (array.length > 0) {
            array[0].isPaused = isPaused;
        }

        return array;
    });
}

// Function used when clearing the queue
export function ClearArray(promise) {
    return promise.then(() => {
        return [];
    })
}

// Function to get the GuildID from URL query
export function GetGuildID() {
    let query = new URLSearchParams(window.location.search);
    return query.get("GuildId");
}

// Function to get the Token from URL query
export function GetToken() {
    let query = new URLSearchParams(window.location.search);
    return query.get('token');
}

// Function to get the hostname and protocol from URL
export function GetHost() {
    return window.location.protocol + "//" + window.location.host;
}

// Function to get the number of folders
export function GetFolders(favorites) {
    let nFolders = 1;
    let folders = [""];
    for(let i = 0; i < favorites.length; i++) {
        if(!folders.includes(favorites[i].folder)) {
            nFolders++;
            folders.push(favorites[i].folder);
        }
    }

    return folders;
}