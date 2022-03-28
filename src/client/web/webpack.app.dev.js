const merge = require("webpack-merge");
const HtmlWebpackPlugin = require("html-webpack-plugin");

const dev = require("./webpack.dev.js");

module.exports = merge(dev, {
  entry: ["./src/app.tsx", "./src/components/api.ts"],
  context: `${__dirname}`,
  output: {
    globalObject: "this",
    path: `${__dirname}/../../../public/static`,
    chunkFilename: "[name].bundle.js",
    filename: "[name].bundle.js",
    library: "Q",
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: `${__dirname}/build/template/index.template.dev.html`,
      hash: true,
      filename: `../index.html`,
    }),
  ],
});
