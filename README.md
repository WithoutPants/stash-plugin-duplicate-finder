# Stash plugin: Duplicate finder

This is a plugin for stash. It adds a `Find duplicate scenes` task. This task processes the vtt sprite files in your stash library, performing a perceptual hash. Any scenes that it detects are output in the plugin log. 

Optionally, it can tag duplicate scenes with a (existing) tag, and it can populate the details field of the duplicate scene with the ids of its duplicates.

# How to use

Untar the release for your platform into your `plugins` stash directory and reload plugins (or restart stash). A new task should be present in the Tasks page. A documented default configuration file is included.

Running the plugin will scan your library, and duplicate scenes will be marked with a 'duplicate' tag by default (this can be edited in the configuration file). 

Please note that you will need to create a tag called 'duplicate' *before scanning* in order for things to be properly labeled, otherwise you will not see any difference to your library.

*NOTE:* the plugin uses the sprite files to find duplicates. This means that if you remove a file from your stash library but do not remove the generated files (specifically the generated sprite file), then the plugin will continue to use the sprite file for duplicate detection.

# How to build

`make build` - builds the plugin executable for your platform
`make build-release-docker` - performs cross compilation in the `stashapp/compiler:develop` docker image and builds release tars

# Command-line mode

Command-line mode can be run by providing the sprite directory as a command line parameter. In this mode, it outputs a `duplicates.csv` file containing matching checksums with the match score. It is intended for debugging and fine-tuning the sensitivity. The execution can be stopped safely by touching a `.stop` file in the cwd.
