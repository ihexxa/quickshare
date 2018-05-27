const common = require("./webpack.config.common.js");
const merge = require("webpack-merge");
const UglifyJS = require("uglifyjs-webpack-plugin");
const webpack = require("webpack");

module.exports = merge(common, {
  devtool: "source-map",
  plugins: [
    new UglifyJS({
      sourceMap: true
    }),
    new webpack.DefinePlugin({
      "process.env.NODE_ENV": JSON.stringify("production")
    })
  ]
});
