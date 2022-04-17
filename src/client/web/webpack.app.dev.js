const merge = require("webpack-merge");
const HtmlWebpackPlugin = require("html-webpack-plugin");

const dev = require("./webpack.dev.js");

module.exports = merge(dev, {
  plugins: [
    new HtmlWebpackPlugin({
      template: `${__dirname}/build/template/index.template.dev.html`,
      hash: true,
      filename: `../index.html`,
    }),
  ],
});
