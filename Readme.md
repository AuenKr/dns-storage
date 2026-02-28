# DNS Storage

## Idea

Idea was to store files binary into DNS TXT records.
And using that we can upload and download a files.

The download and upload will be slow. But video streaming can work this way. This was my initial idea.

But, I hit a cloudflare dns registrar limit.
I am using cloudflare as my domain registrar which limit me to keep 200 records at max.
But there are some provider which provide unlimited dns records in their paid plans.
Example:

1. [EasyDNS Standard Tier](https://easydns.com/dns/)
2. [ClouDNS Enterprise](https://www.cloudns.net/premium/)
3. [BunnyDNS](https://bunny.net/dns/) - Cannot find the limit

## Working

When you upload a file, it upload those file binary(base64 format) into DNS TXT records in chunks.
And after uploading a index file is created, that contains the no of chunks data.

Using that index file, we can download the file or stream that file.

## How to use the project

1. Copy sample.env to .env and fill the values

2. create a .temp/download folder

3. For uploading a file change the FilePath variable, run `go run cmd/upload/main.go`

4. For downloading a file, run `go run cmd/download/main.go`

5. For deleting a file, run `go run cmd/delete/main.go`

## Current Limits

As cloudflare limit the number of records in a zone to 200.

I can only create 200 TXT records each of 255 bytes.
Normally storing bytes directly causing some issue, so i decided to store them in base64 format which further reduce the size.
To around 191 bytes per TXT records.

In cloudflare, I can max store ~37Kb data only.
