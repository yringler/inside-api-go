# inside-api
API to get data about inside chasidus classes

# Plan

For now, really simple. 

Scrapes inside chasidus and uploads dart-ready huge JSON file (each line a seperate JSON object, for easy streaming) (gzipped) if it wasn't already done.

If the data is on dropbox already, return a temporary 3 hour link to the file.

Eventually this will support incremental updates (with time stamps or something), but that effort isn't justifed for an MVP.
