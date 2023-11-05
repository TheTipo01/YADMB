import { Response } from "./error";

// This file contains a function used in the queue.svelte component

// Function to pause or resume the current song 
export async function ToggleSong(GuildID, token, action = "", host) {
    // Request
    if (action === "") {
        return -8
    } else {
        let route = `${host}/song/${action}/${GuildID}?` + new URLSearchParams({"token": token}).toString();
        let response = await fetch(route, {
            method: "GET",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/x-www-form-urlencoded'
            },
        })

        // Error Handling
        switch (response.status) {
            case 401:
                return Response.QUEUE_TOKEN_ERR;
                break;
            case 406:
                return Response.SONG_PAUSED_ERR;
                break;
            case 500:
                return Response.SONG_TOGGLE_ERR;
                break;
        }
    }
}
