// const webpack = require("webpack");
// const CleanWebpackPlugin = require("clean-webpack-plugin");
// const UglifyJsPlugin = require("uglifyjs-webpack-plugin");
const path = require("path");
const TerserPlugin = require("terser-webpack-plugin");
const BundleAnalyzerPlugin = require("webpack-bundle-analyzer")
  .BundleAnalyzerPlugin;

module.exports = {
  module: {
    rules: [
      {
        test: /\.ts|tsx$/,
        loader: "ts-loader",
        include: [path.resolve(__dirname, "src")],
        exclude: [
          /node_modules/,
          /\.test\.(ts|tsx)$/
        ]
      },
      {
        test: /\.css$/,
        use: [
          "style-loader",
          {
            loader: "css-loader",
            options: {
              url: false
            }
          }
        ]
      }
    ]
  },
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".json"]
  },
  plugins: [
    // new BundleAnalyzerPlugin()
  ],
  externals: {
    react: "React",
    "react-dom": "ReactDOM",
    immutable: "Immutable"
  },
  optimization: {
    minimizer: [new TerserPlugin()],
    splitChunks: {
      chunks: "all",
      automaticNameDelimiter: ".",
      cacheGroups: {
        default: {
          name: "main",
          filename: "[name].bundle.js"
        },
        commons: {
          name: "vendors",
          test: /[\\/]node_modules[\\/]/,
          chunks: "all",
          minChunks: 2,
          reuseExistingChunk: true
        }
        // worker: {
        //   name: "worker",
        //   test: /[\\/]worker[\\/]/,
        //   filename: "[name].bundle.js"
        // }
      }
    }
  }
};
