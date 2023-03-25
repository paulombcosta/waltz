# Waltz
 
 Transfer your playlists from Spotify to Youtube Music. Self hosted.

## Usage

TODO instructions on how to create project.

## Limitations

Youtube gives a daily quota of 10.000 with each API call having a different cost. Currently for
each playlist the operations costs are:

| Operation            | Intent                                            | Cost                  |
|----------------------|---------------------------------------------------|-----------------------|
| list playlists       | Find if playlist already exists                   | 1                     |
| insert playlist      | Create playlist if it doesn't exist               | 50                    |
| list playlist items  | Get existing tracks to not insert repeated tracks | 1 for every 50 tracks |
| search               | find videoId by name. Necessary to insert track   | 100                   |
| insert playlist item | Creates the track on the playlist                 | 50                    |

Which is fairly limited to around 66 tracks daily. Even if the read data comes from another source, like
a scrapper, the number would improve to only 200 at best.