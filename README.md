# srrdb Terminal Client

A terminal client to access [srrdb.com](http://www.srrdb.com), written in Go.

## Install

Archlinux users can use the AUR package [srrdb-terminal-client](https://aur.archlinux.org/packages/srrdb-terminal-client/).

Other users should simply download the latest release for your arch from [here](https://github.com/hashworks/srrdb-Terminal-Client/releases/latest) and move the executable to your `$PATH`.

## Usage

See `--help`:
```
-v, --version
	Shows the version and a few informations.

-s, --search <query>[...]
	Searches srrdb.com for releases.
	For a list of available keywords see http://www.srrdb.com/help#keywords

-d, --download <dirname>[...]
	Download one or multiple SRR files from srrdb.com.
	Options:
	-e, --extension=<extension>
		Saves only files with the specified extension from the SRR file.
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
		Optional to --release, this will set the folder of the stored files.
```

## Tips for aliases

You're propably better off to use aliases for up- and downloading:
```sh
alias "srrdown"="srrdb --download --prunePaths"
alias "srrup"="srrdb --upload --username hashworks --password '"'foo$$bar'"'"
```
