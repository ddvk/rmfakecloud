"""
Common code for writing tools.

Code originally from https://github.com/lschwetlick/maxio through https://github.com/chemag/maxio
"""

import logging
import math

from rmscene.scene_items import Pen as PenType
from rmscene.scene_items import PenColor

_logger = logging.getLogger(__name__)

# color_id to RGB conversion
# 1. we use "color_id" for a unique, proprietary ID for colors,
#   (see scene_stream.py):
RM_PALETTE = {
    # Color code can be obtained from extraMetadata in the .content file
    PenColor.BLACK: (0, 0, 0),
    PenColor.GRAY: (144, 144, 144),
    PenColor.WHITE: (255, 255, 255),
    PenColor.YELLOW: (251, 247, 25),
    PenColor.GREEN: (0, 255, 0),
    PenColor.PINK: (255, 192, 203),
    PenColor.BLUE: (78, 105, 201),
    PenColor.RED: (179, 62, 57),
    PenColor.GRAY_OVERLAP: (125, 125, 125),
    #! Skipped as different colors are used for highlights
    #! PenColor.HIGHLIGHT = ...
    PenColor.GREEN_2: (161, 216, 125),
    PenColor.CYAN: (139, 208, 229),
    PenColor.MAGENTA: (183, 130, 205),
    PenColor.YELLOW_2: (247, 232, 81),
}


def clamp(value):
    """
    Clamp value between 0 and 1.
    """
    return min(max(value, 0), 1)


def lookup_pen_color(pen_color_value, explicit_color_rgba):
    if explicit_color_rgba is not None:
        return explicit_color_rgba
    if pen_color_value in RM_PALETTE:
        color = RM_PALETTE[pen_color_value]
    else:
        _logger.warning("Unknown pen color: %s", pen_color_value)
        color = RM_PALETTE[PenColor.BLACK]
    if len(color) == 3:
        return (*color, 255)
    else:
        return color


class Pen:
    def __init__(self, name, base_width, base_color):
        self.base_width = base_width
        self.base_color = base_color  # rgba
        self.name = name
        self.segment_length = 1000
        self.base_opacity = 1
        # initial stroke values
        self.stroke_linecap = "round"
        self.stroke_opacity = 1

    # note that the units of the points have had their units converted
    # in scene_stream.py
    # speed = d.read_float32() * 4
    # ---> replace speed with speed / 4 [input]
    # direction = 255 * d.read_float32() / (math.pi * 2)
    # ---> replace tilt with direction_to_tilt() [input]
    @classmethod
    def direction_to_tilt(cls, direction):
        return direction * (math.pi * 2) / 255

    # width = int(round(d.read_float32() * 4))
    # ---> replace width with width / 4 [input]
    # ---> replace width with 4 * width [output]
    # pressure = d.read_float32() * 255
    # ---> replace pressure with pressure / 255 [input]

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        return self.base_width

    def get_segment_color(self, speed, direction, width, pressure, last_width):
        return "rgba" + str(tuple(self.base_color))

    def get_segment_opacity(self, speed, direction, width, pressure, last_width):
        return self.base_opacity

    @classmethod
    def create(cls, pen_nr, color_id, width):
        # Brush
        if pen_nr in (PenType.PAINTBRUSH_1, PenType.PAINTBRUSH_2):
            return Brush(width, color_id)
        # Calligraphy (spelling mistake in rmscene will eventually be fixed)
        elif pen_nr == PenType.CALIGRAPHY:
            return Calligraphy(width, color_id)
        # Marker
        elif pen_nr in (PenType.MARKER_1, PenType.MARKER_2):
            return Marker(width, color_id)
        # BallPoint
        elif pen_nr in (PenType.BALLPOINT_1, PenType.BALLPOINT_2):
            return Ballpoint(width, color_id)
        # Fineliner
        elif pen_nr in (PenType.FINELINER_1, PenType.FINELINER_2):
            return Fineliner(width, color_id)
        # Pencil
        elif pen_nr in (PenType.PENCIL_1, PenType.PENCIL_2):
            return Pencil(width, color_id)
        # Mechanical Pencil
        elif pen_nr in (PenType.MECHANICAL_PENCIL_1, PenType.MECHANICAL_PENCIL_2):
            return MechanicalPencil(width, color_id)
        # Highlighter
        elif pen_nr in (PenType.HIGHLIGHTER_1, PenType.HIGHLIGHTER_2):
            width = 15
            return Highlighter(width, color_id)
        elif pen_nr == PenType.SHADER:
            # TODO: check if this is correct
            width = 12
            return Shader(width, color_id)
        # Erase area
        elif pen_nr == PenType.ERASER_AREA:
            return EraseArea(width, color_id)
        # Eraser
        elif pen_nr == PenType.ERASER:
            color_id = 2
            return Eraser(width, color_id)
        raise Exception(f'Unknown pen_nr: {pen_nr}')


