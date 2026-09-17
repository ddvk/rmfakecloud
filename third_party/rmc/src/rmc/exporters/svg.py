"""Convert blocks to svg file.

Code originally from https://github.com/lschwetlick/maxio through
https://github.com/chemag/maxio .
"""

import base64
import logging
import string
import typing as tp
from pathlib import Path

from rmscene import CrdtId, SceneTree, read_tree
from rmscene import scene_items as si
from rmscene.text import TextDocument

from .writing_tools import Pen, lookup_pen_color

_logger = logging.getLogger(__name__)

SCREEN_WIDTH = 1404
SCREEN_HEIGHT = 1872
SCREEN_DPI = 226

SCALE = 72.0 / SCREEN_DPI

PAGE_WIDTH_PT = SCREEN_WIDTH * SCALE
PAGE_HEIGHT_PT = SCREEN_HEIGHT * SCALE
X_SHIFT = PAGE_WIDTH_PT // 2


def scale(screen_unit: float) -> float:
    return screen_unit * SCALE


# For now, at least, the xx and yy function are identical to scale
xx = scale
yy = scale

TEXT_TOP_Y = -88
LINE_HEIGHTS = {
    # Based on a rm file having 4 anchors based on the line height I was able to find a value of
    # 69.5, but decided on 70 (to keep integer values)
    si.ParagraphStyle.PLAIN: 70,
    si.ParagraphStyle.BULLET: 35,
    si.ParagraphStyle.BULLET2: 35,
    si.ParagraphStyle.BOLD: 70,
    si.ParagraphStyle.HEADING: 150,
    si.ParagraphStyle.CHECKBOX: 35,
    si.ParagraphStyle.CHECKBOX_CHECKED: 35,

    # There appears to be another format code (value 0) which is used when the
    # text starts far down the page, which case it has a negative offset (line
    # height) of about -20?
    #
    # Probably, actually, the line height should be added *after* the first
    # line, but there is still something a bit odd going on here.
}

SVG_HEADER = string.Template("""<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" height="$height" width="$width" viewBox="$viewbox">""")


def rm_to_svg(rm_path, svg_path):
    """Convert `rm_path` to SVG at `svg_path`."""
    image_path = Path(rm_path).with_suffix("")
    with open(rm_path, "rb") as infile, open(svg_path, "wt") as outfile:
        tree = read_tree(infile)
        tree_to_svg(tree, outfile, image_path=image_path)


def read_template_svg(template_path: Path) -> str:
    lines = template_path.read_text().splitlines()
    return "\n".join(lines[2:-2])


def tree_to_svg(tree: SceneTree, output, include_template: Path | None = None, image_path: Path | None = None):
    """Convert Blocks to SVG."""

    # find the anchor pos for further use
    anchor_pos = build_anchor_pos(tree.root_text)
    _logger.debug("anchor_pos: %s", anchor_pos)

    # find the extremum along x and y
    x_min, x_max, y_min, y_max = get_bounding_box(tree.root, anchor_pos)
    width_pt = xx(x_max - x_min + 1)
    height_pt = yy(y_max - y_min + 1)
    _logger.debug("x_min, x_max, y_min, y_max: %.1f, %.1f, %.1f, %.1f ; scalded %.1f, %.1f, %.1f, %.1f",
                  x_min, x_max, y_min, y_max, xx(x_min), xx(x_max), yy(y_min), yy(y_max))

    # add svg header
    output.write(SVG_HEADER.substitute(width=width_pt,
                                       height=height_pt,
                                       viewbox=f"{xx(x_min)} {yy(y_min)} {width_pt} {height_pt}") + "\n")

    if include_template is not None:
        output.write(read_template_svg(include_template))
        output.write(f'\n\t<rect fill="url(#template)" x="{xx(x_min)}" y="{yy(y_min)}"'
                     f' width="{width_pt}" height="{height_pt}"/>\n')

    output.write(f'\t<g id="p1" style="display:inline">\n')

    if tree.root_text is not None:
        draw_text(tree.root_text, output)

    draw_group(tree.root, output, anchor_pos, image_path=image_path)

    # Closing page group
    output.write('\t</g>\n')
    # END notebook
    output.write('</svg>\n')


