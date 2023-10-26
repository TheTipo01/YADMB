// This file contains a function used in the queue.svelte component

export async function ToggleSong(GuildID, token, action = "", host) {  // AKA Pause/Resume Song
    // Request
    if(action === "") {
        return -8
    }
    else {
        let route = `https://${host}/${action}/${GuildID}?` + new URLSearchParams({"token": token}).toString();
        let response = await fetch(route, {
            method: "GET",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/x-www-form-urlencoded'
            },
        })
    
        // Error Handling
        switch(response.status) {
            case 200:
                return 0;
            case 401:
                return -1;
        }
    }
}
