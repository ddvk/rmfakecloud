const config = {
  imagemin: {
    gifsicle: {
      optimizationLevel: 7,
      interlaced: false
    },
    webp: {
      quality: 75
    },
    optipng: {
      optimizationLevel: 7
    },
    mozjpeg: {
      quality: 20
    },
    pngquant: {
      quality: [0.8, 0.9],
      speed: 4
    },
    svgo: {
      plugins: [
        {
          name: 'removeViewBox'
        },
        {
          name: 'removeStyleElement',
          active: true
        }
      ]
    }
  }
}

export default config