def build_anchor_pos(text: tp.Optional[si.Text]) -> tp.Dict[CrdtId, int]:
    """
    Find the anchor pos

    :param text: the root text of the remarkable file
    """
    anchor_pos = {}

    if text is not None:
        # Save anchor from text
        doc = TextDocument.from_scene_item(text)
        ypos = text.pos_y + TEXT_TOP_Y
        top_of_text = ypos
        for i, p in enumerate(doc.contents):
            anchor_pos[p.start_id] = ypos
            for subp in p.contents:
                for k in subp.i:
                    anchor_pos[k] = ypos  # TODO check these anchor are used
            ypos += LINE_HEIGHTS.get(p.style.value, 70)
        bottom_of_text = ypos

        # Special anchors: groups drawn at the top/bottom of the page,
        # not anchored to any specific paragraph.
        anchor_pos[CrdtId(0, 281474976710654)] = top_of_text
        anchor_pos[CrdtId(0, 281474976710655)] = bottom_of_text
    else:
        # No text: fall back to fixed positions at top/bottom of screen
        anchor_pos[CrdtId(0, 281474976710654)] = 0
        anchor_pos[CrdtId(0, 281474976710655)] = SCREEN_HEIGHT

    return anchor_pos


def get_anchor(item: si.Group, anchor_pos):
    anchor_x = 0.0
    anchor_y = 0.0
    if item.anchor_id is not None:
        assert item.anchor_origin_x is not None
        anchor_x = item.anchor_origin_x.value
        if item.anchor_id.value in anchor_pos:
            anchor_y = anchor_pos[item.anchor_id.value]
            _logger.debug("Group anchor: %s -> y=%.1f (scalded y=%.1f)",
                          item.anchor_id.value,
                          anchor_y,
                          yy(anchor_y))
        else:
            _logger.warning("Group anchor: %s is unknown!", item.anchor_id.value)

    return anchor_x, anchor_y


