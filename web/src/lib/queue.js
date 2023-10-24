// This file contains every function used in the queue.svelte component

export async function AddToQueue() {
    // Values needed for adding a song to a queue
    let GuildID = document.getElementById("id")?.value;
    let song = document.getElementById("song")?.value;
    let token = document.getElementById("token")?.value;
    let shuffle;
    let playlist = document.getElementById("playlist").value;
    let loop = document.getElementById("loop").value;
    let priority = document.getElementById("priority").value
    if(playlist) {
        shuffle = document.getElementById("shuffle").value;
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