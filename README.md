# srrdb Terminal Client

A terminal client to access [srrdb.com](https://www.srrdb.com), written in Go.

## Install

Arch Linux users can use the package in the `extra` repository [srrdb-terminal-client](https://archlinux.org/packages/extra/x86_64/srrdb-terminal-client/).

Other users should simply download the latest release for their arch from [here](https://github.com/hashworks/srrdb-Terminal-Client/releases/latest) and move the executable to their `$PATH`.

## Usage

See `--help`:
```
-v, --version
	Shows the version and general information.
	Options:
	-j, --json
		Output in JSON format.

-s, --search <query>[...]
	Searches srrdb.com for releases.
	For a list of available keywords see https://www.srrdb.com/help#keywords
	Options:
	-j, --json
		Output in JSON format.

-i, --imdb <releasename>[...]
	Shows IMDB information for one or multiple releases.
	Options:
	-j, --json
		Output in JSON format.

-N, --nfo <releasename>[...]
	Shows NFO files with download URLs for one or multiple releases.
	Options:
	-j, --json
		Output in JSON format.

-D, --details <releasename>[...]
	Shows detailed information for one or multiple releases.
	Options:
	-j, --json
		Output in JSON format.

-d, --download <dirname>[...]
	Download one or multiple SRR files from srrdb.com.
	Options:
	-e, --extension=<extension>
		Extract files with the specified extension (sfv, nfo) from the SRR file.
		You can prune file paths with --prunePaths.
	-o, --stdout
		Print file data to stdout instead of saving the file.

-u, --upload <filename>[...]
	Uploads one or multiple files to srrdb.com.
	Options:
	-n, --username=<username> and -p, --password=<password>
		If you provide this it will post files using this account.
	-r, --release=<dirname>
		If you provide this it will post stored files to the specified release.
		Note that you need a valid login for this.
	-f, --folder=<folder>
		Optional to --release, this will set the folder of the stored file.

Options have to be provided before repeating main parameters.
```

## Tips for aliases

You're propably better off to use aliases for up- and downloading:
```sh
alias "srrdown"="srrdb --download --prunePaths"
alias "srrup"="srrdb --upload --username hashworks --password '"'foo$$bar'"'"
```