class Fineliner(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Fineliner", base_width * 1.8, base_color)


class Ballpoint(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Ballpoint", base_width, base_color)
        self.segment_length = 5

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        segment_width = (0.5 + pressure / 255) + (width / 4) - 0.5 * ((speed / 4) / 50)
        return segment_width

    def get_segment_color(self, speed, direction, width, pressure, last_width):
        intensity = (0.1 * - ((speed / 4) / 35)) + (1.2 * pressure / 255) + 0.5
        intensity = clamp(intensity)
        # using segment color not opacity because the dots interfere with each other.
        # Color must be 255 rgb
        segment_color = [min(int(abs(intensity - 1) * 255), 60)] * 3
        return "rgb" + str(tuple(segment_color))

    # def get_segment_opacity(self, speed, direction, width, pressure, last_width):
    #     segment_opacity = (0.2 * - ((speed / 4) / 35)) + (0.8 * pressure / 255)
    #     segment_opacity *= segment_opacity
    #     segment_opacity = self.clamp(segment_opacity)
    #     return segment_opacity


class Marker(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Marker", base_width, base_color)
        self.segment_length = 3

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        segment_width = 0.9 * ((width / 4) - 0.4 * self.direction_to_tilt(direction)) + (0.1 * last_width)
        return segment_width


class Pencil(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Pencil", base_width, base_color)
        self.segment_length = 2

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        segment_width = 0.7 * ((((0.8 * self.base_width) + (0.5 * pressure / 255)) * (width / 4))
                               - (0.25 * self.direction_to_tilt(direction) ** 1.8) - (0.6 * (speed / 4) / 50))
        # segment_width = 1.3*(((self.base_width * 0.4) * pressure) - 0.5 * ((self.direction_to_tilt(direction) ** 0.5)) + (0.5 * last_width))
        max_width = self.base_width * 10
        segment_width = segment_width if segment_width < max_width else max_width
        return segment_width

    def get_segment_opacity(self, speed, direction, width, pressure, last_width):
        segment_opacity = (0.1 * - ((speed / 4) / 35)) + (1 * pressure / 255)
        segment_opacity = clamp(segment_opacity) - 0.1
        return segment_opacity


class MechanicalPencil(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Mechanical Pencil", base_width ** 2, base_color)
        self.base_opacity = 0.7


class Brush(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Brush", base_width, base_color)
        self.segment_length = 2
        self.stroke_linecap = "round"
        self.opacity = 1

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        segment_width = 0.7 * (((1 + (1.4 * pressure / 255)) * (width / 4))
                               - (0.5 * self.direction_to_tilt(direction)) - ((speed / 4) / 50))  # + (0.2 * last_width)
        return segment_width

    def get_segment_color(self, speed, direction, width, pressure, last_width):
        intensity = ((pressure / 255) ** 1.5 - 0.2 * ((speed / 4) / 50)) * 1.5
        intensity = clamp(intensity)
        # using segment color not opacity because the dots interfere with each other.
        # Color must be 255 rgb
        rev_intensity = abs(intensity - 1)
        segment_color = [int(rev_intensity * (255 - self.base_color[0])),
                         int(rev_intensity * (255 - self.base_color[1])),
                         int(rev_intensity * (255 - self.base_color[2])),
                         self.base_color[3]]

        return "rgba" + str(tuple(segment_color))


class Highlighter(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Highlighter", base_width, base_color)
        self.stroke_linecap = "square"
        self.base_opacity = 0.3
        self.stroke_opacity = 0.2


class Shader(Pen):
    
    def __init__(self, base_width, base_color):
        super().__init__("Shader", base_width, base_color)
        self.stroke_linecap = "round"
        self.base_opacity = 0.1
        # self.stroke_opacity = 0.2
        self.name = "Shader"

class Eraser(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Eraser", base_width * 2, base_color)
        self.stroke_linecap = "square"


class EraseArea(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Erase Area", base_width, base_color)
        self.stroke_linecap = "square"
        self.base_opacity = 0


class Calligraphy(Pen):
    def __init__(self, base_width, base_color):
        super().__init__("Calligraphy", base_width, base_color)
        self.segment_length = 2

    def get_segment_width(self, speed, direction, width, pressure, last_width):
        segment_width = 0.9 * (((1 + pressure / 255) * (width / 4))
                               - 0.3 * self.direction_to_tilt(direction)) + (0.1 * last_width)
        return segment_width
