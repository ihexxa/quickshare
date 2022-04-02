const merge = require("webpack-merge");
const HtmlWebpackPlugin = require("html-webpack-plugin");

const prod = require("./webpack.prod.js");

module.exports = merge(prod, {
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
      template: `${__dirname}/build/template/index.template.html`,
      hash: true,
      filename: `../index.html`,
      minify: false,
    }),
  ],
});
