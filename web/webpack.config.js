// @ts-check
const path = require("path");
const CopyPlugin = require("copy-webpack-plugin");

/** @type {import('webpack').Configuration} */
module.exports = {
  entry: "./src/index.ts",
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: "ts-loader",
        exclude: /node_modules/,
      },
    ],
  },
  resolve: {
    extensions: [".tsx", ".ts", ".js"],
  },
  output: {
    filename: "bundle.js",
    path: path.resolve(__dirname, "public"),
    clean: false,
  },
  plugins: [
    // Copy xterm.css from the package so index.html can link it.
    new CopyPlugin({
      patterns: [
        {
          from: path.resolve(__dirname, "node_modules/@xterm/xterm/css/xterm.css"),
          to: path.resolve(__dirname, "public/xterm.css"),
        },
      ],
    }),
  ],
};
