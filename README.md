# tmp-place

> [!WARNING]
> This software is in very early stages. Read the code and contribute if you like, but this is definitely not ready in any shape or form to be deployed.

The idea is to have a small project written in one language that can be used for upload/download files that live for short-term. Something generally akin to [0x0.st](https://git.0x0.st/mia/0x0) but with some modifications.

I started this project as a way to learn Go, and also to scratch a personal itch. The 0x0.st is a great project, but has some parts that imho can be improved such as:
1. Go is superior compared to python in terms of performance and also because it produces one single binary
2. User should be able to specify the TTL (time to live) if it is shorter than what the admin has set
3. User should be able to also specify a password for the file when it is getting downloaded

--------------------------------------------------------------------------------

## Workflow
- upload
	- user use curl to upload the file and can define secret and TTL (Time To Live) in minutes:
		- To upload the `yourfile.png` to  `curl -F'file=@yourfile.png' -F 'ttl=15' -F 'secret=VerySecurePassword' -F 'oneoff=true' https://domain.tld`
	- if the upload was successful and some initial file-checks (e.g., virus scan, etc.) was passed, generate the URL and return it to user
- downoload
	- if the URL is valid, let the user to download the file

## Features
- Users should be able to upload using curl (similar to 0x0.st)
- There should be a simple text-based homepage is user visits the root of the domain. text-based because the user should be able to read if in their browser or terminal.
- There is no admin config web panel, instead there is a json file that they can configure
- Admin should be able to config:
	- white-list and black-list IP and IP ranges
	- The TTL for each uploaded file
	- The number of downloads per file
	- maximum file size
	- number of uploads per IP per second/minute
	- max upload size (default 10)
- the user should be able to configure through curl command arguments:
	- `-F 'ttl=10m'`: the time to live. I.e., how long the file should stay on the server before getting cleaned. This cannot be longer than the TTL admin has defined.
	- `-F 'secret'`: The password that should be provided in the download command, either by curl or by adding `?secret=`.
	- `-Foneoff`: This is boolean, indicating that the file is wiped after the first download

--------------------------------------------------------------------------------

## Usage

### For uploading
Assuming that this software is deployed on `domain.tld`, the user can upload the file by:

```bash
curl -F 'file=@/path/to/file' 'http://domain.tld'
```

user can specify other things like how long this file is available for download using `ttl`. 

```bash
curl -F 'file=@/path/to/file' -F 'ttl=10m' 'http://domain.tld'
```

The TTL can be any of these formats:
- `42` for 42 seconds
- `10m` or `10 minutes` or `10:00`
- `1h` or `1 hour` or `1:00:00`
- `2d` or `2 days`
The TTL cannot exceed the maximum TTL that is allowed by the server. User will get an error if their TTL is larger than server's max TTL.

Also user can add a password for when the file is getting downloaded:

```bash
curl -F 'file=@/path/to/file' -F 'secret=VeryStrongPassword' 'http://domain.tld'
```

The `secret` can only contain `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.` characters with the max length of 42 characters. Any characters that is not listed here will be removed, and any extra characters after sanitization will be truncated to max of 42 characters.

Apart from the `file` argument, the rest are optional can be used in combination. For example:

```bash
curl -F 'file=@/path/to/file' -F 'ttl=10m' -F 'secret=VeryStrongPassword' http://domain.tld
```

upon uploading, user will recieve a message with the information they need. For example:

```console
❯ curl -F 'file=@/path/to/file' -F 'secret=12345' 'http://domain.tld'
File pacman-error uploaded successfully.
TTL set to 36000 seconds.
Download secret is: 12345
Use the secret to download your file securely.
Download link: localhost/MlctfT?secret=12345
```

### For downloading
Use `curl` or `wget` to download the file using the URL. The password secret can be passed in the following ways:

```bash
curl -F 'secret=VeryStringPassword' 'http://domain.tld/n2eu0B'
# or
curl 'http://domain.tld/n2eu0B?secret=VeryStringPassword'
# or
wget 'http://domain.tld/n2eu0B?secret=VeryStringPassword'
```

--------------------------------------------------------------------------------
## Installation
This project comes with setup wizard that guide you step-by-step to have get a `config.json` and database.
You can [get the binary](#get-the-binary) or you can [build this project from source](#buiding-from-source) to get the binary file.

When you got the binary file and the schema template file, just run the setup:

```sh
./tmp-place setup
```

### Get the binary
You need two files that you can download:
- [The binary file](https://github.com/mmahmoudian/tmp-place/releases/latest)
- [The database schema template](https://raw.githubusercontent.com/mmahmoudian/tmp-place/refs/heads/main/db_schema.sql)

### Buiding from source
```sh
# get the code
git clone --depth 1 https://github.com/mmahmoudian/tmp-place.git

# make sure you have Go installed. You can check by:
go version

# build the binary
make build

# move the file to where you want
mv bin/tmp-place /the/path/of/your/choice/

# copy the database schema
cp db_schema.sql /the/path/of/your/choice/
```

--------------------------------------------------------------------------------

## Project structure
This project will produce one binary file with four subcommands for each of the modules of this software:
1. `setup` to provide a step-by-step and guided wizard to setup the project on server
2. `server` to start the webserver
3. `admin` to provide administrative functionality to the admin
4. `janitor` which clean up the files based on TTL and other factors, and updates the database

The main file is `main.go` in the root of the project which decides what module should be called based on the subcommand. For each of the subcommands there is a dedicated folder under `cmd/` that contains the functions used by that module. There is a `main.go` file in each of the module-specific folder (i.e., `cmd/*/main.go`) that contains the main functions of that module. Treat them as entry points of each module. There is also an additional `cmd/shared/` which contains functions that are used by more than one module, such as reading config and etc. Every file in `cmd/` is named after the general functionality of the functions inside that file (e.g., `file.go` for handling files, `time.go` for handling time conversion and etc).

For every file under the `cmd/`, there is a file with same name but suffixed with `_test.go` which contains the test units for functions in their respective files.

--------------------------------------------------------------------------------

## ToDo
- [x] change to single app with subcommands
- [ ] implement proper logging
- server
	- [x] implement file upload from the user
	- [x] generate 6 character file tag
	- [x] allow the user to specify the secret that should be provided during the download
	- [ ] store the data in a sqlite database
	- [ ] implement file download
	- [ ] implement using the secret for file download
	- [ ] delete the file if it had the one-off flag
	- [ ] virus scan (should be configurable in the config)
- janitor
	- [ ] clean the files based on the TTL
	- [ ] log rotation
- admin
	- [ ] general statistics for the admin
		- [ ] active TTLs
		- [ ] active one-offs
		- [ ] total file size
		- [ ] total uploads in last minute, hour, day, week, month, year
	- [ ] lock-down mode (temporarily reject any uploads with a message)
	- [ ] IP blocking
	- [ ] IP rate limit