def get_bounding_box(item: si.Group,
                     anchor_pos: tp.Dict[CrdtId, int],
                     default: tp.Tuple[int, int, int, int] = (- SCREEN_WIDTH // 2, SCREEN_WIDTH // 2, 0, SCREEN_HEIGHT)) \
        -> tp.Tuple[int, int, int, int]:
    """
    Get the bounding box of the given item.
    The minimum size is the default size of the screen.

    :return: x_min, x_max, y_min, y_max: the bounding box in screen units (need to be scalded using xx and yy functions)
    """
    x_min, x_max, y_min, y_max = default

    for child_id in item.children:
        child = item.children[child_id]
        if isinstance(child, si.Group):
            anchor_x, anchor_y = get_anchor(child, anchor_pos)
            x_min_t, x_max_t, y_min_t, y_max_t = get_bounding_box(child, anchor_pos, (0, 0, 0, 0))
            x_min = min(x_min, x_min_t + anchor_x)
            x_max = max(x_max, x_max_t + anchor_x)
            y_min = min(y_min, y_min_t + anchor_y)
            y_max = max(y_max, y_max_t + anchor_y)
        elif isinstance(child, si.Line):
            x_min = min([x_min] + [p.x for p in child.points])
            x_max = max([x_max] + [p.x for p in child.points])
            y_min = min([y_min] + [p.y for p in child.points])
            y_max = max([y_max] + [p.y for p in child.points])
        elif isinstance(child, si.Image) and child.vertices:
            xs = child.vertices[0::2]
            ys = child.vertices[1::2]
            x_min = min([x_min] + xs)
            x_max = max([x_max] + xs)
            y_min = min([y_min] + ys)
            y_max = max([y_max] + ys)

    return x_min, x_max, y_min, y_max


def draw_group(item: si.Group, output, anchor_pos, image_path: Path | None = None):
    anchor_x, anchor_y = get_anchor(item, anchor_pos)
    output.write(f'\t\t<g id="{item.node_id}" transform="translate({xx(anchor_x)}, {yy(anchor_y)})">\n')
    for child_id in item.children:
        child = item.children[child_id]
        _logger.debug("Group child: %s %s", child_id, type(child))
        if _logger.root.level == logging.DEBUG:
            output.write(f'\t\t<!-- child {child_id} {type(child)} -->\n')
        if isinstance(child, si.Group):
            draw_group(child, output, anchor_pos, image_path=image_path)
        elif isinstance(child, si.Line):
            draw_stroke(child, output)
        elif isinstance(child, si.Image):
            draw_image(child, output, image_path=image_path)
    output.write(f'\t\t</g>\n')


def draw_stroke(item: si.Line, output):
    # print debug infos
    if _logger.root.level == logging.DEBUG:
        _logger.debug("Writing line: %s", item)
        output.write(f'\t\t\t<!-- Stroke tool: {item.tool.name} '
                     f'color: {item.color.name} thickness_scale: {item.thickness_scale} -->\n')

    # initiate the pen
    pen_color = lookup_pen_color(item.color.value, item.color_rgba)
    pen = Pen.create(item.tool.value, pen_color, item.thickness_scale)

    last_xpos = -1.
    last_ypos = -1.
    last_segment_width = segment_width = 0
    # Iterate through the point to form a polyline
    for point_id, point in enumerate(item.points):
        # align the original position
        xpos = point.x
        ypos = point.y
        if point_id % pen.segment_length == 0:
            # if there was a previous segment, end it
            if last_xpos != -1.:
                output.write('"/>\n')

            segment_color = pen.get_segment_color(point.speed, point.direction, point.width, point.pressure,
                                                  last_segment_width)
            segment_width = pen.get_segment_width(point.speed, point.direction, point.width, point.pressure,
                                                  last_segment_width)
            segment_opacity = pen.get_segment_opacity(point.speed, point.direction, point.width, point.pressure,
                                                      last_segment_width)
            # create the next segment of the stroke
            output.write('\t\t\t<polyline ')
            output.write(f'style="fill:none; stroke:{segment_color}; '
                         f'stroke-width:{scale(segment_width):.3f}; opacity:{segment_opacity}" ')
            output.write(f'stroke-linecap="{pen.stroke_linecap}" ')
            output.write('points="')
            if last_xpos != -1.:
                # Join to previous segment
                output.write(f'{xx(last_xpos):.3f},{yy(last_ypos):.3f} ')
        # store the last position
        last_xpos = xpos
        last_ypos = ypos
        last_segment_width = segment_width

        # add current point
        output.write(f'{xx(xpos):.3f},{yy(ypos):.3f} ')

    # end stroke
    output.write('" />\n')


def draw_text(text: si.Text, output):
    output.write('\t\t<g class="root-text" style="display:inline">')

    # add some style to get readable text
    output.write('''
            <style>
                text.heading {
                    font: 14pt serif;
                }
                text.bold {
                    font: 8pt sans-serif bold;
                }
                text, text.plain {
                    font: 7pt sans-serif;
                }
            </style>
''')

    y_offset = TEXT_TOP_Y

    doc = TextDocument.from_scene_item(text)
    for p in doc.contents:
        y_offset += LINE_HEIGHTS.get(p.style.value, 70)

        xpos = text.pos_x
        ypos = text.pos_y + y_offset
        cls = p.style.value.name.lower()
        if str(p):
            # TODO: this doesn't take into account the CrdtStr.properties (font-weight/font-style)
            if _logger.root.level == logging.DEBUG:
                output.write(f'\t\t\t<!-- Text line char_id: {p.start_id} -->\n')
            output.write(f'\t\t\t<text x="{xx(xpos)}" y="{yy(ypos)}" class="{cls}">{str(p).strip()}</text>\n')
    output.write('\t\t</g>\n')


def _image_mime(data: bytes) -> str:
    if data.startswith(b"\x89PNG\r\n\x1a\n"):
        return "image/png"
    if data[:3] == b"\xff\xd8\xff":
        return "image/jpeg"
    if data[:6] in (b"GIF87a", b"GIF89a"):
        return "image/gif"
    if data.startswith(b"RIFF") and data[8:12] == b"WEBP":
        return "image/webp"
    return "image/png"


def draw_image(image: si.Image, output, image_path: Path | None = None):
    if not image.filename or image_path is None:
        _logger.warning("Skipping image without filename/path: %s", image.filename)
        return
    path = image_path / image.filename
    if not path.is_file():
        _logger.warning("Missing page image file: %s", path)
        return
    if len(image.vertices) < 10:
        _logger.warning("Image vertices too short: %s", len(image.vertices))
        return

    left = image.vertices[0]
    top = image.vertices[1]
    right = image.vertices[4]
    bottom = image.vertices[9]

    height = bottom - top
    width = right - left
    data = path.read_bytes()
    mime = _image_mime(data)
    href = base64.b64encode(data).decode()
    output.write(f'\t\t\t<g transform="translate({xx(left)}, {yy(top)})">\n')
    output.write(f'\t\t\t\t<image height="{yy(height)}" width="{xx(width)}" href="data:{mime};base64,{href}"/>\n')
    output.write('\t\t\t</g>\n')
