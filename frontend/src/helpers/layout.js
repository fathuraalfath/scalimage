// Config for paper sizes at standard screen resolution (72 DPI)
export const PAPER_SIZES = {
  A4: { name: 'A4', width: 595, height: 842 },
  F4: { name: 'F4', width: 609, height: 935 },
  A5: { name: 'A5', width: 420, height: 595 },
  Square: { name: 'Square', width: 600, height: 600 }
};

// Pre-defined percentage-based grid layout templates.
// x, y, w, h are defined from 0.0 to 1.0.
export const TEMPLATES = {
  1: [
    {
      name: 'Single Image',
      slots: [{ x: 0, y: 0, w: 1, h: 1 }]
    }
  ],
  2: [
    {
      name: 'Split Vertical',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 1 },
        { x: 0.5, y: 0, w: 0.5, h: 1 }
      ]
    },
    {
      name: 'Split Horizontal',
      slots: [
        { x: 0, y: 0, w: 1, h: 0.5 },
        { x: 0, y: 0.5, w: 1, h: 0.5 }
      ]
    }
  ],
  3: [
    {
      name: '3 Columns',
      slots: [
        { x: 0, y: 0, w: 0.3333, h: 1 },
        { x: 0.3333, y: 0, w: 0.3333, h: 1 },
        { x: 0.6666, y: 0, w: 0.3334, h: 1 }
      ]
    },
    {
      name: '1 Left, 2 Right',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 1 },
        { x: 0.5, y: 0, w: 0.5, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.5, h: 0.5 }
      ]
    },
    {
      name: '1 Top, 2 Bottom',
      slots: [
        { x: 0, y: 0, w: 1, h: 0.5 },
        { x: 0, y: 0.5, w: 0.5, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.5, h: 0.5 }
      ]
    }
  ],
  4: [
    {
      name: '2x2 Grid',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 0.5 },
        { x: 0.5, y: 0, w: 0.5, h: 0.5 },
        { x: 0, y: 0.5, w: 0.5, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.5, h: 0.5 }
      ]
    },
    {
      name: '1 Left, 3 Right',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 1 },
        { x: 0.5, y: 0, w: 0.5, h: 0.3333 },
        { x: 0.5, y: 0.3333, w: 0.5, h: 0.3333 },
        { x: 0.5, y: 0.6666, w: 0.5, h: 0.3334 }
      ]
    },
    {
      name: '4 Columns',
      slots: [
        { x: 0, y: 0, w: 0.25, h: 1 },
        { x: 0.25, y: 0, w: 0.25, h: 1 },
        { x: 0.5, y: 0, w: 0.25, h: 1 },
        { x: 0.75, y: 0, w: 0.25, h: 1 }
      ]
    }
  ],
  5: [
    {
      name: '2 Top, 3 Bottom',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 0.5 },
        { x: 0.5, y: 0, w: 0.5, h: 0.5 },
        { x: 0, y: 0.5, w: 0.3333, h: 0.5 },
        { x: 0.3333, y: 0.5, w: 0.3333, h: 0.5 },
        { x: 0.6666, y: 0.5, w: 0.3334, h: 0.5 }
      ]
    },
    {
      name: '1 Left, 4 Grid',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 1 },
        { x: 0.5, y: 0, w: 0.25, h: 0.5 },
        { x: 0.75, y: 0, w: 0.25, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.75, y: 0.5, w: 0.25, h: 0.5 }
      ]
    }
  ],
  6: [
    {
      name: '3x2 Grid',
      slots: [
        { x: 0, y: 0, w: 0.3333, h: 0.5 },
        { x: 0.3333, y: 0, w: 0.3333, h: 0.5 },
        { x: 0.6666, y: 0, w: 0.3334, h: 0.5 },
        { x: 0, y: 0.5, w: 0.3333, h: 0.5 },
        { x: 0.3333, y: 0.5, w: 0.3333, h: 0.5 },
        { x: 0.6666, y: 0.5, w: 0.3334, h: 0.5 }
      ]
    },
    {
      name: '2x3 Grid',
      slots: [
        { x: 0, y: 0, w: 0.5, h: 0.3333 },
        { x: 0.5, y: 0, w: 0.5, h: 0.3333 },
        { x: 0, y: 0.3333, w: 0.5, h: 0.3333 },
        { x: 0.5, y: 0.3333, w: 0.5, h: 0.3333 },
        { x: 0, y: 0.6666, w: 0.5, h: 0.3334 },
        { x: 0.5, y: 0.6666, w: 0.5, h: 0.3334 }
      ]
    }
  ],
  7: [
    {
      name: '3 Top, 4 Bottom',
      slots: [
        { x: 0, y: 0, w: 0.3333, h: 0.5 },
        { x: 0.3333, y: 0, w: 0.3333, h: 0.5 },
        { x: 0.6666, y: 0, w: 0.3334, h: 0.5 },
        { x: 0, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.25, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.75, y: 0.5, w: 0.25, h: 0.5 }
      ]
    }
  ],
  8: [
    {
      name: '4x2 Grid',
      slots: [
        { x: 0, y: 0, w: 0.25, h: 0.5 },
        { x: 0.25, y: 0, w: 0.25, h: 0.5 },
        { x: 0.5, y: 0, w: 0.25, h: 0.5 },
        { x: 0.75, y: 0, w: 0.25, h: 0.5 },
        { x: 0, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.25, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.5, y: 0.5, w: 0.25, h: 0.5 },
        { x: 0.75, y: 0.5, w: 0.25, h: 0.5 }
      ]
    }
  ]
};

// Resolves normalized slots to pixel values based on width and height.
export function resolveSlots(slots, width, height) {
  return slots.map(slot => ({
    x: Math.round(slot.x * width),
    y: Math.round(slot.y * height),
    width: Math.round(slot.w * width),
    height: Math.round(slot.h * height)
  }));
}
