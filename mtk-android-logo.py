#!/usr/bin/python3
# -*- coding: utf-8 -*-
"""
mtk-android-logo -- Unpack/repack the logo.bin of a MediaTex Android ROM.

The logo.bin contains several pictures used in early boot stage.
Generally there are some fullscreen logos and several small
pictures used to display battery charge status before the phone
is fully bootstrapped.

  ./mtk-android-logo unpack

Run the script into the same directory containing the "logo.bin"
file; a "logo.d" directory will be created with compressed
(.z) and uncompressed (.bin) files.

  ./mtk-android-logo repack

For repacking, the script needs the original "logo.bin" and a
"logo.d" directory containing the pictures. If the z-compressed
picture exists, it will be used, otherwise the .bin one is
used after z-compressing it.

The resulting image is saved as "logo.repack.bin".

The overall structure of a logo.bin is expected as this:

  * MTK Header, 512 bytes
  * Pictures count: 4 bytes (little-endian 32bit integer)
  * Total Block Size: 4 bytes (little-endian 32bit integer)
  * Offsets Map: 4 bytes x Pictures count
  * Pictures Data Block

Pictures are expected to be raw RGB, stored with z-compression.
The pixel size (width x height) of each picture must be guessed.

Suppose that the screen size is 720x1280 and the biggest,
uncompressed picture is 3686400 bytes:

  3686400 / (720 * 1280) = 4

so we can desume that the picture is 4 bytes/pixel, i.e. a 32bit
RGB. By trial and error we found that it is an BGRA (Blue, Green,
Red and Alpha), we can convert it into PNG with ffmpeg:

  ffmpeg -vcodec rawvideo -f rawvideo \
      -pix_fmt bgra -s 720x1280 -i "img-01.bin" \
      -f image2 -vcodec png "img-01.png"

Once edited, we can convert it back to raw BGRA with:

  ffmpeg -vcodec png -i "img-01.png" \
      -vcodec rawvideo -f rawvideo -pix_fmt bgra "img-01.bin"

It is known that also 16bit raw formats are used (eg. bgr565le),
you can guess it, if the size of the picture is 2 bytes per pixel.
You can list all the pixel formats supprted by ffmpeg by running:

  ffmpeg -pix_fmts

"""

import os
import sys
import zlib

__author__ = "Niccolo Rigacci"
__copyright__ = "Copyright 2019 Niccolo Rigacci <niccolo@rigacci.org>"
__license__ = "GPLv3-or-later"
__email__ = "niccolo@rigacci.org"
__version__ = "0.1.0"

LOGO_BIN = 'logo.img'
LOGO_REPACK = 'logo.repack.bin'
UNPACK_DIR = './logo.d'
SIZE_MTK_HEADER = 512
SIZE_INT = 4
PAD_TO = 32


#--------------------------------------------------------------------------
#--------------------------------------------------------------------------
def unpack_logo():
    with open(LOGO_BIN, 'rb') as logo_bin:
        print('')
        print('Reading MTK header (%d bytes)...' % (SIZE_MTK_HEADER,))
        mtk_header = logo_bin.read(SIZE_MTK_HEADER)
        if mtk_header[8:12].decode('ascii').upper() == 'LOGO':
            print('Found "LOGO" signature at offset 0x08')
        else:
            print('No "LOGO" signature found at offset 0x08, continue anyway...')
        print('Reading pictures count (%d bytes)...' % (SIZE_INT,))
        picture_count = int.from_bytes(logo_bin.read(SIZE_INT), byteorder='little')
        print('File contains %d pictures!' % (picture_count,))
        print('Reading block size (%d bytes)...' % (SIZE_INT,))
        bloc_size = int.from_bytes(logo_bin.read(SIZE_INT), byteorder='little')
        print('Total block size (8 bytes + map + pictures): %d' % (bloc_size,))
        offset_map_size = picture_count * SIZE_INT
        print('Reading offsets map (%d * %d = %d bytes)...' % (picture_count, SIZE_INT, offset_map_size,))
        offsets = {}
        sizes = {}
        for i in range(0, picture_count):
            offsets[i] = int.from_bytes(logo_bin.read(SIZE_INT), byteorder='little')
        for i in range(0, picture_count - 1):
            sizes[i] = offsets[i+1] - offsets[i]
        i += 1
        sizes[i] = bloc_size - offsets[i];
        print('')
        print('    img   |   offset   | size (bytes)')
        print('-'*37)
        images_size = 0
        for i in range(0, picture_count):
            print('     %02d   | 0x%08X | %10d ' % (i + 1, offsets[i], sizes[i]))
            image_z = logo_bin.read(sizes[i])
            images_size += sizes[i]
            image_bin = zlib.decompress(image_z)
            filname = os.path.join(UNPACK_DIR, 'img-%02d.bin' % (i + 1,))
            with open(filname, 'wb') as image:
                image.write(image_bin)
            filname = os.path.join(UNPACK_DIR, 'img-%02d.z' % (i + 1,))
            with open(filname, 'wb') as image:
                image.write(image_z)
        print('-'*37)
        print('%35d' % (images_size,))
        print('')


