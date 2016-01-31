# srrdb Terminal Client

A terminal client to access [srrdb.com](http://www.srrdb.com), written in Go.

## Install

Simply download the latest release for your arch from [here](https://github.com/hashworks/srrdbTerminalClient/releases/latest) and execute it.
For global access place the executable in your `$PATH`.

## Usage

See `--help`:
```
-v, --version
	Shows the version and a few informations.

-s, --search <query>[...]
	Searches srrdb.com for releases.
	For a list of available keywords see http://www.srrdb.com/help#keywords

-d, --download <dirname>[...]
	Download a SRR file from srrdb.com.
	Options:
	-e, --extension=<extension>
		Saves only files with the specified extension from the SRR file.
		You can prune file paths with -p, --prunePaths.
	-o, --stdout
		Print file data to stdout instead of saving the file.

-u, --upload <filename>[...]
	Uploads one or multiple files to srrdb.com.
	Options:
	--username=<username> and --password=<password>
		If you provide this it will post the SRR file using this account.
```

## TODO

* Add Tests
