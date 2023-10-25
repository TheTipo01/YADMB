// This file contains every function used in the queue.svelte component

export async function AddToQueue(GuildID, token) {
    // Values needed for adding a song to a queue
    let song = document.getElementById("song").value;
    let shuffle;
    let playlist = document.getElementById("playlist")?.checked;
    let loop = document.getElementById("loop")?.checked;
    let priority = document.getElementById("priority")?.checked
    if(playlist) {
        shuffle = document.getElementById("shuffle")?.checked;
    }
    // Request
    let route = `https://gerry.thetipo.rocks/queue/${GuildID}`
    let response = await fetch(route, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'token': token,
            'song': song,
            'shuffle': shuffle,
            'playlist': playlist,
            'loop': loop,
            'priority': priority
        }).toString(),
    })

    //Error Handling
    switch(response.status) {
        case 200:
            return 0;
        case 401:
            return -1;
        case 403:
            return -2;
        case 406:
            return -3;
    }
}

export async function RemoveFromQueue(GuildID, token, clear=false) { // AKA skip
    // Request
    let route = `https://gerry.thetipo.rocks/queue/${GuildID}` + new URLSearchParams({'clean': clear, 'token': token}).toString();
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
            return 0;
        case 401:
            return -1;
        case 403:
            return -2;
        case 406:
            return -3;
    }
}

export async function GetQueue(GuildID, token) {
    // Request
    let route = `https://gerry.thetipo.rocks/queue/${GuildID}?` + new URLSearchParams({"token": token}).toString()
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
            return -1;
    }

}