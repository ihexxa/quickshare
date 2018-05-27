const merge = require("webpack-merge");
const common = require("./webpack.config.common.js");

module.exports = merge(common, {
  entry: {
    api_test: "./libs/test/api_test"
  },
  devtool: "inline-source-map",
  devServer: {
    contentBase: "./dist"
  },
  watchOptions: {
    aggregateTimeout: 1000,
    poll: 1000,
    ignored: /node_modules/
  }
});
