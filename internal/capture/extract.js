(opts) => {
  const imageMode = (opts && opts.imageMode) || 'full';
  const jpegQuality = (opts && opts.jpegQuality) || 0.7;

  const SKIP_TAGS = new Set([
    'SCRIPT', 'STYLE', 'NOSCRIPT', 'LINK', 'META', 'HEAD', 'BR', 'WBR',
  ]);

  function parseColor(str) {
    if (!str || str === 'transparent' || str === 'rgba(0, 0, 0, 0)') {
      return { r: 0, g: 0, b: 0, a: 0 };
    }
    const m = str.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/);
    if (m) {
      return {
        r: parseInt(m[1], 10),
        g: parseInt(m[2], 10),
        b: parseInt(m[3], 10),
        a: m[4] !== undefined ? parseFloat(m[4]) : 1,
      };
    }
    return { r: 0, g: 0, b: 0, a: 0 };
  }

  function parseBorder(style, side) {
    const width = parseFloat(style.getPropertyValue(`border-${side}-width`)) || 0;
    const borderStyle = style.getPropertyValue(`border-${side}-style`) || 'none';
    const color = parseColor(style.getPropertyValue(`border-${side}-color`));
    return { width, style: borderStyle, color };
  }

  function parseBorderRadius(style) {
    return {
      topLeft: parseFloat(style.borderTopLeftRadius) || 0,
      topRight: parseFloat(style.borderTopRightRadius) || 0,
      bottomRight: parseFloat(style.borderBottomRightRadius) || 0,
      bottomLeft: parseFloat(style.borderBottomLeftRadius) || 0,
    };
  }

  function parseBoxShadows(str) {
    if (!str || str === 'none') return [];
    const shadows = [];
    // Split on commas that aren't inside parentheses
    const parts = [];
    let depth = 0, start = 0;
    for (let i = 0; i < str.length; i++) {
      if (str[i] === '(') depth++;
      else if (str[i] === ')') depth--;
      else if (str[i] === ',' && depth === 0) {
        parts.push(str.substring(start, i).trim());
        start = i + 1;
      }
    }
    parts.push(str.substring(start).trim());

    for (const part of parts) {
      const inset = part.includes('inset');
      const cleaned = part.replace('inset', '').trim();
      // Extract the color portion (rgb/rgba or named color)
      const colorMatch = cleaned.match(/(rgba?\([^)]+\))/);
      let color = { r: 0, g: 0, b: 0, a: 1 };
      let rest = cleaned;
      if (colorMatch) {
        color = parseColor(colorMatch[1]);
        rest = cleaned.replace(colorMatch[1], '').trim();
      }
      const nums = rest.match(/-?[\d.]+/g);
      if (nums && nums.length >= 2) {
        shadows.push({
          offsetX: parseFloat(nums[0]) || 0,
          offsetY: parseFloat(nums[1]) || 0,
          blur: parseFloat(nums[2]) || 0,
          spread: parseFloat(nums[3]) || 0,
          color,
          inset,
        });
      }
    }
    return shadows;
  }

  function hasTransparency(ctx, w, h) {
    const step = Math.max(1, Math.floor((w * h) / 1000));
    const data = ctx.getImageData(0, 0, w, h).data;
    for (let i = 3; i < data.length; i += step * 4) {
      if (data[i] < 255) return true;
    }
    return false;
  }

  function captureImage(el) {
    try {
      if (el.tagName === 'SVG' || el instanceof SVGElement) {
        const serializer = new XMLSerializer();
        const svgStr = serializer.serializeToString(el);
        return 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svgStr)));
      }
      if (el.tagName === 'IMG' && el.naturalWidth > 0) {
        const canvas = document.createElement('canvas');
        const rect = el.getBoundingClientRect();
        if (imageMode === 'compact') {
          // Draw at rendered dimensions, use JPEG for opaque images
          const drawW = Math.round(rect.width) || el.naturalWidth;
          const drawH = Math.round(rect.height) || el.naturalHeight;
          canvas.width = drawW;
          canvas.height = drawH;
          const ctx = canvas.getContext('2d');
          ctx.drawImage(el, 0, 0, drawW, drawH);
          if (hasTransparency(ctx, drawW, drawH)) {
            return canvas.toDataURL('image/png');
          }
          return canvas.toDataURL('image/jpeg', jpegQuality);
        }
        // full mode: natural dimensions, PNG
        canvas.width = el.naturalWidth;
        canvas.height = el.naturalHeight;
        const ctx = canvas.getContext('2d');
        ctx.drawImage(el, 0, 0);
        return canvas.toDataURL('image/png');
      }
      if (el.tagName === 'CANVAS') {
        return el.toDataURL('image/png');
      }
    } catch (e) {
      // CORS or tainted canvas
    }
    return '';
  }

  function extractTextRuns(el) {
    const runs = [];
    for (const child of el.childNodes) {
      if (child.nodeType !== Node.TEXT_NODE) continue;
      const text = child.textContent;
      if (!text || !text.trim()) continue;

      const range = document.createRange();
      range.selectNodeContents(child);
      const rects = range.getClientRects();

      if (rects.length <= 1) {
        // Single line - use full text
        for (const r of rects) {
          if (r.width === 0 || r.height === 0) continue;
          runs.push({
            text: text.trim(),
            bounds: {
              x: r.x + window.scrollX,
              y: r.y + window.scrollY,
              width: r.width,
              height: r.height,
            },
          });
        }
      } else {
        // Multi-line: split text across rects using character-level ranging.
        // Build an array of {charIndex, rectIndex} by binary-searching line breaks.
        const lineTexts = [];
        const lineRects = [];
        for (const r of rects) {
          if (r.width === 0 || r.height === 0) continue;
          lineRects.push(r);
          lineTexts.push('');
        }
        if (lineRects.length === 0) continue;

        // For each character, determine which rect it belongs to by its Y position
        const len = child.textContent.length;
        let currentLine = 0;
        for (let i = 0; i < len; i++) {
          const charRange = document.createRange();
          charRange.setStart(child, i);
          charRange.setEnd(child, Math.min(i + 1, len));
          const charRect = charRange.getBoundingClientRect();
          // Find which line rect this char belongs to (by Y overlap)
          while (currentLine < lineRects.length - 1 &&
                 charRect.top >= lineRects[currentLine].bottom - 1) {
            currentLine++;
          }
          lineTexts[currentLine] += child.textContent[i];
        }

        for (let i = 0; i < lineRects.length; i++) {
          const trimmed = lineTexts[i].trim();
          if (!trimmed) continue;
          const r = lineRects[i];
          runs.push({
            text: trimmed,
            bounds: {
              x: r.x + window.scrollX,
              y: r.y + window.scrollY,
              width: r.width,
              height: r.height,
            },
          });
        }
      }
    }
    return runs;
  }

  function walkElement(el) {
    if (!(el instanceof Element)) return null;
    if (SKIP_TAGS.has(el.tagName)) return null;

    const style = window.getComputedStyle(el);
    if (style.display === 'none' || style.visibility === 'hidden') return null;

    const opacity = parseFloat(style.opacity) || 0;
    if (opacity === 0 && el !== document.body) return null;

    const rect = el.getBoundingClientRect();
    const bounds = {
      x: rect.x + window.scrollX,
      y: rect.y + window.scrollY,
      width: rect.width,
      height: rect.height,
    };

    if (bounds.width === 0 && bounds.height === 0 && el !== document.body) return null;

    const node = {
      bounds,
      backgroundColor: parseColor(style.backgroundColor),
      borders: [
        parseBorder(style, 'top'),
        parseBorder(style, 'right'),
        parseBorder(style, 'bottom'),
        parseBorder(style, 'left'),
      ],
      borderRadius: parseBorderRadius(style),
      opacity,
      overflow: style.overflow,
      boxShadows: parseBoxShadows(style.boxShadow),
      fontFamily: style.fontFamily,
      fontSize: parseFloat(style.fontSize) || 0,
      fontWeight: style.fontWeight,
      fontStyle: style.fontStyle,
      color: parseColor(style.color),
      textDecoration: style.textDecorationLine || style.textDecoration || 'none',
      lineHeight: parseFloat(style.lineHeight) || 0,
      textAlign: style.textAlign,
      tag: el.tagName.toLowerCase(),
      id: el.id || '',
      classes: el.className && typeof el.className === 'string' ? el.className : '',
      role: el.getAttribute('role') || '',
      ariaLabel: el.getAttribute('aria-label') || '',
      href: el.tagName === 'A' ? (el.href || '') : '',
      imageDataURL: captureImage(el),
      textRuns: extractTextRuns(el),
      children: [],
    };

    for (const child of el.children) {
      const childNode = walkElement(child);
      if (childNode) {
        node.children.push(childNode);
      }
    }

    return node;
  }

  const tree = walkElement(document.body);
  return JSON.stringify({
    viewport: {
      width: Math.max(document.documentElement.scrollWidth, window.innerWidth),
      height: Math.max(document.documentElement.scrollHeight, window.innerHeight),
    },
    root: tree,
  });
}
