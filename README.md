# inside-api
API to ensure app has latest version of data

# Plan

There will one endpoint to get data. App will pass in a date parameter. If that date is earlier than current data date, redirects to dropbox json file.

If it's after, returns NOCONTENT

Another endpoint will allow to update current data date. It will take a singular query paramter with the current data date.

Actually, I might just put both at one endpoint
