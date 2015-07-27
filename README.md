# Patchy

**Work in progress, do not use.**

An online jukebox by [Luminarys](https://github.com/Luminarys), 
[minus](https://github.com/minus7), and [SirCmpwn](https://github.com/SirCmpwn).

![](https://sr.ht/6d07.png)

Usage instructions to come once this software is more mature.

#Things Done:
* Load in music library
* Setup websockets
* Update Now Playing using websocket
* Implement the music timer properly 
* Implement proper music sync -- Use a dual jPlayer setup where one player loads the next song while the other plays using an initial mpd stream. N.B. Something goes wrong in Firefox that results in a cached song being played instead of the next one.


#Things to Do:
* Implement the Queue properly
* Implement client requests
* Implement client music uploads