#--------------------------------------------------------------------------
#--------------------------------------------------------------------------
def repack_logo():

    images = []
    offsets = {}
    sizes = {}
    for file in os.listdir(UNPACK_DIR):
        if file.startswith('img-') and file.endswith('.bin'):
            images.append(file)
    images.sort()

    with open(LOGO_BIN, 'rb') as logo_bin:
        print('')
        print('Reading MTK header from file "%s" (%d bytes)' % (LOGO_BIN, SIZE_MTK_HEADER))
        mtk_header = logo_bin.read(SIZE_MTK_HEADER)
        if mtk_header[8:12].decode('ascii').upper() == 'LOGO':
            print('Found "LOGO" signature at offset 0x08')
        else:
            print('No "LOGO" signature found at offset 0x08, continue anyway...')

    bytes_written = 0
    with open(LOGO_REPACK, 'wb') as new_logo:
        picture_count = len(images)
        bloc_size = (SIZE_INT * 2) + (SIZE_INT * picture_count)
        print('Writing MTK header to file "%s" (%d bytes)...' % (LOGO_REPACK, SIZE_MTK_HEADER,))
        new_logo.write(mtk_header)
        bytes_written += len(mtk_header)
        print('Writing pictures count, which is %d (%d bytes)...' % (picture_count, SIZE_INT))
        new_logo.write(picture_count.to_bytes(SIZE_INT, byteorder='little'))
        bytes_written += SIZE_INT
        # Read and compress images.
        i = 0
        for img in images:
            filename_bin = os.path.join(UNPACK_DIR, img)
            filename_z = filename_bin[:-3] + 'z'
            with open(filename_bin, 'rb') as f:
                image_bin = f.read()
            if (os.path.exists(filename_z)):
                with open(filename_z, 'rb') as f:
                    image_z = f.read()
            else:
                image_z = zlib.compress(image_bin)
                with open(filename_z, 'wb') as f:
                    f.write(image_z)
            sizes[i] = len(image_z)
            if i == 0:
                offsets[i] = bloc_size
            else:
                offsets[i] = offsets[i - 1] + sizes[i - 1]
            bloc_size += sizes[i]
            i += 1
        print('Total block size (8 bytes + map + pictures): %d' % (bloc_size,))
        print('Writing total block size (%d bytes)...' % (SIZE_INT,))
        new_logo.write(bloc_size.to_bytes(SIZE_INT, byteorder='little'))
        bytes_written += SIZE_INT
        # Write offsets map.
        print('Writing offsets map (%d * %d = %d bytes)' % (picture_count, SIZE_INT, picture_count * SIZE_INT))
        print('')
        print('    img   |   offset   | size (bytes)')
        print('-'*37)
        images_size = 0
        for i in range(0, picture_count):
            print('     %02d   | 0x%08X | %10d ' % (i + 1, offsets[i], sizes[i]))
            new_logo.write(offsets[i].to_bytes(SIZE_INT, byteorder='little'))
            bytes_written += SIZE_INT
            images_size += sizes[i]
        print('-'*37)
        print('%35d' % (images_size,))
        print('')

        # Write compressed images.
        for img in images:
            filename_bin = os.path.join(UNPACK_DIR, img)
            filename_z = filename_bin[:-3] + 'z'
            with open(filename_z, 'rb') as f:
                image_z = f.read()
            new_logo.write(image_z)
            bytes_written += len(image_z)

        # Pad to 16 bytes?
        pad_len = PAD_TO - (bytes_written % PAD_TO)
        print('Writing %d bytes to pad to %d' % (PAD_TO, pad_len))
        new_logo.write(b'\0'*pad_len)
        bytes_written += pad_len


#--------------------------------------------------------------------------
#--------------------------------------------------------------------------
if len(sys.argv) < 2:
    print('Usage: %s [unpack|repack]' % (os.path.basename(sys.argv[0])))
    sys.exit(1)

if sys.argv[1] == 'unpack':
    if not os.path.isfile(LOGO_BIN):
        print('ERROR: File "%s" not found, cannot unpack it.' % (LOGO_BIN,))
        sys.exit(1)
    if os.path.exists(UNPACK_DIR):
        print('ERROR: Directory "%s" already exists. Remove it before unpacking "%s".' % (UNPACK_DIR, LOGO_BIN))
        sys.exit(1)
    os.makedirs(UNPACK_DIR)
    unpack_logo()

elif sys.argv[1] == 'repack':
    if not os.path.exists(UNPACK_DIR):
        print('ERROR: Directory "%s" does not exists. Unpack your "%s" before.' % (UNPACK_DIR, LOGO_BIN))
        sys.exit(1)
    repack_logo()
