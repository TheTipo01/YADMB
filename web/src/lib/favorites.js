import { Response } from "./error";

export async function AddFavorite(token) {
    let name = document.getElementById("name").value;
    let link = document.getElementById("link").value;
    let folder = document.getElementById("folder")?.value;

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
            return Response.SUCCESS;
        case 401:
            return Response.FAVORITE_TOKEN_ERROR;
        case 500:
            return Response.DUPLICATE_ERROR;
    }
}

export async function RemoveFavorite(token, name) {
    // Request
    let route = `https://gerry.thetipo.rocks/favorites?` + new URLSearchParams({'name': name, 'token': token}).toString();
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
            return Response.SUCCESS;
        case 401:
            return Response.FAVORITE_TOKEN_ERROR;
        case 500:
            return Response.FAVORITE_ERROR;
    }
}

export async function GetFavorites(token) {
    // Request
    let route = `https://gerry.thetipo.rocks/favorites?` + new URLSearchParams({"token": token}).toString();
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
            return await response.json();
        case 401:
            return Response.FAVORITE_TOKEN_ERROR;
    }
}