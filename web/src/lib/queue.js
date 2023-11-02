// This file contains every function used in the queue.svelte component
import { Response } from "./error";

export async function AddToQueueHTML(GuildID, token, host) {
    // Values needed for adding a song to a queue
    let song = document.getElementById("song").value.trim();
    let shuffle = document.getElementById("shuffle")?.checked;
    let playlist = document.getElementById("playlist")?.checked;
    let loop = document.getElementById("loop")?.checked;
    let priority = document.getElementById("priority")?.checked

    return AddToQueue(GuildID, token, host, song, shuffle, playlist, loop, priority);
}

// Function to add a song to the queue
export async function AddToQueue(GuildID, token, host, song, shuffle, playlist, loop, priority) {
    // Request
    let route = `${host}/queue/${GuildID}`
    let response = await fetch(route, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'token': token,
            'song': song,
            'shuffle': shuffle.toString(),
            'playlist': playlist.toString(),
            'loop': loop.toString(),
            'priority': priority.toString(),
        }).toString(),
    })

    // Error Handling
    switch(response.status) {
        case 200:
            return Response.SUCCESS;
        case 401:
            return Response.QUEUE_TOKEN_ERR;
        case 403:
            return Response.QUEUE_ADD_ERR;
        case 406:
            return Response.QUEUE_PLAYLIST_ERR;
    }
}

// Function to remove a song from the queue or clear it
export async function RemoveFromQueue(GuildID, token, clear=false, host) { // AKA skip
    // Request
    let route = `${host}/queue/${GuildID}?` + new URLSearchParams({'clean': clear.toString(), 'token': token}).toString();
    let response = await fetch(route, {
        method: "DELETE",
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded'
        },
    })

    // Error Handling
    switch(response.status) {
        case 200:
            return Response.SUCCESS;
        case 401:
            return Response.QUEUE_TOKEN_ERR;
        case 403:
            return Response.QUEUE_CHANNEL_ERR;
        case 406:
            return Response.QUEUE_PLAY_ERR;
    }
}

// Function to get the queue
export async function GetQueue(GuildID, token, host) {
    // Request
    let route = `${host}/queue/${GuildID}?` + new URLSearchParams({"token": token}).toString();
    let response = await fetch(route, {
        method: "GET",
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded',
        },
    })

    // Error Handling
    switch(response.status) {
        case 200:
            return await response.json();
        case 401:
            return Response.QUEUE_TOKEN_ERR;
    }
}