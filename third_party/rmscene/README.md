# rmscene

Python library to read v6 files from reMarkable tables (software version 3).

In particular, this version introduces the ability to include text as well as drawn lines. Extracting this text is the original motivation to develop this library, but it also can read much of the other types of data in the reMarkable files.

To convert rm files to other formats, you can use [rmc](https://github.com/ricklupton/rmc), which combines this library with code for converting lines to SVG, PDF, and simple Markdown.

## Changelog

### Unreleased

### v0.8.0
  
New features:

- Store highlighter colour in `color_rgba` field on `GlyphRange` ([#50](https://github.com/ricklupton/rmscene/pull/50))
- Store `SceneInfo` on `SceneTree` for convenient access

Performance improvements:

- Use Kahn's algorithm for CRDT topological sort -- thanks to @andymatuschak ([#49](https://github.com/ricklupton/rmscene/pull/49))

Fixes:

- Fix author sort order for concurrent CRDT inserts
 

### v0.7.0

Change in block properties:

- Some SceneInfo values are now optional ([#40](https://github.com/ricklupton/rmscene/issues/40))

New feature:

- Add support for `paper_size` field on some SceneInfo blocks

### v0.6.1

Fixes:

- Fix AssertionError when some ids are missing in a `CrdtSequence` ([#36](https://github.com/ricklupton/rmscene/pull/36))
- Fix ValueError when the node_id is missing in a `SceneGroupItemBlock` ([#16](https://github.com/ricklupton/rmscene/issues/16)) 
- Store any unparsed data in blocks as raw bytes to allow for round-trip saving of files written in a newer format than the parsing code knows about.

### v0.6.0

New features:

- Add support for new blocks: `0x0D` (SceneInfo) and `0x08` (SceneTombstoneItemBlock) ([#24](https://github.com/ricklupton/rmscene/pull/24/))
- Add support for `move_id` field on some SceneLineItems ([#24](https://github.com/ricklupton/rmscene/pull/24/))
- Add support for new pen types and colours ([#31](https://github.com/ricklupton/rmscene/pull/31))

### v0.5.0

Breaking changes:

- The `start` property of `GlyphRange` items is now optional
  ([#15](https://github.com/ricklupton/rmscene/pull/15/)).
- The representation of formatted text spans has changed. Rather than
  using nested structures like `BoldSpan` and `ItalicSpan`, the
  `CrdtStr` objects now have optional text properties like
  `font-weight` and `font-style`. This simplifies the parsing code and
  the resulting data structure.

New features:

- Improved error recovery. An error during parsing, or an unknown block type,
  results in an `UnreadableBlock` containing the data that could not be read, so
  that parsing of other blocks can continue.
- Compatible with new reMarkable software version 3.6 format for
  highlighted text
  ([#15](https://github.com/ricklupton/rmscene/pull/15/)).
- New methods `read_bool_optional` and similar of `TaggedBlockReader`
  which return a default value if no matching tagged value is present
  in the block.
  
Other changes and fixes:

- The `value` attribute of scene item blocks, which was not being used, has been
  removed.
- Check more carefully for sub-blocks
  ([#17](https://github.com/ricklupton/rmscene/issues/17#issuecomment-1701071477)).
- Type hints fixed for `expand_text_items`.

### v0.4.0

Breaking changes:

- Rename `scene_items.TextFormat` to `ParagraphStyle` to better describe its
  meaning, now that we have inline bold/italic text styles.
- Remove methods from `scene_items.Text` object; use `text.TextDocument`
  instead.
- Writer: experimental change to emulate different reMarkable software versions
  by passing `{"version": "3.2.2"}` options to `write_blocks`. This allows us to
  continue to test round-trip reading and writing of old test files as new data
  values are added. Replaces `"line_version"` option.
  
New features:

- Parse text formatting information (bold and italic) introduced in reMarkable
  software version 3.3.

Other changes:

- Allow empty text items and unknown text formats without throwing exceptions.
- When extra data is present in the file, log the unrecognised bytes at DEBUG
  logging level along with the call stack, to make it easier to figure out where
  the code needs to be modified to read new data.
- Parse new data values (with unknown meaning) in PageInfoBlock and
  MigrationInfoBlock.

### v0.3.0

- Introduce `CrdtSequence` type to handle the different places that CRDT
  sequences are used, not just for text.
- Introduce `scene_items` module with data structures representing the data,
  independently from the `Block`s used to serialize them to `.rm` files.
- Introduce a `SceneTree` structure which holds the `SceneItem`s in
  groups/layers.
- Move Text data from `RootTextBlock` to `scene_items.Text` class, which
  includes methods for extracting lines of text and formatting.
- Text lines now include the trailing newline character.
- Read `GlyphRange` scene items, representing highlighted text in PDFs.

### v0.2.0

- Try to be more robust to unexpected data introduced by newer reMarkable software versions.
- Only warn once if unknown data is present, rather than for every block.
- Small API change to return type of `read_block` and `read_subblock` methods.

### v0.1.0

- Initial release

## Acknowledgements

https://github.com/ddvk/reader helped a lot in figuring out the structure and meaning of the files.  [@adq](https://github.com/adq) discovered a means to get debug output (see [issue 25](https://github.com/ricklupton/rmscene/issues/25)) which is very helpful for understanding the format.

Contributors:
- [@Azeirah](https://github.com/Azeirah) -- code and reporting issues
- [@adq](https://github.com/adq) -- code and reporting issues
- [@dotlambda](https://github.com/dotlambda) -- packaging
- [@ChenghaoMou](https://github.com/ChenghaoMou)
