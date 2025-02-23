# Go Image Proxy

## Description

This project is a Go-based image proxy server that allows you to manipulate images on-the-fly. You can resize, watermark, and apply various transformations to images by specifying options in the URL.

Supported input image formats:

- JPEG
- PNG
- GIF
- WebP (not animated)

Supported output image formats:

- JPEG
- PNG

## Usage

URL format:

```
http://localhost:8080/{options}/url/{image_url}
http://localhost:8080/{options}/urlb/{base64_encoded_image_url}
```

The `/urlb/{base64_encoded_image_url}` pattern allows you to provide a base64-encoded URL. Any suffix after a dot (e.g., `.jpg`, `.png`) in the base64-encoded URL will be trimmed before processing.

Options are specified in the format `{key}-{value}` and can be combined using underscores (`_`). For example, to set the width to 100 and the height to 200, you would use `w-100_h-200`.

Options:

- `w`: Width
- `h`: Height
- `f`: Format `jpg` or `png` (default is `jpg`)
- `nw`: No watermark

Examples:

- Without options:
  ```
  http://localhost:8080/-/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- With options:
  ```
  http://localhost:8080/w-100_f-png/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- With multiple options:
  ```
  http://localhost:8080/w-100_h-200/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- With disabled watermark:
  ```
  http://localhost:8080/w-100_nw-1/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- Using base64-encoded URL:
  ```
  http://localhost:8080/w-100_f-png/urlb/aHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEvY29tbW9ucy90aHVtYi9jL2NiL0ZyYW5jZXNjb19NZWx6aV8tX1BvcnRyYWl0X29mX0xlb25hcmRvLnBuZy8yMjBweC1GcmFuY2VzY29fTWVsemlfLV9Qb3J0cmFpdF9vZl9MZW9uYXJkby5wbmc=
  ```
  or
  ```
  http://localhost:8080/w-100_f-png/urlb/aHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEvY29tbW9ucy90aHVtYi9jL2NiL0ZyYW5jZXNjb19NZWx6aV8tX1BvcnRyYWl0X29mX0xlb25hcmRvLnBuZy8yMjBweC1GcmFuY2VzY29fTWVsemlfLV9Qb3J0cmFpdF9vZl9MZW9uYXJkby5wbmc=.png
  ```

## Configuration Options

The server can be configured using command-line flags or environment variables:

- `--quality` or `QUALITY`: Quality of the JPEG image. Default is `85`.
- `--port` or `PORT`: Port to run the server on. Default is `8080`.
- `--watermark-img` or `WATERMARK_IMG`: Path to the watermark image. Default is `logo.png`.
- `--opacity` or `OPACITY`: Opacity of the watermark (0-255). Default is `128`.
- `--random` or `RANDOM`: Apply watermark at a random position.
- `--watermark-size-percent` or `WATERMARK_SIZE_PERCENT`: Size of the watermark as a percentage of the original image. Default is `20`.
- `--offset-x-percent` or `OFFSET_X_PERCENT`: X offset as a percentage of the image width. Default is `10`.
- `--offset-y-percent` or `OFFSET_Y_PERCENT`: Y offset as a percentage of the image height. Default is `10`.
- `--position` or `POSITION`: Watermark position (topleft, topright, bottomleft, bottomright, center). Default is `bottomright`.

## Environment Variables

This project uses the following environment variables:

- `PORT`: The port on which the server will run. Default is `8080`.
- `LOG_LEVEL`: The level of logging detail. Options are `DEBUG`, `INFO`, `WARN`, `ERROR`. Default is `INFO`.
- `IMAGE_CACHE_SIZE`: The maximum number of images to cache in memory. Default is `100`.

Make sure to set these variables in your environment before running the server.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License.
