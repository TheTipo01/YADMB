import { AddFavorite } from "./favorites";
import { AddToQueue } from "./queue";

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

export function GetHost() {
    return window.location.protocol + "//" + window.location.host;
}

export function KeyPressed(e, scope, GuildId = "", token = "", host = "") {
    switch(e.key) {
        case 'Enter':
            switch(scope) {
                case 'queue':
                    AddToQueue(GuildId, token, host);
                case 'favorite':
                    AddFavorite(token, host);
            }
            break;
    }
}