const merge = require("webpack-merge");
const HtmlWebpackPlugin = require("html-webpack-plugin");

const prod = require("./webpack.prod.js");

module.exports = merge(prod, {
  plugins: [
    new HtmlWebpackPlugin({
      template: `${__dirname}/build/template/index.template.html`,
      hash: true,
      filename: `../index.html`,
      minify: false,
    }),
  ],
});
