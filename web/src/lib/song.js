export async function ToggleSong(action = "") {
    let GuildID = document.getElementById("id").value;
    let token = document.getElementById("token").value;

    // Request
    if(route === "") {
        return -8
    }
    else {
        let route = `https://gerry.thetipo.rocks/song/${action}/${GuildID}`;
        let response = await fetch(route, {
            method: "GET",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: {
                'token': token,
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
