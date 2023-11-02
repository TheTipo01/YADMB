export const Response = Object.freeze({
    SUCCESS: 0,
    QUEUE_TOKEN_ERR: 1, // Token or Guild not valid
    QUEUE_ADD_ERR: 2,  // Code 403: not in the channel or dj mode
    QUEUE_PLAYLIST_ERR: 3, // playlist flag not raised
    QUEUE_EMPTY_ERR: 4,
    QUEUE_CHANNEL_ERR: 5, // user not in channel 
    QUEUE_PLAY_ERR: 6, // playback error
    SONG_PAUSED_ERR: 7,  // Song already paused
    SONG_RESUME_ERR: 8, // Song already playing
    SONG_TOGGLE_ERR: 9, // Can't toggle song
    DUPLICATE_ERROR: 10,  // Duplicate name
    FAVORITE_ERROR: 11, // Favorite doesn't exist
    FAVORITE_TOKEN_ERROR: 12,
})

// Function to handle the status code of each function
export function get_message(code) {
    let msg = ""
    switch(code) {
        case Response.SUCCESS:
            msg = "Done.";
            break;
        case Response.QUEUE_ADD_ERR:
            msg = "You are not in a channel or dj mode is enabled and you're not a dj";
            break;
        case Response.QUEUE_CHANNEL_ERR:
            msg = "You are not in a channel";
            break;
        case Response.QUEUE_TOKEN_ERR:
            msg = "Token or Guild ID not valid";
            break;
        case Response.QUEUE_EMPTY_ERR:
            msg = "Queue is empty";
            break;
        case Response.QUEUE_PLAY_ERR:
            msg = "Playback error";
            break;
        case Response.QUEUE_PLAYLIST_ERR:
            msg = "Playlist flag not raised";
            break;
        case Response.SONG_PAUSED_ERR:
            msg = "Song is already paused";
            break;
        case Response.SONG_RESUME_ERR:
            msg = "Song is already playing";
            break;
        case Response.SONG_TOGGLE_ERR:
            msg = "Can't toggle song";
            break;
        case Response.DUPLICATE_ERROR:
            msg = "Name is already present";
            break;
        case Response.FAVORITE_ERROR:
            msg = "Favorite does not exist";
            break;
        case Response.FAVORITE_TOKEN_ERROR:
            msg = "User token is invalid";
            break;
        default:
            msg = "Unknown error";
    }

    return msg;
}