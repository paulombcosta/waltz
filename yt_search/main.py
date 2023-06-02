from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional
import ytmusicapi
import json

yt = ytmusicapi.YTMusic()

app = FastAPI()

auth_client : Optional[ytmusicapi.YTMusic] = None

class AuthPayload(BaseModel):
    access_token : str
    refresh_token : str
    token_type : str
    expires_in : float
    expires_at : int

class CreatePlaylistPayload(BaseModel):
    title : str
    auth : AuthPayload

class CreatePlaylistItemPayload(BaseModel):
    playlist_id : str
    track_id : str
    auth : AuthPayload

def get_auth_client(auth: AuthPayload) -> ytmusicapi.YTMusic:
    if (auth_client is None):
        auth_client = ytmusicapi.YTMusic(json.dumps(auth))
    return auth_client


@app.get("/track/search")
async def root(q: str = ""):
    if q == "":
        raise HTTPException(status_code=400, detail="query cannot be empty")
    res = yt.search(q)
    if len(res) == 0 or "videoId" not in res[0]:
        raise HTTPException(status_code=404)
    return {'video_id': res[0]["videoId"]}

@app.get("/playlists/{channelId}")
async def get_playlists(channelId : str):

    if channelId == "":
        raise HTTPException(status_code=400, detail="channelId cannot be empty")

    u = yt.get_user(channelId=channelId)

    playlists = yt.get_user_playlists(channelId=channelId, params=u["playlists"]["params"])
    return playlists

@app.get("/playlist/{playlist_id}")
async def get_playlist(playlist_id : str):
    if playlist_id == "":
        raise HTTPException(status_code=400, detail="playlist_id cannot be empty")
    return yt.get_playlist(playlistId=playlist_id)

@app.post("/playlist")
async def insert_playlist(payload : CreatePlaylistPayload):
    id = get_auth_client(payload.auth).create_playlist(title=payload.title, description="created by waltz", privacy_status="PUBLIC")
    return {'id': id }

@app.post("/track")
async def insert_playlist_item(payload : CreatePlaylistItemPayload):
    status = get_auth_client(payload.auth).add_playlist_items(playlistId=payload.playlist_id, videoIds=[payload.track_id], source_playlist=None, duplicates=False)
    return {'status': status}
