# DNS Storage

## Idea

Idea was to store files binary into DNS TXT records.
And using that we can upload and download a files.

The download and upload will be slow. But video streaming can work this way. This was my initial idea.
Most of the provider have limit on number of creation of records in a zone.

BunnyDNS has limit has no limit on number of records in a zone. So i used that.

1. [BunnyDNS](https://bunny.net/dns/) - No limit in free tier.
2. [EasyDNS Standard Tier](https://easydns.com/dns/)
3. [ClouDNS Enterprise](https://www.cloudns.net/premium/)

## Working

When you upload a file, it upload those file binary(base64 format) into DNS TXT records in chunks.
And after uploading a index file is created, that contains the no of chunks data.

Using that index file, we can download the file or stream that file.

## How to use the project

1. Copy sample.env to .env and fill the values

2. create a .temp/download folder

3. For uploading a file change the FilePath variable, run `go run cmd/upload/main.go  --path <file_path>`

4. For downloading a file, run `go run cmd/download/main.go --index <index.domain>`

5. For deleting a file, run `go run cmd/delete/main.go --index <index.domain>`

## Current Limits

~As I am using cloudflare limit the number of records in a zone to 200.~

I can only create 200 TXT records each of 255 bytes.
Normally storing bytes directly causing some issue, so i decided to store them in base64 format which further reduce the size.
To around 191 bytes per TXT records.

In cloudflare, I can max store ~37Kb data only.

> I switched to bunnydns as it has not limit define, and it is working fine.

You can download this file for testing: `13caaf91-cc8e-4897-9f95-4bfbca56dc3e.auenkr.qzz.io`
It an mp3 file
