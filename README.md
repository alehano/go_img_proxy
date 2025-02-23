# Go Image Proxy

## Description

This project is a Go-based image proxy server that allows you to manipulate images on-the-fly. You can resize, watermark, and apply various transformations to images by specifying options in the URL.

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   ```
2. Navigate to the project directory:
   ```bash
   cd go_img_proxy
   ```
3. Build the project:
   ```bash
   go build
   ```

## Usage

Run the server:

```bash
./go_img_proxy
```

Access images with transformations:

- Without options:
  ```
  http://localhost:8080/-/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- With options:
  ```
  http://localhost:8080/w-100_format-png/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
  ```
- With disabled watermark:
  ```
  http://localhost:8080/w-100_nw-1/upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/220px-Francesco_Melzi_-_Portrait_of_Leonardo.png
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
