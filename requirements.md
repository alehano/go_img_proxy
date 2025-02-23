Image proxy writtern in Go.

It has /img endpoit that accept url param with img url and return the image converted

Other params:

- url - url of the image to proxy
- w - width of the image
- h - height of the image
- fit - fit the image to the width and height
- quality - quality of the image
- format - format of the image
- aspect - aspect ratio of the image

if w, h are empty, it will return the image with original size

It supports output: png, jpg formats
Input can be: png, jpg, webp

Use this lib for image processing:
https://pkg.go.dev/github.com/sunshineplan/imgconv@v1.1.13#section-readme

App uses config for watermark (path, position etc)

Using for routing:
https://github.com/go-chi/chi
