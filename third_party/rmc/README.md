# rmc

Command line tool for converting to/from remarkable `.rm` version 6 (software version 3) files.

## Installation

To install in your current Python environment:

    pip install rmc
    
Or use [pipx](https://pypa.github.io/pipx/) to install in an isolated environment (recommended):

    pipx install rmc

## Usage

Convert a remarkable v6 file to other formats, specified by `-t FORMAT`:

    $ rmc -t markdown file.rm
    Text in the file is printed to standard output.

Specify the filename to write the output to with `-o`:

    $ rmc -t svg -o file.svg file.rm
    
The format is guessed based on the filename if not specified:
    
    $ rmc file.rm -o file.pdf

Create a `.rm` file containing the text in `text.md`:

    $ rmc -t rm text.md -o text.rm

## SVG/PDF Conversion Status

Right now the converter works well while there are no text boxes. If you add text boxes, there are x issues:

1. if the text box contains multiple lines, the lines are actually printed in the same line, and
2. the position of the strokes gets corrupted.

# Acknowledgements

`rmc` uses [rmscene](https://github.com/ricklupton/rmscene) to read the `.rm` files, for which https://github.com/ddvk/reader helped a lot in figuring out the structure and meaning of the files.

[@chemag](https://github.com/chemag) added initial support for converting to svg and pdf.

[@Seb-sti1](https://github.com/Seb-sti1) made lots of improvements to svg export and updating to newer `rmscene` versions.

[@ChenghaoMou](https://github.com/ChenghaoMou) added support for new pen types/colours.

[@EelcovanVeldhuizen](https://github.com/EelcovanVeldhuizen) for code updates/fixes.

[@p4xel](https://github.com/p4xel) for code fixes.

[@Azeirah](https://github.com/Azeirah)

[@darobin](https://github.com/darobin)
