#!/usr/bin/env python3
"""Generate a minimal valid PNG file saved as test-image.jpg for Makefile tests."""
import zlib, struct, sys

def png_chunk(chunk_type, data):
    c = chunk_type + data
    return struct.pack('>I', len(data)) + c + struct.pack('>I', zlib.crc32(c) & 0xffffffff)

w, h = 10, 10
sig = b'\x89PNG\r\n\x1a\n'
ihdr = png_chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 2, 0, 0, 0))
raw_row = b'\x00' + b'\xff\x00\x00' * w  # filter byte + 10 red RGB pixels
idat = png_chunk(b'IDAT', zlib.compress(raw_row * h))
iend = png_chunk(b'IEND', b'')

out = sys.argv[1] if len(sys.argv) > 1 else 'test-image.jpg'
with open(out, 'wb') as f:
    f.write(sig + ihdr + idat + iend)
print(f'{out} created')
