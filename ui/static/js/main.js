let socket = undefined;

function setup() {
    el = document.getElementById("transfer")
    if (el !== undefined) {
        document.getElementById("transfer").onclick = () => {
            disableTransferButton();
            startTransfer();
        }
    }
}

function disableTransferButton() {
    const el = document.getElementById("transfer")
    el.classList.add("disabled")
}

function stopSocket() {
    if (socket !== undefined) {
        socket.close();
    }
}

function startTransfer() {
    socket = new WebSocket("ws://localhost:8080/transferSocket")
    socket.addEventListener('open', (event) => {
        socket.send(JSON.stringify({"command": "START"}));
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
            stopSocket()
            break;
        case "error":
            updateProgressEndText(`error: ${msg.body}`)
            break;
        default:
            updateProgressEndText(`invalid message received from server: ${msg.type}`)
            break;
    }
}

function increaseTrackProgress() {
    const el = document.getElementById("trackProgressValue");
    const currentCount = parseInt(el.innerText);
    el.innerText = `${currentCount + 1}`;
}

function increasePlaylistProgress() {
    const el = document.getElementById("playlistProgressValue");
    const currentCount = parseInt(el.innerText);
    el.innerText = `${currentCount + 1}`
}

function updatePlaylistName(name) {
    document.getElementById("playlistNameValue").innerText = name;
}

function updateProgressEndText(text) {
    const el = document.getElementById("progressEndText")
    el.innerText = text
    el.classList.remove("disabled")
    el.classList.add("enabled")
    stopSocket();
}

window.onload = setup
