import { AddFavorite } from "./favorites";
import { AddToQueue } from "./queue";

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