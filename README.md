# Invisible Items

A CLI tool to generate an invisible custom model for every minecraft item available.

## Using

Providing you have installed GO, just run the program with `go run .`. (Alternatively, you can ofcourse build an executeable with `go build .`)

The finished resource pack is directly in the `pack` folder. Ready to use.

## Flags

There are two command line flags available, `model` and `mc`.
They can be used with either a single dash (`-flagname`) or a double dash (`--flagname`). Assigning a value can be done by having a Space (` `) or an equal sign (`=`) between the flagname and the value.

So this would all be the same result:

- `-flagname foo`
- `-flagname=foo`
- `--flagname foo`
- `--flagname=foo`


### `model` unsigned int

Default Value of `1`.

Specify the custom model data to use for the invisible items.

Examples:

- `go run . --model 9001` (`"custom_model_data": 9001`)
- `go run . -model=9`
- `go run . -model 100`
- `go run . --model=0` (results in every item beeing infisible by default)

### `mc` string

Default value of `release`.

Specify the Minecraft version to use. The tool automatically downloads and generates all models as well as the correct pack version number for the `pack.mcmeta`.

Special values are `release` and `snapshot`, which fetch the lastest release version or lastest snapsot version respectively.  
Any other values are used as a Minecraft version ID.

Examples:

- `go run . --mc 1.16.5`
- `go run . -mc=snapshot`
- `go run . --mc=23w04a`
- `go run . -mc 1.20.4`

Keep in mind, that `custom_model_data` was introduced in [18w43a (1.14)](https://minecraft.wiki/w/Java_Edition_18w43a#:~:text=custom_model_data) and won't work in clients with a lower version! However, this tool just generates the pack like normal and does not check if the specified Minecraft version is before 18w43a.
