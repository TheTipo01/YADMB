import { onMount } from "svelte";

// Function to add an object to the array 
export function addObjectToArray(promise, objectToAdd) {
    return promise.then((array) => {
        return [...array, ...objectToAdd];
        });
    }

// Function to remove the first object from the array 
export function RemoveObjectFromArray(promise) {
return promise.then((array) => {
        array.shift();
        return array;
    });
}

export function ClearArray(promise) {
    return promise.then(() => {
        return [];
    })
}

export function GetGuildID() {
    let query = new URLSearchParams(window.location.search);
    return query.get("GuildId");
}

export function GetToken() {
    let query = new URLSearchParams(window.location.search);
    return query.get('token');
}
