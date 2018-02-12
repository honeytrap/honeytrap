const webpack = require('webpack');

const { gitDescribeSync } = require('git-describe');
const gitInfo = gitDescribeSync();

const merge = require('webpack-merge');
const common = require('./webpack.common.js');

module.exports = merge(common, {
      devtool: 'source-map',
    plugins: [
        new webpack.DefinePlugin({
            "process.env": {
                NODE_ENV: JSON.stringify(process.env.NODE_ENV || "development"),
                WEBSOCKET_URI: process.env.WEBSOCKET_URI ? JSON.stringify(process.env.WEBSOCKET_URI) : null,
                CLIENT_VERSION: JSON.stringify(gitInfo.raw)
            }
        }),
      new webpack.HotModuleReplacementPlugin(),
    ]
});
