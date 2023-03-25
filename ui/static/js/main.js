function getSelectedPlaylists() {
    return $("#table input[type=checkbox]:checked").map(function() {
        const table = document.getElementById(this.id)
            .parentElement
            .parentElement
        const name = table
            .getElementsByClassName("name")[0].textContent;
        const totalTracks = table
            .getElementsByClassName("totalTracks")[0].textContent;
        return {id: this.id, name: name, totalTracks: totalTracks};
    }).get();
}

function toggleSelectAll(checked) {
    if (checked) {
        $('#table input[type=checkbox]').prop('checked', true)
    } else {
        $('#table input[type=checkbox]').prop('checked', false)
    }
    toggleSubmitButton();
}

function toggleSubmitButton() {
    if ($("#table input[type=checkbox]:checked").length == 0) {
        document.getElementById("submit").classList.add("disabled")
    } else {
        document.getElementById("submit").classList.remove("disabled")
    }
}

function setup() {
    $(".checkbox").change(function() {
        toggleSubmitButton();
    })
    document.getElementById("submit").onclick = () => {
        const playlists = getSelectedPlaylists()
        setupProgress(playlists);
        startTransfer(playlists)
    }
    document.getElementById("bulk").onchange = (event) => {
        toggleSelectAll(event.target.checked)
    }
}

function startTransfer(playlists) {
    const socket = new WebSocket("ws://localhost:8080/transfer")
    const payload = playlists.map(x => {
        return {"id": x.id, "name": x.name}
    })
    socket.addEventListener('open', (event) => {
        socket.send(JSON.stringify({"playlists": payload}));
    });
    
    socket.addEventListener('message', (event) => {
        data = JSON.parse(event.data);
        handleMessage(data);
    });
}

function handleMessage(msg) {
    switch (msg.type) {
        case "playlist-start":
            updatePlaylistName(msg.body)
            break;
        case "track-done":
            increaseTrackProgress()
            break;
        case "playlist-done":
            increasePlaylistProgress()
            break;
        case "done":
            updateProgressEndText("Finished")
            break;
        default:
            updateProgressEndText(`invalid message received from`)
            break;
    }
}

function increaseTrackProgress() {
    const el = document.getElementById("trackProgressCount");
    const currentCount = parseInt(el.innerText.split(" ")[2]);
    el.innerText = `Tracks Transferred: ${currentCount + 1} of ${window.totalTracks}`;
}

function increasePlaylistProgress() {
    const el = document.getElementById("playlistProgressCount");
    const currentCount = parseInt(el.innerText.split(" ")[2]);
    el.innerText = `Playlists Transferred: ${currentCount + 1} of ${window.playlistsTotal}`
}

function updatePlaylistName(name) {
    document.getElementById("currentPlaylist").innerText = `"Transfering Playlist: ${name}"`;
}

function getTotalOfTracks(playlists) {
    return playlists.map(p => parseInt(p.totalTracks));
}

function updateProgressEndText(text) {
    const el = document.getElementById("progressEndText")
    el.innerText = text
    el.classList.remove("disabled")
    el.classList.add("enabled")
}

function setupProgress(playlists) {
    progressContainer = document.createElement("div");
    progressContainer.classList.add("progressContainer")

    title = document.createElement("p");
    title.classList.add("progressTitle")
    title.textContent = "Transfer in Progress"

    currentPlaylist = document.createElement("p");
    currentPlaylist.classList.add("currentPlaylist");
    currentPlaylist.id = "currentPlaylist"
    currentPlaylist.textContent = "Transfering Playlist: -";

    playlistProgressCount = document.createElement("p")
    playlistProgressCount.classList.add("playlistProgressCount")
    playlistProgressCount.id = "playlistProgressCount"
    window.playlistsTotal = playlists.length
    playlistProgressCount.textContent = `Playlists Transferred: 0 of ${playlists.length}`

    trackProgressCount = document.createElement("p")
    trackProgressCount.classList.add("trackProgressCount")
    trackProgressCount.id = "trackProgressCount"
    const totalTracks = getTotalOfTracks(playlists);
    window.totalTracks = totalTracks;
    trackProgressCount.textContent = `Tracks Transferred: 0 of ${totalTracks}`

    progressEndText = document.createElement("p");
    progressEndText.classList.add("progressEndText");
    progressEndText.classList.add("disabled");
    progressEndText.id = "progressEndText";

    progressContainer.appendChild(title);
    progressContainer.appendChild(currentPlaylist);
    progressContainer.appendChild(playlistProgressCount);
    progressContainer.appendChild(trackProgressCount);
    progressContainer.appendChild(progressEndText);

    document.getElementsByTagName("body")[0].replaceChildren(progressContainer)
}

window.onload = setup
