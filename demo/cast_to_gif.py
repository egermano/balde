#!/usr/bin/env python3
"""Render demo/balde-quickstart.cast into an animated GIF.

The GIF is what GitHub can display inline in the README (a .cast needs the
asciinema player, which GitHub strips). The cast events produced by
record_cast.py contain only printable text and \\r\\n — no cursor control —
so replaying them into a line buffer is a faithful terminal emulation.

Output: demo/balde-quickstart.gif
"""

import json
import os

from PIL import Image, ImageDraw, ImageFont

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CAST = os.path.join(REPO, "demo", "balde-quickstart.cast")
OUT = os.path.join(REPO, "demo", "balde-quickstart.gif")

COLS, ROWS = 80, 24
FPS = 3.33  # one frame every 0.3 s
BG = (13, 17, 23)      # github dark canvas
FG = (230, 237, 243)   # github dark text
PROMPT_FG = (63, 185, 80)  # green for typed command lines
FONT_PATH = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"
FONT_SIZE = 15

def load_events(path: str):
    with open(path, encoding="utf-8") as f:
        header = json.loads(f.readline())
        events = [json.loads(line) for line in f]
    return header, [e for e in events if len(e) == 3 and e[1] == "o"]

class Screen:
    """Fixed-size text screen fed by cast output events."""

    def __init__(self, cols: int, rows: int):
        self.cols, self.rows = cols, rows
        self.lines: list[str] = []
        self.typed = ""  # current command line being typed (starts with "$ ")

    def feed(self, text: str) -> None:
        for ch in text:
            if ch == "\r":
                continue
            if ch == "\n":
                self.lines.append(self.typed)
                self.typed = ""
            else:
                self.typed += ch
                if len(self.typed) > self.cols:
                    self.typed = self.typed[-self.cols:]

    def snapshot(self) -> list[tuple[str, bool]]:
        """Return visible lines as (text, is_prompt) pairs."""
        visible = (self.lines + [self.typed] if self.typed else self.lines)[-self.rows:]
        out = []
        for line in visible:
            out.append((line[: self.cols], line.startswith("$ ")))
        while len(out) < self.rows:
            out.append(("", False))
        return out

def render_frame(snapshot, font, char_w: int, line_h: int) -> Image.Image:
    pad = 12
    img = Image.new(
        "RGB",
        (COLS * char_w + pad * 2, ROWS * line_h + pad * 2),
        BG,
    )
    draw = ImageDraw.Draw(img)
    y = pad
    for text, is_prompt in snapshot:
        if text:
            draw.text((pad, y), text, font=font, fill=PROMPT_FG if is_prompt else FG)
        y += line_h
    return img

def main() -> None:
    header, events = load_events(CAST)
    duration = events[-1][0]
    total_frames = int(duration * FPS) + 1
    frame_interval = 1.0 / FPS

    font = ImageFont.truetype(FONT_PATH, FONT_SIZE)
    char_w = int(font.getbbox("M")[2] - font.getbbox("M")[0]) + 1
    line_h = FONT_SIZE + 4

    screen = Screen(COLS, ROWS)
    frames = []
    ei = 0
    for fi in range(total_frames):
        target = fi * frame_interval
        while ei < len(events) and events[ei][0] <= target:
            screen.feed(events[ei][2])
            ei += 1
        frames.append(render_frame(screen.snapshot(), font, char_w, line_h))

    frames[0].save(
        OUT,
        save_all=True,
        append_images=frames[1:],
        duration=int(frame_interval * 1000),
        loop=0,
        optimize=True,
    )
    size = os.path.getsize(OUT) / 1024
    print(f"wrote {OUT}: {len(frames)} frames, {duration:.1f}s, {size:.0f} KiB")

if __name__ == "__main__":
    main()
