# Command line client for Amazon Glacier.

## Basic usage

Run `glacier -help` for a list of all supported commands.

### Keys

Keys can be provided on the command line (-secret=xxx -access=xxx), in a file (-keys=keyfile which has the secret key before the access id key), or by environment variables (AWS_SECRET_KEY, AWS_ACCESS_KEY).

### Uploading

Uploading an archive, returns the archive id:
  `glacier archive upload us-east-1 testvault file.zip "Description of file.zip"`

Upload a large archive in 64 MiB parts, returns the archive id:
  `glacier multipart run us-east-1 testvault file.zip 64 "Description of file.zip"`

### Downloading

Initiate a download job, returns job id, and retrieve it once it's complete:
  `glacier job archive us-east-1 testvault archiveid snstopic "Descripiton of retrieval job"`
  `glacier job get archive us-east-1 testvault jobid file.zip`

Downloading a large file in 64 MiB parts by initiating the retrieval job, waiting and periodically poll to see if the retrieval is complete, and then download the file:
  `glacier job run us-east-1 testvault archiveid 64 file.zip`
