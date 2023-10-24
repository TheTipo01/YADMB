export async function AddFavorite() {
    let name = document.getElementById("name")?.value;
    let link = documnent.getElementById("link")?.value;
    let folder = document.getElementById("folder")?.value;
    let token = document.getElementById("token")?.token;

    // Request
    let route = `https://gerry.thetipo.rocks/favorites`
    let response = await fetch(route, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'name': name,
            'link': link,
            'folder': folder,
            'token': token,
        }).toString(),
    })

    //Error Handling
    switch(response.status) {
        case 200:
            return 0;
        case 401:
            return -1;
        case 500:
            return -5;
    }
}

export async function RemoveFavorite() {
    let name = document.getElementById("name")?.value;
    let token = document.getElementById("token")?.value;

    // Request
    let route = `https://gerry.thetipo.rocks/favorites`
    let response = await fetch(route, {
        method: 'DELETE',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'name': name,
            'token': token,
        }).toString(),
    })

    //Error Handling
    switch(response.status) {
        case 200:
            return 0;
        case 401:
            return -1;
        case 500:
            return -5;
    }
}

export async function GetFavorites() {
    let token = document.getElementById("token")?.value;

    // Request
    let route = `https://gerry.thetipo.rocks/favorites`
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